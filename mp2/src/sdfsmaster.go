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
        okAck = make(chan bool)
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
func (node *SdfsNode) memberListen() {
        for {
                select {
                case id := <-failCh:
                        // if master failed
                        if id == node.MasterId {
                                node.Election()
                        }
                }
        }
}

// Initiate election
func (node *SdfsNode) Election() {
        // Send ElectionMsg to nodes with higher IDs than itself
        for _, mem := range node.Member.membershipList {
                if mem.MemberID > node.Member.memberID {
                        Send(mem.IPaddr.String()+":"+fmt.Sprint(Configuration.Service.rpcReqPort), ElectionMsg, []byte{mem.MemberID})
                }
        }
        // Wait for timeout and send CoordinatorMsg to all nodes if determined that it has the highest ID
        select {
        case <-okAck:
                return
        case <-time.After(2 * time.Second):
                node.handleCoordinator(node.Member.memberID)
                for _, mem := range node.Member.membershipList {
                        Send(mem.IPaddr.String()+":"+fmt.Sprint(Configuration.Service.rpcReqPort), CoordinatorMsg, []byte{node.Member.memberID})
                }
        }
}

// handle election message
func (node *SdfsNode) handleElection(senderAddr string, id uint8)  {
        // Elect again
        if id < node.Member.memberID {
                Send(senderAddr+":"+fmt.Sprint(Configuration.Service.rpcReqPort), OkMsg, []byte{node.Member.memberID})
                node.Election()
        }
}

// Set new coordinator/master
func (node *SdfsNode) handleCoordinator(id uint8) {
        node.MasterId = id
}

// Handle election ok message
func (node *SdfsNode) handleOk() {
        okAck <- true
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

