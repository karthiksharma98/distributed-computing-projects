package main

import (
	"net"
)

type PutRequest struct {
	LocalFName  string
	RemoteFName string
}

type PutResponse struct {
	IpAddr net.IP
}
