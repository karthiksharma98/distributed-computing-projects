package main

import (
	"errors"
	"fmt"
	"net"
)

type SdfsRequest struct {
	LocalFName  string
	RemoteFName string
	Type        ReqType
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

func (mem *Member) HandlePutRequest(req SdfsRequest, reply *SdfsResponse) error {

	if req.Type != PutReq {
		return errors.New("Error: Invalid request type for Put Request")
	}

	if len(fileMap) == 0 {
		fileMap = make(map[string][]net.IP)
	}

	var response SdfsResponse
	var ipList []net.IP

	counter := 0

	// TODO: make randomly chosen 4 IPs instead of iteration
	for _, v := range mem.membershipList {
		if v.Health == Alive {
			ipList = append(ipList, v.IPaddr)
			counter++
		}
		if counter == 4 {
			break
		}
	}

	if counter == 4 {
		fileMap[req.RemoteFName] = ipList
		response.IPList = ipList
		*reply = response
		return nil
	}

	return errors.New("Error: Could not find 4 alive nodes")
}

func (mem *Member) HandleGetRequest(req SdfsRequest, reply *SdfsResponse) error {

	var response SdfsResponse

	if val, ok := fileMap[req.RemoteFName]; ok {
		fmt.Println("located", val)
		response.IPList = val
		*reply = response
		return nil
	}

	return errors.New("Error: File not found")

}
