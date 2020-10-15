package main

import (
	"errors"
	"fmt"
	"net"
)

type PutRequest struct {
	LocalFName  string
	RemoteFName string
}

type PutResponse struct {
	IpAddr net.IP
}

func (mem *Member) HandlePutRequest(putReq PutRequest, reply *PutResponse) error {

	var response PutResponse
	// var ipList []net.IP

	var testIP net.IP
	// go through membership list and return 4 IPs

	counter := 0

	fmt.Println("entered put request")

	// TODO: make randomly chosen
	for _, v := range mem.membershipList {
		if v.Health == Alive {
			testIP = v.IPaddr
			counter++
		}
		if counter == 1 {
			break
		}
	}

	fmt.Println(testIP)

	if counter == 1 {
		response.IpAddr = testIP
		*reply = response
		return nil
	}

	return errors.New("Error: Could not find 4 alive nodes")
}
