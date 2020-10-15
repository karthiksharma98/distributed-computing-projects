package main

import (
	"errors"
	"net"
)

type PutRequest struct {
	localFName  string
	remoteFName string
}

type PutResponse struct {
	ipList []net.IP
}

func (mem *Member) PutRequest(putReq PutRequest, reply *PutResponse) error {

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
		response.ipList = ipList
		*reply = response
		return nil
	}

	return errors.New("Error: Could not find 4 alive nodes")
}
