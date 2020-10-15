package main

import (
	"errors"
	"fmt"
	"net"
)

type PutRequest struct {
	localFName  string
	remoteFName string
}

type PutResponse struct {
	ipList []net.IP
}

func (mem *Member) HandlePutRequest(putReq PutRequest, reply *PutResponse) error {

	var response PutResponse
	var ipList []net.IP
	// go through membership list and return 4 IPs

	counter := 0

	fmt.Println("entered put request")

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

	fmt.Println(ipList)

	if counter == 4 {
		response.ipList = ipList
		*reply = response
		return nil
	}

	return errors.New("Error: Could not find 4 alive nodes")
}
