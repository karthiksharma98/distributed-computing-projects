package main

import (
	"net"
)

type PutRequest struct {
	localFName  string
	remoteFName string
}

type PutResponse struct {
	ip net.IP
}
