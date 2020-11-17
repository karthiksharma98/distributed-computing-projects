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
	fileName           string
	blockNum           int
}

type MapleJuiceReply struct {
	Completed bool
}

type Task struct {
	Request  MapleRequest
	Replicas []net.IP
}

type Status int

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

func (node *SdfsNode) RemoveFromQueue(idx int) {
	queue = append(queue[:idx], queue[idx+1:]...)
}

// call master to add task to its queue
func (node *SdfsNode) QueueTask(mapleQueueReq MapleJuiceQueueRequest) error {
	var res MapleJuiceReply
	return client.Call("SdfsNode.AddToQueue", mapleQueueReq, &res)
}

// master's function to add to the blocking queue
func (node *SdfsNode) AddToQueue(mapleQueueReq MapleJuiceQueueRequest, reply *MapleJuiceReply) error {
	queue = append(queue, mapleQueueReq)

	if mapleQueueReq.IsMaple {
		mapleJuiceCh <- RequestingMaple
	} else {
		mapleJuiceCh <- RequestingJuice
	}

	return nil
}

// master prompts worker machines to run maple on their uploaded blocks
func (node *SdfsNode) Maple(mapleQueueReq MapleJuiceQueueRequest) {
	lastStatus = MapleOngoing

	// TODO: don't just read from file map
	for fileName, blockMap := range node.Master.fileMap {
		// initiate maple on each block of each file
		var req MapleRequest
		req.ExeName = mapleQueueReq.ExeName
		req.IntermediatePrefix = mapleQueueReq.IntermediatePrefix
		req.fileName = sdfsDirName + "/" + fileName

		numBlocks := node.Master.numBlocks[fileName]
		for i := 0; i < numBlocks; i++ {
			req.blockNum = i

			// start with calling maple on the first ip
			ips := blockMap[i]
			chosenIp := ips[0]
			go node.RequestMapleOnBlock(chosenIp.String(), req)

			currTasks[req.fileName] = Task{req, ips}
		}

	}
}

// master makes rpc call to worker machine
func (node *SdfsNode) RequestMapleOnBlock(chosenIp string, req MapleRequest) {
	mapleClient, err := rpc.DialHTTP("tcp", chosenIp+":"+fmt.Sprint(Configuration.Service.masterPort))
	if err != nil {
		fmt.Println("Error in connecting to maple client ", err)
		node.RescheduleTask(req.fileName)
	}

	var res MapleJuiceReply
	err = mapleClient.Call("SdfsNode.RpcMaple", req, &res)
	if err != nil || !res.Completed {
		fmt.Println("Error: ", err, "res.completed = ", res.Completed)
		node.RescheduleTask(req.fileName)
	} else {
		node.MarkCompleted(req.fileName)
	}
}

func (node *SdfsNode) Juice(mapleQueueReq MapleJuiceQueueRequest) {
	lastStatus = JuiceOngoing

	// TODO: shuffling
	// do some juice stuff

	// indicate when it's done
	fmt.Println("Completed Juice phase.")
	fmt.Print("> ")
	lastStatus = None
	mapleJuiceCh <- None
}

// master receives acknowledgement from worker that it finished a file block
func (node *SdfsNode) MarkCompleted(fileName string) {
	delete(currTasks, fileName)

	if len(currTasks) == 0 {
		fmt.Println("Completed Maple phase.")
		fmt.Print("> ")
		mapleJuiceCh <- MapleFinished
	}
}

// master reassigns failed node's ongoing tasks, if any
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

// reschedule task to another machine that has that file
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

			node.RequestMapleOnBlock(replicas[0].String(), currTasks[fileName].Request)
		}
	}

}

// worker machine receives Request to run a maple_exe on some file block
func (node *SdfsNode) RpcMaple(req MapleRequest, reply *MapleJuiceReply) error {
	// format: fileName.blk_#
	blockNum := strconv.Itoa(req.blockNum)
	filePath := req.fileName + ".blk_" + blockNum

	var response MapleJuiceReply

	app := "bash"
	arg0 := "./" + req.ExeName
	arg1 := filePath

	cmd := exec.Command(app, arg0, arg1)
	output, err := cmd.Output()
	if err != nil {
		fmt.Println("Error in executing maple.")
		response.Completed = false
	} else {
		response.Completed = WriteMapleKeys(string(output), req.IntermediatePrefix)
	}

	*reply = response
	return err
}

// Scan output in the format [key,key's value and store to intermediate files by key
func WriteMapleKeys(output string, prefix string) bool {
	// read output one line at a time
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		keyVal := strings.Split(scanner.Text(), ",")
		if len(keyVal) < 2 {
			continue
		}
		key := keyVal[0]
		val := keyVal[1]

		// TODO: convert key to an appropriate string for a file name
		keyString := key

		// write to MapleJuice/prefix_key
		// key -> ip
		// need method to get all keys
		filePath := path.Join([]string{mapleJuiceDirName, prefix + "_" + keyString}...)
		f, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
		if err != nil {
			fmt.Println("Error opening ", filePath, ". Error: ", err)
			continue
		}
		defer f.Close()

		if _, err := f.WriteString(val + "\n"); err != nil {
			fmt.Println("Error writing val [", val, "] to ", filePath, ". Error: ", err)
		}
	}

	return true
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
