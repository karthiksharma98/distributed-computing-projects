package main

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"net/rpc"
	"os"
	"time"
)

type SdfsNode struct {
	*Member
	NumFiles  int
	DiskSpace int
	FileList  []string

	// Master metadata
	MasterId uint8
	isMaster bool
	Master   *SdfsMaster
}

type SdfsMaster struct {
	fileMap map[string][]net.IP
	lockMap map[string]*SdfsMutex
	sessMap map[int32](chan bool)
}

type connectionError struct {
	ip net.IP
}

func (e connectionError) Error() string {
	return e.ip.String()
}

// stores file metadata

var (
	electionFlag = false
	okAck        chan bool
	sdfsListener *net.UDPConn
)

func NewSdfsNode(mem *Member, setMaster bool) *SdfsNode {
	node := &SdfsNode{
		mem,
		0,
		0,
		make([]string, 0),
		0,
		setMaster,
		nil,
	}

	if setMaster {
		node.Master = NewSdfsMaster()
	}
	return node
}

func NewSdfsMaster() *SdfsMaster {
	master := &SdfsMaster{
		make(map[string][]net.IP),
		make(map[string]*SdfsMutex),
		make(map[int32](chan bool)),
	}
	return master
}

// Listen for failed nodes
func (node *SdfsNode) MemberListen() {
	for {
		select {
		case id := <-failCh:
			// If detected master failed, call for Election
			if id == node.MasterId {
				node.Election()
			}
			if node.isMaster {
				// handle replication if you're the master
				err := node.handleReplicationOnFailure(id)
				if err != nil {
					fmt.Println(err)
				}
			}
			continue
		}
	}
}

// Initiate election
func (node *SdfsNode) Election() {
	// Check if election already started
	if electionFlag {
		return
	}
	electionFlag = true
	Info.Println("Starting election. Election in progress.")

	okAck = make(chan bool)
	// Send ElectionMsg to nodes with higher IDs than itself
	for _, mem := range node.Member.membershipList {
		if mem.MemberID > node.Member.memberID {
			Send(mem.IPaddr.String()+":"+fmt.Sprint(Configuration.Service.masterPort), ElectionMsg, []byte{node.Member.memberID})
		}
	}
	// Wait for timeout and send CoordinatorMsg to all nodes if determined that it has the highest ID
	select {
	case <-okAck:
		return
	case <-time.After(2 * time.Second):
		Info.Println("Coordinator is self.")
		node.handleCoordinator(node.Member.memberID)
		// Send CoordinatorMsg to nodes lower than itself
		for _, mem := range node.Member.membershipList {
			if mem.MemberID < node.Member.memberID {
				Info.Println("Sending to", mem.IPaddr.String())
				Send(mem.IPaddr.String()+":"+fmt.Sprint(Configuration.Service.masterPort), CoordinatorMsg, []byte{node.Member.memberID})
			}
		}
	}
}

// handle election message
func (node *SdfsNode) handleElection(senderAddr net.IP, id uint8) {
	if id < node.Member.memberID {
		Send(senderAddr.String()+":"+fmt.Sprint(Configuration.Service.masterPort), OkMsg, []byte{node.Member.memberID})

		// Start election again
		go node.Election()
	}
}

// Set new coordinator/master
func (node *SdfsNode) handleCoordinator(id uint8) {
	if !electionFlag {
		return
	}
	// Update masterId and rpc's connection
	Info.Println("Elected", id, ". Election complete.")
	electionFlag = false
	node.MasterId = id

	// If self is elected, initialize new SdfsMaster object
	if id == node.Member.memberID {
		node.isMaster = true
		node.Master = NewSdfsMaster()
	}

	// Encode file list
	b := new(bytes.Buffer)
	e := gob.NewEncoder(b)
	err := e.Encode(node.FileList)
	if err != nil {
		panic(err)
	}

	// Send fileList, numFiles to new coordinator/master
	Send(node.Member.membershipList[id].IPaddr.String()+":"+fmt.Sprint(Configuration.Service.masterPort), RecoverMasterMsg, b.Bytes())

	// Redirect RPC connection to new IP when Master ready
	node.closeRPCClient()
	node.startRPCClient(node.Member.membershipList[id].IPaddr.String(), fmt.Sprint(Configuration.Service.masterPort))
}

// Handle election ok message
func (node *SdfsNode) handleOk() {
	select {
	case okAck <- true:
		Info.Println("Ok")
	default:
		Info.Println("Ok returned")
	}
}

// Master recovery
func (node *SdfsNode) handleRecoverMaster(senderAddr net.IP, fileListBytes []byte) {
	if !node.isMaster || node.Master == nil {
		return
	}

	var newFileList []string
	// Decode member metadata
	b := bytes.NewBuffer(fileListBytes)
	d := gob.NewDecoder(b)
	err := d.Decode(&newFileList)

	if err != nil {
		panic(err)
	}
	// Read incoming filelist and set sender IP as map value
	for _, fname := range newFileList {
		if val, ok := node.Master.fileMap[fname]; ok {
			ipList := append(val, senderAddr)
			node.Master.AddIPToFileMap(fname, ipList)
		} else {
			node.Master.fileMap[fname] = []net.IP{senderAddr}
		}
	}
}

// Delete file local
func (node *SdfsNode) deleteFile(localFilename string) bool {
	// delete file given file name
	err := os.Remove(localFilename)
	if err != nil {
		return false
	}
	return true
}

// Cleanup files, file lists, etc
func (node *SdfsNode) cleanupLocal() {
	for _, fname := range node.FileList {
		node.deleteFile(fname)
	}
}

// Remove ip from iplist
func (node *SdfsMaster) deleteIP(fname string, ipAddr string) {
	ipList, ok := node.fileMap[fname]
	if !ok {
		return
	}

	for idx, ip := range ipList {
		if ip.String() != ipAddr {
			continue
		}
		ipList[idx] = ipList[len(ipList)-1]
		node.fileMap[fname] = ipList[:len(ipList)-1]
	}
}

// Cleanup node
func (node *SdfsNode) cleanupNode(id uint8) {
	if node.Master != nil {
		return
	}
	// get ip of node to cleanup
	ipAddr := node.Member.membershipList[id].IPaddr.String()
	// read fileMap, remove ip from list if matches node
	for fname, _ := range node.Master.fileMap {
		node.Master.deleteIP(fname, ipAddr)
	}
}

// List set of file names replicated on process
func (node *SdfsNode) Store() error {
	file, err := os.Open("./SDFS/")
	if err != nil {
		return err
	}
	names, err := file.Readdirnames(0)
	if err != nil {
		return err
	}
	fmt.Println(names)
	return nil
}

// Add IPList to file map
func (node *SdfsMaster) AddIPToFileMap(fname string, ipList []net.IP) {
	if ipList != nil {
		node.fileMap[fname] = ipList
	}
}

// Chooses random set of nodes to replicate
func (node *SdfsNode) pickRandomNodes(minReplicas int) []net.IP {

	// TODO: should master store files? return minReplicas based on that (rn we return 3 replicas)

	i := 0
	iplist := make([]net.IP, 0)

	// first get all alive IP Addresses in list
	for _, mem := range node.Member.membershipList {
		if mem.Health == Alive {
			iplist = append(iplist, mem.IPaddr)
		}
		i++
	}

	if len(iplist) < minReplicas {
		return nil
	}

	// shuffle and choose first few
	rand.Shuffle(len(iplist), func(i, j int) { iplist[i], iplist[j] = iplist[j], iplist[i] })
	return iplist[:minReplicas]
}

// asks aliveIP to upload filename to newIP
func sendUploadCommand(aliveIP net.IP, newIP net.IP, filename string) error {
	client, err := rpc.DialHTTP("tcp", aliveIP.String()+":"+fmt.Sprint(Configuration.Service.masterPort))
	if err != nil {
		// fmt.Println("sendUploadCommand error, aliveIP dead: ", aliveIP.String(), err)
		return connectionError{ip: aliveIP}
	}

	var req SdfsRequest
	var res SdfsResponse

	req.LocalFName = filename
	req.RemoteFName = filename
	req.IPAddr = newIP
	req.Type = UploadReq

	err = client.Call("SdfsNode.UploadAndModifyMap", req, &res)
	if err != nil {
		// fmt.Println("sendUploadCommand error, chosenIP dead: ", newIP.String(), err)
		return connectionError{ip: newIP}
	}

	return nil
}

func chooseIP(A []net.IP, B []net.IP) net.IP {
	// chooses an IP from A not in B
	var chosenIP net.IP = nil
	for _, ip := range A {
		if checkMember(ip, B) == -1 {
			chosenIP = ip
			break
		}
	}
	return chosenIP
}

// check if ip is in iplist
func checkMember(ip net.IP, iplist []net.IP) int {
	for i, val := range iplist {
		if ip.Equal(val) {
			return i
		}
	}
	return -1
}

// find an alive IP from membership list that is not in failedIPList and not in existing replicas
func findNewReplicaIP(membershipList map[uint8]membershipListEntry, filename string, failedIPList []net.IP, replicas []net.IP) net.IP {
	for _, listEntry := range membershipList {
		if checkMember(listEntry.IPaddr, failedIPList) == -1 && listEntry.Health == Alive && checkMember(listEntry.IPaddr, replicas) == -1 {
			return listEntry.IPaddr
		}
	}
	return nil
}

func (node *SdfsNode) handleReplicationOnFailure(memberID uint8) error {
	//TODO: sleep for a bit to ensure all failures quiesce before doing this?

	failedIP := node.Member.membershipList[memberID].IPaddr
	failedIPList := []net.IP{failedIP}

	fmt.Println("I am handling failure")

	// iterate over fileMap and find files that this member stores
	for filename, ipList := range node.Master.fileMap {
		if failedIndex := checkMember(failedIP, ipList); failedIndex != -1 {
			fmt.Println("failed file ", filename)
			//remove failedIP from fileMap
			node.Master.fileMap[filename] = append(ipList[:failedIndex], ipList[failedIndex+1:]...)

			if len(node.Master.fileMap[fileName]) == int(Configuration.Settings.replicationFactor) {
				// if file already has three alive replicas then don't do anything
				continue
			}

			// find an alive IP that doesn't already contain file
			newIP := findNewReplicaIP(node.Member.membershipList, filename, failedIPList, ipList)
			if newIP == nil {
				return errors.New("No available IP to upload to for " + filename)
			}
			// choose alive IP containing the file that will upload to newIP
			chosenIP := chooseIP(ipList, failedIPList)
			if chosenIP == nil {
				return errors.New("All replicas dead for " + filename)
			}

			// request chosenIP to upload file to newIP and add IP to fileMap
			err := sendUploadCommand(chosenIP, newIP, filename)

		errorLoop:
			for {
				fmt.Println("new, chosen= ", newIP, chosenIP)
				switch err := err.(type) {
				case nil:
					break errorLoop
				case connectionError:
					if err.ip.Equal(newIP) {
						// can't connect to newIP and upload there, try another
						failedIPList = append(failedIPList, newIP)
						fmt.Println("failed IPS:", failedIPList)
						newIP = findNewReplicaIP(node.Member.membershipList, filename, failedIPList, ipList)
						if newIP == nil {
							return errors.New("No available IP to upload to for " + filename)
						}
					} else if err.ip.Equal(chosenIP) {
						// can't connect to existing replica to upload, try another
						failedIPList = append(failedIPList, chosenIP)
						fmt.Println("failed IPS:", failedIPList)
						chosenIP = chooseIP(ipList, failedIPList)
						if chosenIP == nil {
							return errors.New("All replicas dead for " + filename)
						}
					}
				default:
					break errorLoop
				}
				err = sendUploadCommand(chosenIP, newIP, filename)
			}
		}
	}
	return nil
}
