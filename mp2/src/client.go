package main

import (
	"net"
)

type PutRequest struct {
	localFName  string
	remoteFName string
}

type PutResponse struct {
	ipList []net.IP
}
