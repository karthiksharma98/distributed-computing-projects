package main

import (
	"errors"
	"fmt"
	"math"
	"net"
	"net/rpc"
	"os"
	"strconv"
)

type SdfsNode struct {
	*Member
	// Master metadata
	MasterId uint8
	isMaster bool
	Master   *SdfsMaster
}

type SdfsRequest struct {
	LocalFName  string
	RemoteFName string
	IPAddr      net.IP
	Type        ReqType
	BlockID     int
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

func NewSdfsNode(mem *Member, setMaster bool) *SdfsNode {
	node := &SdfsNode{
		mem,
		0,
		setMaster,
		nil,
	}

	if setMaster {
		node.Master = NewSdfsMaster()
	}
	return node
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

// GetLogicalSplits returns indices of byte array containing '\n' closest to block boundaries
func GetLogicalSplits(fileContents []byte) []int {

	var indices []int
	for i := 0; i < len(fileContents); i += int(Configuration.Settings.blockSize) {
		for j := i; j < int(math.Min(float64(i+int(Configuration.Settings.blockSize)), float64(len(fileContents)))); j++ {
			if fileContents[j] == '\n' {
				indices = append(indices, j)
				break
			}
		}
	}
	return indices
}

func (node *SdfsNode) RpcPut(localFname string, remoteFname string) {

	numAlive := process.GetNumAlive()
	numSuccessful := 0

	fileContents := GetFileContents(localFname)
	logicalSplitBoundaries := GetLogicalSplits(fileContents)
	numBlocks := len(logicalSplitBoundaries)

	for blockIdx, _ := range logicalSplitBoundaries {
		var ipsAttempted map[string]bool
		req := SdfsRequest{LocalFName: localFname, RemoteFName: remoteFname, Type: PutReq, BlockID: blockIdx}
		// attempt to get as many replications needed, until you've attempted all the IPs
		for numSuccessful < int(Configuration.Settings.replicationFactor) && len(ipsAttempted) <= numAlive {
			var res SdfsResponse
			var err error
			if len(ipsAttempted) == 0 {
				err = client.Call("SdfsNode.HandlePutRequest", req, &res)
			} else {
				err = client.Call("SdfsNode.GetRandomNodes", req, &res)
			}

			if err != nil {
				fmt.Println("Put failed", err)
				return
			}
			// attempt upload each file
			for _, ipAddr := range res.IPList {
				if _, exists := ipsAttempted[ipAddr.String()]; !exists {
					ipsAttempted[ipAddr.String()] = true

					var blockStart int
					var blockEnd int

					if blockIdx == 0 {
						blockStart = 0
					} else {
						blockStart = logicalSplitBoundaries[blockIdx-1] + 1
					}
					blockEnd = logicalSplitBoundaries[blockIdx] + 1

					err := Upload(ipAddr.String(), fmt.Sprint(Configuration.Service.filePort), req.LocalFName, req.RemoteFName+".blk_"+string(blockIdx), fileContents[blockStart:blockEnd])

					if err != nil {
						fmt.Println("error in upload process.", err)
					} else {
						numSuccessful += 1
						// succesfull upload -> add to master's file map
						mapReq := SdfsRequest{LocalFName: ipAddr.String(), RemoteFName: remoteFname, Type: AddReq, BlockID: blockIdx}
						var mapRes SdfsResponse
						mapErr := client.Call("SdfsNode.AddToFileMap", mapReq, &mapRes)
						if mapErr != nil {
							fmt.Println(mapErr)
						}
					}
				}
			}

			// update alive nodes in case there's not enough anymore
			numAlive = process.GetNumAlive()
		}
	}
}

func (node *SdfsNode) RpcGet(remoteFname string, localFname string) {
	req := SdfsRequest{LocalFName: localFname, RemoteFName: remoteFname, Type: GetReq}
	var res SdfsResponse

	err := client.Call("SdfsNode.HandleGetRequest", req, &res)
	if err != nil {
		fmt.Println(err)
	} else {
		for _, ipAddr := range res.IPList {
			err := Download(ipAddr.String(), fmt.Sprint(Configuration.Service.filePort), req.RemoteFName, req.LocalFName)

			if err != nil {
				fmt.Println("error in download process at ", ipAddr, ": ", err)
			} else {
				// successful download
				return
			}
		}
	}
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

	if val, ok := node.Master.fileMap[req.RemoteFName][req.BlockID]; ok && len(val) != 0 {
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

func (node *SdfsNode) AddToFileMap(req SdfsRequest, reply *SdfsResponse) error {
	if node.isMaster == false && node.Master == nil {
		return errors.New("Error: Master not initialized")
	}

	// convert string -> ip.net
	stringIp := req.LocalFName
	ipToModify := net.ParseIP(stringIp)

	if req.Type == AddReq {
		// Don't add duplicate IP
		if val, ok := node.Master.fileMap[req.RemoteFName][req.BlockID]; ok && checkMember(ipToModify, val) != -1 {
			return nil
		}
		if val, ok := node.Master.fileMap[req.RemoteFName][req.BlockID]; !ok {
			node.Master.fileMap[req.RemoteFName] = make(map[int][]net.IP)
		}
		ogList := node.Master.fileMap[req.RemoteFName][req.BlockID]
		ogList = append(ogList, ipToModify)
		node.Master.fileMap[req.RemoteFName][req.BlockID] = ogList

		if val, ok := node.Master.numBlocks
		node.Master.numBlocks[req.RemoteFName] 
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
		mapErr := client.Call("SdfsNode.AddToFileMap", mapReq, &mapRes)
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
