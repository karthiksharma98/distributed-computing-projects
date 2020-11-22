package main

import (
	"fmt"
	"os/exec"
)

const (
	mapleJuiceDirName = ""
	juiceTempDir      = "juicetemp"
)

type MapleJuiceQueueRequest struct {
	IsMaple            bool
	FileList           []string
	ExeName            string
	NumTasks           int
	IntermediatePrefix string
	DeleteInput        bool
}

// Juice related jobs
type JuiceRequest struct {
	ExeName            string
	IntermediatePrefix string
	fileName           string
	key                string
}

type JuiceJob struct {
	mu           sync.Mutex
	Request      MapleJuiceQueueRequest
	PendingTasks map[string]JuiceTask
	RunningTasks map[string]JuiceTask
	DoneTasks    map[string]JuiceTask
}

type JuiceTask struct {
	Request JuiceRequest
	Nodes   []uint8
}

// Pull data and shuffle/sort into a single value
func ShuffleSort(prefix string, key string) []byte {
	// Get ips of file
	ipList := []string{"10.0.0.1", "10.0.0.2"} // TODO: Call GetIPsForKey(prefix, key)
	sorted := make([]byte, 0)
	juiceTempPath := juiceTempDir + "/" + prefix + "_" + key
	filePath := mapleJuiceDirName + "/" + prefix + "_" + key
	for _, ipAddr := range ipList {
		// Download files of key
		err := Download(ipAddr.String(), fmt.Sprint(Configuration.Service.filePort), filePath, juiceTempPath)
		// Read and append data by new line
		content, err := ioutil.ReadFile(juiceTempPath)
		if err != nil {
			log.Fatal(err)
		}
		sorted = append(sorted, content...)
		// Remove file
		os.Remove(juiceTempPath)
	}
	// Return combined data
	return sorted
}

// Runs on node assigned to task
func (node *SdfsNode) RpcJuice(req JuiceRequest, reply *MapleJuiceReply) {
	key = req.key
	pref = req.IntermediatePrefix
	exeName = "./" + req.ExeName
	// Run shuffler
	sortedFruits := ShuffleSort(key, pref)
	// Execute juicer on key
	ExecuteJuice(exeName, pref, key, sortedFruits)
	// Set completed
	reply.Completed = true
	return err
}

// Runs on Master
func (node *SdfsNode) Juice(juiceQueueReq mapleJuiceQueueRequest) {
	// run partitioner at juice request
	keys := []string{"key1", "key2"} // TODO: Call GetAllKeys
	numJuices := juiceQueueReq.NumTasks
	partitions := partitioner(keys, numJuices, false)
	outputFname := juiceQueueReq.FileList[0]
	job := &JuiceJob{
		Request:      juiceQueueReq,
		PendingTasks: make(map[string]JuiceTask),
		RunningTasks: make(map[string]JuiceTask),
		DoneTasks:    make(map[string]JuiceTask),
	}

	// RequestJuiceTask
	for id, keyList := range partitions {
		// Get ip list, choose node with key id % numNodes
		for _, key := range keyList {
			var req JuiceRequest
			req.ExeName = juiceQueueReq.ExeName
			req.IntermediatePrefix = juiceQueueReq.IntermediatePrefix
			req.fileName = outputFname
			req.key = key
			// Send Juice Request to that partition, with key to reduce
			// Get ip address of id
			nodeID := node.FindAvailableNode(id)
			go node.RequestJuiceTask(node.membershipList[nodeID].IPaddr, req)
			// Add task to runqueue/pendingqueue
			job.EnqueueJuice(JuiceTask{[]uint8{nodeID}, req})
		}
	}
	// Tasks complete, create a new file with all juice outputs
	node.CollectJuices(job, outputFname)
	// Upload file to SDFS
	sessionId := node.RpcLock(int32(node.Member.memberID), outputFname, SdfsLock)
	node.RpcPut(outputFname, outputFname)
	_ := node.RpcUnlock(sessionId, outputFname, SdfsLock)
	return
}

// Enqueue juice task to pending or run queue
func (job *JuiceJob) EnqueueJuice(task JuiceTask) {
	job.mu.Lock()
	defer job.mu.Unlock()
	// Check if runqueue full, add if not
	if len(job.RunningTasks) < job.Request.NumTasks {
		job.RunningTasks[task.Request.key] = task
	} else {
		// Add to pending queue if runqueue full
		job.PendingTasks[task.Request.key] = task
	}
}

// Dequeue juice task from runqueue, enqueue to runqueue if tasks pending
func (job *JuiceJob) DequeueJuice(task JuiceTask, success bool) {
	job.mu.Lock()
	defer job.mu.Unlock()
	// Add to doneTasks and remove from runqueue
	if _, ok := job.RunningTasks[task.Request.key]; ok {
		if success {
			job.DoneTasks[task.Request.key] = task
		}
		delete(job.RunningTasks, task.Request.key)
	}
	// Get tasks from pending if exist, add to runqueue if possible
	if len(job.PendingTasks) > 0 && len(job.RunningTasks) < job.Request.NumTasks {
		for key, currTask := range job.PendingTasks {
			job.RunningTasks[key] = currTask
			break
		}
	}
}

// Find an available node for partition
func FindAvailableIP(id int) uint8 {
	id = id % len(node.Member.membershipList)
	for _, currNode := range node.Member.membershipList {
		if id == 0 {
			return currNode.MemberID
		}
		id -= 1
	}
	return nil
}

// Collect all juice
func (node *SdfsNode) CollectJuices(job *JuiceJob, outFname string) {
	allJuices := ""

	for key, task := range job.DoneTasks {
		// Download juice output from corresponding file path of key + prefix
		ipAddr := node.Member.membershipList[task.Nodes[0]].IPaddr
		juiceFilePath := juiceTempDir + "/" + task.Request.IntermediatePrefix + "_" + key
		_ := Download(ipAddr, fmt.Sprint(Configuration.Service.filePort), juiceFilePath, juiceFilePath)
		// Open juice output files
		file, err := os.Open(juiceFilePath)
		if err != nil {
			panic(err)
		}
		defer file.Close()

		// Append to allJuices file
		s := bufio.NewScanner(file)
		for s.Scan() {
			allJuices = allJuices + s.Text() + "\n"
		}
	}
	// Save all juices to local folder of collected juice
	fileFlags := os.O_CREATE | os.O_WRONLY
	file, err := os.OpenFile(outFname, fileFlags, 0777)
	if err != nil {
		fmt.Println(err)
	}
	defer file.Close()

	if _, err := file.WriteString(allJuices); err != nil {
		fmt.Println(err)
	}
}

func (node *SdfsNode) RequestJuiceTask(job *JuiceJob, chosenIp net.IP, req JuiceRequest) {
	mapleClient, err := rpc.DialHTTP("tcp", chosenIp.String()+":"+fmt.Sprint(Configuration.Service.masterPort))
	// Call RpcMaple at chosen IP
	if err != nil {
		fmt.Println(err)
	}

	var res MapleJuiceReply
	err = mapleClient.Call("SdfsNode.RpcJuice", req, &res)
	if err != nil || !res.Completed {
		// TODO: Reschedule juicer
		job.DequeueJuice(job.RunningTasks[req.key], false)
		fmt.Println(err)
	} else {
		// Complete juice and check if all juices finished
		job.DequeueJuice(job.RunningTasks[req.key], true)
		if len(job.RunningTasks) == 0 && len(job.PendingTasks) == 0 {
			mapleJuiceCh <- JuiceFinished
		}
	}
	return
}

/*
   case JuiceFinished:
           lastStatus =
*/

// Execute Juice executable
func ExecuteJuice(exeName string, prefix string, key string, fruits []byte) error {
	fmt.Println("Executing juice")
	juiceCmd := exec.Command(exeName, prefix, key)
	juiceIn, err := juiceCmd.StdinPipe()
	if err != nil {
		return err
	}
	juiceCmd.Start()
	// Write map output data to pipe
	juiceIn.Write(fruits)
	juiceIn.Close()
	juiceCmd.Wait()
	return nil
}

func main() {
	fruits := MockFile()
	fmt.Println("Fruits created")
	fmt.Println(string(fruits[:10]))
	exeName := "juice"
	ExecuteJuice("./"+exeName, "temp", fruits)
}
