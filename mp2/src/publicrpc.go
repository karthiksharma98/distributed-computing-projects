package main

import (
	"errors"
	"fmt"
	"math/rand"
	"net"
	"net/rpc"
)

type SdfsRequest struct {
	LocalFName  string
	RemoteFName string
	Type        ReqType
}

type UploadAck struct {
	RemoteFname string
	IPaddr      net.IP
}

type SdfsResponse struct {
	IPList []net.IP
}

type ReqType int

const (
	PutReq ReqType = iota
	GetReq
	DelReq
)

// stores file metadata
var (
	fileMap map[string][]net.IP
)

func (mem *Member) pickRandomNodes(minReplicas int) []net.IP {

	// TODO: should master store files? return minReplicas based on that (rn we return 3 replicas)

	i := 0
	iplist := make([]net.IP, 0)

	// first get all alive IP Addresses in list
	for k := range mem.membershipList {
		if mem.membershipList[k].Health == Alive {
			iplist = append(iplist, mem.membershipList[k].IPaddr)
		}
		i++
	}

	if len(iplist) < minReplicas {
		return nil
	}

	// shuffle and choose first few
	rand.Shuffle(len(iplist), func(i, j int) { iplist[i], iplist[j] = iplist[j], iplist[i] })
	return iplist[:minReplicas+1]
}

func (mem *Member) HandlePutRequest(req SdfsRequest, reply *SdfsResponse) error {

	if req.Type != PutReq {
		return errors.New("Error: Invalid request type for Put Request")
	}

	if len(fileMap) == 0 {
		fileMap = make(map[string][]net.IP)
	}

	var response SdfsResponse
	ipList := mem.pickRandomNodes(3)

	if ipList != nil {
		fileMap[req.RemoteFName] = ipList
		response.IPList = ipList
		*reply = response
		return nil
	}

	return errors.New("Error: Could not find 4 alive nodes")
}

func (mem *Member) AddIPToFileMap(ack UploadAck, reply *SdfsResponse) error {
	fileMap[ack.RemoteFname] = append(fileMap[ack.RemoteFname], ack.IPaddr)
	return nil
}

func (mem *Member) HandleGetRequest(req SdfsRequest, reply *SdfsResponse) error {

	if req.Type != GetReq {
		return errors.New("Error: Invalid request type for Get Request")
	}

	var response SdfsResponse

	if val, ok := fileMap[req.RemoteFName]; ok && len(val) != 0 {
		response.IPList = val
		*reply = response
		return nil
	}

	return errors.New("Error: File not found")
}

func (mem *Member) DeleteFile(req SdfsRequest, reply *SdfsResponse) error {
	// TODO: delete the local file before returning nil, else return error
	return nil
}

func sendDeleteCommand(ip net.IP, RemoteFName string) error {
	client, err := rpc.DialHTTP("tcp", ip.String()+":"+fmt.Sprint(Configuration.Service.rpcReqPort))
	if err != nil {
		fmt.Println("Delete connection error: ", err)
	}

	var req SdfsRequest
	var res SdfsResponse

	req.RemoteFName = RemoteFName
	req.Type = DelReq

	return client.Call("Member.DeleteFile", req, &res)
}

func (mem *Member) HandleDeleteRequest(req SdfsRequest, reply *SdfsResponse) error {
	if req.Type != DelReq {
		return errors.New("Error: Invalid request type for Delete Request")
	}

	if val, ok := fileMap[req.RemoteFName]; ok && len(val) != 0 {
		failedIndices := make([]int, 0)

		for index, ip := range val {
			err := sendDeleteCommand(ip, req.RemoteFName)
			if err != nil {
				failedIndices = append(failedIndices, index)
			}
		}

		if len(failedIndices) == 0 {
			delete(fileMap, req.RemoteFName)
			return nil

		} else {
			// make list of failed IPs
			failedIps := make([]net.IP, 0)
			for _, i := range failedIndices {
				failedIps = append(failedIps, fileMap[req.RemoteFName][i])
			}

			// replace old list with this one
			fileMap[req.RemoteFName] = failedIps

			// send list of failed deletes back to process, exit with error
			var res SdfsResponse
			res.IPList = failedIps
			*reply = res
			return errors.New("Failed deleting files")
		}
	}
	return nil
}
