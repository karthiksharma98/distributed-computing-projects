package main

import (
	"errors"
	"fmt"
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
	IPList  []net.IP
	fileMap map[string][]string
}

type ReqType int

const (
	PutReq ReqType = iota
	GetReq
	DelReq
	LsReq
	AddReq
)

func (node *SdfsNode) HandlePutRequest(req SdfsRequest, reply *SdfsResponse) error {
	if node.isMaster == false && node.Master == nil {
		return errors.New("Error: Master not initialized")
	}

	if req.Type != PutReq {
		return errors.New("Error: Invalid request type for Put Request")
	}

	ipList := node.pickRandomNodes(int(Configuration.Settings.replicationFactor))

	if ipList == nil {
		return errors.New("Error: Could not find 3 alive nodes")
	}

	var resp SdfsResponse
	resp.IPList = ipList
	*reply = resp

	return nil

}

func (node *SdfsNode) HandleGetRequest(req SdfsRequest, reply *SdfsResponse) error {
	if node.isMaster == false && node.Master == nil {
		return errors.New("Error: Master not initialized")
	}

	if req.Type != GetReq {
		return errors.New("Error: Invalid request type for Get Request")
	}

	var response SdfsResponse

	if val, ok := node.Master.fileMap[req.RemoteFName]; ok && len(val) != 0 {
		response.IPList = val
		*reply = response
		return nil
	}

	return errors.New("Error: File not found")
}

func (node *SdfsNode) DeleteFile(req SdfsRequest, reply *SdfsResponse) error {
	// TODO: delete the local file before returning nil, else return error
	return nil
}

func (node *SdfsNode) sendDeleteCommand(ip net.IP, RemoteFName string) error {
	if node.isMaster == false && node.Master == nil {
		return errors.New("Error: Master not initialized")
	}

	client, err := rpc.DialHTTP("tcp", ip.String()+":"+fmt.Sprint(Configuration.Service.masterPort))
	if err != nil {
		fmt.Println("Delete connection error: ", err)
	}

	var req SdfsRequest
	var res SdfsResponse

	req.RemoteFName = RemoteFName
	req.Type = DelReq

	return client.Call("Member.DeleteFile", req, &res)
}

func (node *SdfsNode) ModifyMasterFileMap(req SdfsRequest, reply *SdfsResponse) error {
	if node.isMaster == false && node.Master == nil {
		return errors.New("Error: Master not initialized")
	}

	// convert string -> ip.net
	// req.LocalFName here is ip address, need the 27 for the method call to work
	stringIp := req.LocalFName + "/27"
	ipToModify, _, _ := net.ParseCIDR(stringIp)

	if req.Type == AddReq {
		ogList := node.Master.fileMap[req.RemoteFName]
		ogList = append(ogList, ipToModify)
		node.Master.fileMap[req.RemoteFName] = ogList

	}

	return nil
}

func (node *SdfsNode) HandleDeleteRequest(req SdfsRequest, reply *SdfsResponse) error {
	if req.Type != DelReq {
		return errors.New("Error: Invalid request type for Delete Request")
	}

	if val, ok := node.Master.fileMap[req.RemoteFName]; ok && len(val) != 0 {
		failedIndices := make([]int, 0)

		for index, ip := range val {
			err := node.sendDeleteCommand(ip, req.RemoteFName)
			if err != nil {
				failedIndices = append(failedIndices, index)
			}
		}

		if len(failedIndices) == 0 {
			delete(node.Master.fileMap, req.RemoteFName)
			return nil

		} else {
			// make list of failed IPs
			failedIps := make([]net.IP, 0)
			for _, i := range failedIndices {
				failedIps = append(failedIps, node.Master.fileMap[req.RemoteFName][i])
			}

			// replace old list with this one
			node.Master.fileMap[req.RemoteFName] = failedIps

			// send list of failed deletes back to process, exit with error
			var res SdfsResponse
			res.IPList = failedIps
			*reply = res
			return errors.New("Failed deleting files")
		}
	}
	return nil
}

func (node *SdfsNode) HandleLsRequest(req SdfsRequest, reply *SdfsResponse) error {
	if req.Type != LsReq {
		return errors.New("Error: Invalid request type for Ls Request")
	}

	var resp SdfsResponse
	mapCopy := make(map[string][]string)
	for fileName, ipList := range node.Master.fileMap {
		listCopy := make([]string, 0)
		for _, ipAddr := range ipList {
			listCopy = append(listCopy, ipAddr.String())
		}
		mapCopy[fileName] = listCopy
	}

	resp.fileMap = mapCopy
	*reply = resp

	return nil
}
