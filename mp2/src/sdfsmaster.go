package main

import (
        "net"
	"math/rand"
)

type SdfsNode struct {
        *Member
        MasterID int
        NumFiles int
        DiskSpace int
        FileList []string
        isMaster bool
}

type SdfsMaster struct {
        *SdfsNode
        fileMap map[string][]net.IP
}
// stores file metadata

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

func (node *SdfsMaster) AddIPToFileMap(fname string, ipList []net.IP) {
	if ipList != nil {
		node.fileMap[fname] = ipList
	}
}

func NewSdfsNode(mem *Member) *SdfsNode {
        node := &SdfsNode{
                mem,
                0,
                0,
                0,
                make([]string, 0),
                false,
        }
        return node
}

func NewSdfsMaster(node *SdfsNode) *SdfsMaster {
        master := &SdfsMaster{
                node,
                make(map[string][]net.IP),
        }
        return master
}

/*
func FindMaxID() int {
        masterID := 0
        for k, v := range mem {
                if masterID < k {
                        masterID = k
                }
        }
        return masterID
}*/
