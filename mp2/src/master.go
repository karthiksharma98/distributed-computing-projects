package main

import (
	"errors"
	"net"
)

type PutRequest struct {
	LocalFName  string
	RemoteFName string
}

type PutResponse struct {
	IPList []net.IP
}

func (mem *Member) HandlePutRequest(putReq PutRequest, reply *PutResponse) error {

	var response PutResponse
	var ipList []net.IP

	// go through membership list and return 4 IPs

	counter := 0

	// TODO: make randomly chosen
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
		response.IPList = ipList
		*reply = response
		return nil
	}

	return errors.New("Error: Could not find 4 alive nodes")
}
