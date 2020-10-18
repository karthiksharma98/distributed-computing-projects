package main

import (
        "fmt"
        "net"
	"math/rand"
        "time"
)

type SdfsNode struct {
        *Member
        NumFiles int
        DiskSpace int
        FileList []string

        // Master metadata
        MasterId uint8
        isMaster bool
        Master *SdfsMaster
}

type SdfsMaster struct {
        fileMap map[string][]net.IP
}
// stores file metadata

var (
        electionFlag = false
        okAck chan bool
        sdfsListener *net.UDPConn
)

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
                // Send CoordinatorMsg to nodes lower than itself
                for _, mem := range node.Member.membershipList {
                        node.handleCoordinator(node.Member.memberID)
                        if mem.MemberID < node.Member.memberID {
                                Info.Println("Sending to", mem.IPaddr.String())
                                Send(mem.IPaddr.String()+":"+fmt.Sprint(Configuration.Service.rpcReqPort), CoordinatorMsg, []byte{node.Member.memberID})
                        }
                }
        }
}

// handle election message
func (node *SdfsNode) handleElection(senderAddr net.IP, id uint8)  {
        if id < node.Member.memberID {
                Send(senderAddr.String()+":"+fmt.Sprint(Configuration.Service.rpcReqPort), OkMsg, []byte{node.Member.memberID})

                // Start election again
                go node.Election()
        }
}

// Set new coordinator/master
func (node *SdfsNode) handleCoordinator(id uint8) {
        Info.Println("Elected", id, ". Election complete.")
        electionFlag = false
        node.MasterId = id
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

func (node *SdfsMaster) AddIPToFileMap(fname string, ipList []net.IP) {
	if ipList != nil {
		node.fileMap[fname] = ipList
        }
}

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
                master := NewSdfsMaster()
                node.Master = master
        }
        return node
}

func NewSdfsMaster() *SdfsMaster {
        master := &SdfsMaster{
                make(map[string][]net.IP),
        }
        return master
}

