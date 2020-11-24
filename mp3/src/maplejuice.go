package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"net"
	"net/rpc"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
)

type MapleJuiceQueueRequest struct {
	IsMaple            bool
	FileList           []string
	ExeName            string
	NumTasks           int
	IntermediatePrefix string
	DeleteInput        bool
}

type MapleRequest struct {
	ExeName            string
	IntermediatePrefix string
	FileName           string
	BlockNum           int
}

type MapleJuiceReply struct {
	Completed bool
	KeyList   []string
}

type Task struct {
	Request  MapleRequest
	Replicas []net.IP
}

type Status int

type Mapler interface {
	Maple(inputFilePath string) error //takes location of inputFile as input and writes to stdout
}

const (
	None Status = iota
	RequestingMaple
	MapleOngoing
	MapleFinished
	RequestingJuice
	JuiceOngoing
)

var (
	mapleJuiceCh = make(chan Status, 1)
	queue        []MapleJuiceQueueRequest
	currTasks           = make(map[string]Task) // fileName -> [replicaIPs]
	lastStatus   Status = None
)

// (master) listens for changes in the maple/juice process for the task run queue
func (node *SdfsNode) ListenMapleJuice() {
	for {
		// blocks until there is a change in the status
		switch status := <-mapleJuiceCh; status {
		case None:
			node.RunFirstMaple()

		case RequestingMaple:
			if lastStatus == None {
				node.RunFirstMaple()
			}

		case MapleFinished:
			lastStatus = MapleFinished
			node.RunFirstJuice()

		case RequestingJuice:
			if lastStatus == MapleFinished {
				node.RunFirstJuice()
			}

		default:
			break
		}
	}
}

// (master) loop through the queue and initiate maple on the first maple task
func (node *SdfsNode) RunFirstMaple() {
	// initiate maple
	for idx := 0; idx < len(queue); idx += 1 {
		task := queue[idx]
		if task.IsMaple {
			node.Maple(task)
			node.RemoveFromQueue(idx)
			break
		}
	}
}

// (master) loop through the queue and initiate juice on the first maple task
func (node *SdfsNode) RunFirstJuice() {
	for idx := 0; idx < len(queue); idx += 1 {
		task := queue[idx]
		if !task.IsMaple {
			node.Juice(task)
			node.RemoveFromQueue(idx)
			break
		}
	}
}

// (master) remove task from queue
func (node *SdfsNode) RemoveFromQueue(idx int) {
	queue = append(queue[:idx], queue[idx+1:]...)
}

// (worker) call master to add task to its queue
func (node *SdfsNode) QueueTask(mapleQueueReq MapleJuiceQueueRequest) error {
	var res MapleJuiceReply
	return client.Call("SdfsNode.AddToQueue", mapleQueueReq, &res)
}

// (master) add task to blocking queue and fill channel to notify the listener
func (node *SdfsNode) AddToQueue(mapleQueueReq MapleJuiceQueueRequest, reply *MapleJuiceReply) error {
	queue = append(queue, mapleQueueReq)

	if mapleQueueReq.IsMaple {
		mapleJuiceCh <- RequestingMaple
	} else {
		mapleJuiceCh <- RequestingJuice
	}

	return nil
}

// (master) prompts worker machines to run maple on their uploaded blocks
func (node *SdfsNode) Maple(mapleQueueReq MapleJuiceQueueRequest) {
	lastStatus = MapleOngoing
	fmt.Println("Beginning Map phase.")
	fmt.Print("> ")

	for _, localFName := range mapleQueueReq.FileList {
		sdfsFName := node.Master.sdfsFNameMap[localFName]
		if blockMap, ok := node.Master.fileMap[sdfsFName]; ok {
			// initiate maple on each block of each file
			var req MapleRequest
			req.ExeName = mapleQueueReq.ExeName
			req.IntermediatePrefix = mapleQueueReq.IntermediatePrefix
			req.FileName = sdfsDirName + "/" + sdfsFName

			numBlocks := node.Master.numBlocks[sdfsFName]
			for i := 0; i < numBlocks; i++ {
				req.BlockNum = i

				// start with calling maple on the first ip
				ips := blockMap[i]
				chosenIp := ips[0]

				go node.RequestMapleOnBlock(chosenIp, req)

				currTasks[req.FileName] = Task{req, ips}
			}
		}

	}
}

// (master) makes rpc call to worker machine to run maple on a specified file block
func (node *SdfsNode) RequestMapleOnBlock(chosenIp net.IP, req MapleRequest) {
	mapleClient, err := rpc.DialHTTP("tcp", chosenIp.String()+":"+fmt.Sprint(Configuration.Service.masterPort))
	if err != nil {
		fmt.Println("Error in connecting to maple client ", err)
		node.RescheduleTask(req.FileName)
	}

	var res MapleJuiceReply
	err = mapleClient.Call("SdfsNode.RpcMaple", req, &res)
	if err != nil || !res.Completed {
		fmt.Println("Error: ", err, "res.completed = ", res.Completed)
		node.RescheduleTask(req.FileName)
	} else {
		node.MarkCompleted(req.FileName)
		for _, key := range res.KeyList {
			node.Master.keyLocations[key] = append(node.Master.keyLocations[key], chosenIp)
		}
	}
}

// (master) prompts worker machines to run juice on their uploaded blocks
func (node *SdfsNode) Juice(mapleQueueReq MapleJuiceQueueRequest) {
	lastStatus = JuiceOngoing
	fmt.Println("Beginning Juice phase.")
	fmt.Print("> ")

	// TODO: shuffling
	// do some juice stuff

	// indicate when it's done
	fmt.Println("Completed Juice phase.")
	fmt.Print("> ")
	lastStatus = None
	mapleJuiceCh <- None
}

// (master) receive acknowledgement from worker that it finished a file block
func (node *SdfsNode) MarkCompleted(fileName string) {
	delete(currTasks, fileName)

	if len(currTasks) == 0 {
		fmt.Println("Completed Maple phase.")
		fmt.Print("> ")
		mapleJuiceCh <- MapleFinished
	}
}

// (master) reassigns failed node's ongoing tasks, if any
// 		and remove from replicas list, otherwise
func (node *SdfsNode) HandleTaskReassignments(memberId uint8) {
	failedIp := node.Member.membershipList[memberId].IPaddr
	for fileName, task := range currTasks {
		ips := task.Replicas
		for i := 0; i < len(ips); i += 1 {
			if ips[i].Equal(failedIp) {
				if i == 0 {
					// ongoing task
					node.RescheduleTask(fileName)
				} else {
					// remove from list
					newIps := append(ips[:i], ips[i+1:]...)
					currTasks[fileName] = Task{task.Request, newIps}
				}
			}

		}
	}
}

// (master) reschedule task to another machine that has that file
// 			initiated when a worker has failed
func (node *SdfsNode) RescheduleTask(fileName string) {
	fmt.Println("Rescheduling task ", fileName)
	if task, ok := currTasks[fileName]; ok {
		replicas := task.Replicas
		if len(replicas) < 2 {
			// no more to try
			// TODO: re-replicate file elsewhere, and try again?
			//		for now, delete + ignore lol
			delete(currTasks, fileName)
			Err.Println("Could not successfully finish task on ", fileName)
		} else {
			replicas = replicas[1:]
			currTasks[fileName] = Task{task.Request, replicas}

			node.RequestMapleOnBlock(replicas[0], currTasks[fileName].Request)
		}
	}

}

// (worker) receives Request to run a maple_exe on some file block from the master
func (node *SdfsNode) RpcMaple(req MapleRequest, reply *MapleJuiceReply) error {
	// format: fileName.blk_#
	blockNum := strconv.Itoa(req.BlockNum)
	filePath := req.FileName + ".blk_" + blockNum

	var response MapleJuiceReply

	app := "./" + req.ExeName
	arg0 := filePath

	cmd := exec.Command(app, arg0)

	output, err := cmd.Output()

	if err != nil {
		fmt.Println("Error in executing maple.")
		response.Completed = false
	} else {
		response = WriteMapleKeys(string(output), req.IntermediatePrefix)
	}

	*reply = response
	return err
}

// (worker) Scan output of maple in the format [key,key's value] and store to intermediate files by key
func WriteMapleKeys(output string, prefix string) MapleJuiceReply {
	// read output one line at a time
	scanner := bufio.NewScanner(strings.NewReader(output))
	keySet := make(map[string]bool)
	for scanner.Scan() {
		keyVal := strings.Split(scanner.Text(), ",")
		if len(keyVal) < 2 {
			continue
		}
		key := keyVal[0]
		val := keyVal[1]

		// TODO: convert key to an appropriate string for a file name
		keyString := key
		prefixKey := prefix + "_" + keyString

		// write to MapleJuice/prefix_key
		// key -> ip
		// need method to get all keys
		filePath := path.Join([]string{mapleJuiceDirName, prefix + "_" + keyString}...)
		f, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)

		if err != nil {
			fmt.Println("Error opening ", filePath, ". Error: ", err)
			continue
		}

		if _, err := f.WriteString(val + "\n"); err != nil {
			fmt.Println("Error writing val [", val, "] to ", filePath, ". Error: ", err)
		}
		keySet[prefixKey] = true
		f.Close()
	}
	keyList := make([]string, 0, len(keySet))
	for k := range keySet {
		keyList = append(keyList, k)
	}
	return MapleJuiceReply{Completed: true, KeyList: keyList}
}

// locally grab files in the directory
func GetFileNames(dirName string) []string {
	var fileNames []string

	files, err := ioutil.ReadDir(dirName)
	if err != nil {
		fmt.Println("Error: ", dirName, " is not a valid directory.")
	} else {
		for _, f := range files {
			fileNames = append(fileNames, f.Name())
		}
	}

	return fileNames
}
