package main

import (
	"errors"
	"fmt"
	"net"
	"net/rpc"
	"os"
	"strconv"
)

type SdfsRequest struct {
	LocalFName  string
	RemoteFName string
	IPAddr      net.IP
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
	LsReq
	AddReq
	UploadReq
)

func InitSdfs(mem *Member, setMaster bool) *SdfsNode {
	sdfs := NewSdfsNode(mem, setMaster)
	// start SDFS listener
	go sdfs.ListenSdfs(fmt.Sprint(Configuration.Service.masterPort))
	// start RPC Server for handling requests get/put/delete/ls
	sdfs.startRPCServer(fmt.Sprint(Configuration.Service.masterPort))
	sdfs.startRPCClient(Configuration.Service.masterIP, fmt.Sprint(Configuration.Service.masterPort))
	return sdfs
}

func (node *SdfsNode) GetRandomNodes(req SdfsRequest, reply *SdfsResponse) error {
	repFactor := int(Configuration.Settings.replicationFactor)
	ipList := node.pickRandomNodes(repFactor)
	if ipList == nil {
		return errors.New("Error: Could not find " + strconv.Itoa(repFactor) + " alive nodes")
	}

	var resp SdfsResponse
	resp.IPList = ipList
	*reply = resp

	return nil
}

// Rpc wrapper for ls
func (node *SdfsNode) RpcListIPs(fname string) {
	var res SdfsResponse
	req := SdfsRequest{LocalFName: "", RemoteFName: fname, Type: GetReq}

	err := client.Call("SdfsNode.HandleGetRequest", req, &res)
	if err != nil {
		fmt.Println("Failed ls. ", err)
	} else {
		fmt.Print(fname, " =>   ")
		for _, ip := range res.IPList {
			fmt.Print(ip.String(), ", ")
		}
		fmt.Println()
	}
}

// Rpc wrapper for delete
func (node *SdfsNode) RpcDelete(fname string) {
	var res SdfsResponse
	req := SdfsRequest{LocalFName: "", RemoteFName: fname, Type: DelReq}

	err := client.Call("SdfsNode.HandleDeleteRequest", req, &res)

	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("Deleted successfully: ", req.RemoteFName)
}

func (node *SdfsNode) HandlePutRequest(req SdfsRequest, reply *SdfsResponse) error {
	if node.isMaster == false && node.Master == nil {
		return errors.New("Error: Master not initialized")
	}

	if req.Type != PutReq {
		return errors.New("Error: Invalid request type for Put Request")
	}

	if val, ok := node.Master.fileMap[req.RemoteFName]; ok && len(val) != 0 {
		// if file exists already, return those IPs
		ipList := val
		var resp SdfsResponse
		resp.IPList = ipList
		*reply = resp

		return nil
	}

	return node.GetRandomNodes(req, reply)
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
	return os.Remove("./" + dirName + "/" + req.RemoteFName)
}

func (node *SdfsNode) sendDeleteCommand(ip net.IP, RemoteFName string) error {
	if node.isMaster == false && node.Master == nil {
		return errors.New("Error: Master not initialized")
	}

	client, err := rpc.DialHTTP("tcp", ip.String()+":"+fmt.Sprint(Configuration.Service.masterPort))
	if err != nil {
		fmt.Println("Delete connection error: ", err)
		return err
	}

	var req SdfsRequest
	var res SdfsResponse

	req.RemoteFName = RemoteFName
	req.Type = DelReq

	return client.Call("SdfsNode.DeleteFile", req, &res)
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

func (node *SdfsNode) UploadAndModifyMap(req SdfsRequest, reply *SdfsResponse) error {
	fileContents := GetFileContents(req.LocalFName)
	err := Upload(req.IPAddr.String(), fmt.Sprint(Configuration.Service.filePort), req.LocalFName, req.RemoteFName, fileContents)

	if err != nil {
		return err
	} else {
		// succesfull upload -> add to master's file map
		mapReq := SdfsRequest{LocalFName: req.IPAddr.String(), RemoteFName: req.RemoteFName, Type: AddReq}
		var mapRes SdfsResponse
		mapErr := client.Call("SdfsNode.ModifyMasterFileMap", mapReq, &mapRes)
		if mapErr != nil {
			return mapErr
		}
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
