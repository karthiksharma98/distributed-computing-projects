package main

import (
	"errors"
	"fmt"
	"net"
	"net/rpc"
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

// check if ip is in iplist
func checkMember(ip net.IP, iplist []net.IP) int {
	for i, val := range iplist {
		if ip.Equal(val) {
			return i
		}
	}
	return -1
}

func findNewReplicaIP(membershipList map[uint8]membershipListEntry, filename string, failedIP net.IP, replicas []net.IP) net.IP {
	for _, listEntry := range membershipList {
		if !listEntry.IPaddr.Equal(failedIP) && listEntry.Health == Alive && checkMember(listEntry.IPaddr, replicas) == -1 {
			return listEntry.IPaddr
		}
	}
	return nil
}

func (node *SdfsNode) UploadAndModifyMap(req SdfsRequest, reply *SdfsResponse) error {
	if req.Type != UploadReq {
		return errors.New("Error: Invalid request type for Upload Request")
	}

	err := Upload(req.IPAddr.String(), fmt.Sprint(Configuration.Service.filePort), req.LocalFName, req.RemoteFName)

	if err != nil {
		fmt.Println("error in upload process.")
	}

	mapReq := SdfsRequest{LocalFName: "", RemoteFName: req.RemoteFName, IPAddr: req.IPAddr, Type: AddReq}
	var mapRes SdfsResponse
	client.Call("SdfsNode.ModifyMasterFileMap", mapReq, &mapRes)
	*reply = mapRes

	return nil
}

func (node *SdfsNode) handleReplicationOnFailure(memberID uint8) error {
	//TODO: sleep for a bit to ensure all failures quiesce before doing this?

	failedIP := node.Member.membershipList[memberID].IPaddr

	// iterate over fileMap and find files that this member stores
	for filename, ipList := range node.Master.fileMap {
		if failedIndex := checkMember(failedIP, ipList); failedIndex != -1 {
			//remove failedIP from fileMap
			node.Master.fileMap[filename] = append(ipList[:failedIndex], ipList[failedIndex+1:]...)

			// find an alive IP that doesn't already contain file
			newIP := findNewReplicaIP(node.Member.membershipList, filename, failedIP, ipList)
			if newIP == nil {
				return errors.New("No available IP to upload to for " + filename)
			}
			// choose alive IP containing the file that will upload to newIP
			var chosenIP net.IP
			for _, aliveIP := range ipList {
				if !aliveIP.Equal(failedIP) {
					chosenIP = aliveIP
					break
				}
			}
			// request chosenIP to upload file to newIP and add IP to fileMap
			sendUploadCommand(chosenIP, newIP, filename)
		}

	}
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
	// TODO: delete the replica before returning nil, else return error
	fmt.Println("DeleteFile not implemented")
	return nil
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

	ipToModify := req.IPAddr

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
