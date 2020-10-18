package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"math/rand"
	"net"
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
			Send(mem.IPaddr.String()+":"+fmt.Sprint(Configuration.Service.rpcReqPort), ElectionMsg, []byte{node.Member.memberID})
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
				Send(mem.IPaddr.String()+":"+fmt.Sprint(Configuration.Service.rpcReqPort), CoordinatorMsg, []byte{node.Member.memberID})
			}
		}
	}
}

// handle election message
func (node *SdfsNode) handleElection(senderAddr net.IP, id uint8) {
	if id < node.Member.memberID {
		Send(senderAddr.String()+":"+fmt.Sprint(Configuration.Service.rpcReqPort), OkMsg, []byte{node.Member.memberID})

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

	node.closeRPCClient()
	node.startRPCClient(node.Member.membershipList[id].IPaddr.String(), fmt.Sprint(Configuration.Service.rpcReqPort))
	// If self is elected, initialize an SdfsMaster object and set
	if id == node.Member.memberID {
		node.isMaster = true
		node.Master = NewSdfsMaster()
		return
	}

	// Encode file list
	b := new(bytes.Buffer)
	e := gob.NewEncoder(b)
	err := e.Encode(node.FileList)
	if err != nil {
		panic(err)
	}

	// Send fileList, numFiles to new coordinator/master
	Send(node.Member.membershipList[id].IPaddr.String()+":"+fmt.Sprint(Configuration.Service.port), RecoverMasterMsg, b.Bytes())
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

func (node *SdfsMaster) AddIPToFileMap(fname string, ipList []net.IP) {
	if ipList != nil {
		node.fileMap[fname] = ipList
	}
}

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
