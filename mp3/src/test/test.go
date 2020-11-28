package main

import (
	"fmt"
	"os/exec"
)

const (
	juiceTempDir      = "juicetemp"
)

// Juice related jobs
type JuiceRequest struct {
	ExeName            string
	IntermediatePrefix string
	key                string
	partitionId        int
}

type JuiceTask struct {
	Request JuiceRequest
	Nodes   []net.IP
}

// (worker) Pull data and shuffle/sort into a single value
func ShuffleSort(prefix string, key string) []byte {
        prefixKey := prefix + "_" + key
        // Get ips of file
        ipList := keyLocations[prefixKey]
	sorted := make([]byte, 0)
	juiceTempPath := juiceTempDir + prefixKey
	filePath := mapleJuiceDirName + prefixKey
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

// (worker) Runs on node assigned to task
func (node *SdfsNode) RpcJuice(req JuiceRequest, reply *MapleJuiceReply) error {
	key = req.key
	pref = req.IntermediatePrefix
	exeName = "./" + req.ExeName
	// Run shuffler
	sortedFruits := ShuffleSort(key, pref)
	// Execute juicer on key
        err := ExecuteJuice(exeName, pref, key, sortedFruits)
	// Set completed
        var resp MapleJuiceReply
        resp.Completed = true
        *reply = resp
	return err
}

// (master) prompts worker machines to run juice on their uploaded blocks
func (node *SdfsNode) Juice(juiceQueueReq mapleJuiceQueueRequest) {
        // Update status
	lastStatus = JuiceOngoing
	fmt.Println("Beginning Juice phase.")
	fmt.Print("> ")

        // Get keys of given prefix
	keysMap := node.Master.prefixKeyMap[juiceQueueReq.IntermediatePrefix]
        // Mapkeys to key list
        keys := make([]int, len(keysMap))
        i := 0
        for k, _ := range keysMap {
                keys[i] = k
                i++
        }
	// Run partitioner at juice request
	numJuices := juiceQueueReq.NumTasks
	partitions := partitioner(keys, numJuices, false)
	outputFname := juiceQueueReq.FileList[0]

        // Create job scheduling structs
        juiceCh := (chan JuiceTask, len(keys))
        var wg sync.WaitGroup

        // Start workers
        node.RunJuiceWorkers(wg, juiceCh, juiceQueueReq.NumTasks)

	// Request a juice task 
	for id, keyList := range partitions {
		// Get ip list, choose node with key id % numNodes
		for _, key := range keyList {
			var req JuiceRequest
			req.ExeName = juiceQueueReq.ExeName
			req.IntermediatePrefix = juiceQueueReq.IntermediatePrefix
			req.key = key
                        req.partitionId = id
			// Get ip address of id
			nodeID := node.FindAvailableNode(req.partitionId)
			// Send Juice Request to that partition, with key to reduce
                        wg.Add(1)
                        juiceCh <- JuiceTask{[]net.IP{node.Member.membershipList[i]}, req}
		}
	}
        // Wait for all workers/tasks to complete
        wg.Wait()
        close(tasks)
	// Tasks complete, create a new file with all juice outputs
	node.CollectJuices(job, outputFname)
	// Upload file to SDFS
	sessionId := node.RpcLock(int32(node.Member.memberID), outputFname, SdfsLock)
	node.RpcPut(outputFname, outputFname)
	_ := node.RpcUnlock(sessionId, outputFname, SdfsLock)
	// indicate when it's done
	fmt.Println("Completed Juice phase.")
	fmt.Print("> ")
	lastStatus = None
	mapleJuiceCh <- None
}

// (master) Start numTasks # of juice workers
func (node *SdfsNode) RunJuiceWorkers(wg *sync.WaitGroup, tasks chan Task, numTasks int) {
        for workerId := 0; workerId < numTasks; i++ {
                go node.RunJuiceWorker(workerId, wg, tasks)
        }
}

// (master) Reschedule a juice task
func (node *SdfsNode) RescheduleJuiceTask(wg *sync.WaitGroup, task JuiceTask, tasks chan JuiceTask) {
        wg.Add(1)
        nodeId := node.FindAvailableNode(id)
        tasks <- JuiceTask{[]net.IP{node.Member.membershipList[nodeId]}, req}
        wg.Done()
}

// (master)
func (node *SdfsNode) RunJuiceWorker(id int, wg *sync.WaitGroup, tasks chan Task) {
        for task := range tasks {
                // Request juice task on a worker
                err := node.RequestJuiceTask(task.Nodes[0], task.req)
                if err != nil {
                        wg.Add(1)
                        node.RescheduleJuiceTask(wg, task, tasks)
                } else {
                        // Download juice output from corresponding file path of key + prefix
                        wg.Add(1)
                        go func() {
                                juiceFilePath := juiceTempDir + "/" + task.Request.IntermediatePrefix + "_" + task.Request.key
                                _ := Download(task.Nodes[0], fmt.Sprint(Configuration.Service.filePort), juiceFilePath, juiceFilePath)
                                wg.Done()
                        }()
                }
                wg.Done()
        }
}

// Find an available node for partition
func (node *SdfsNode) FindAvailableIP(id int) uint8 {
	id = id % len(node.Member.membershipList)
	for _, currNode := range node.Member.membershipList {
		if id == 0 {
			return currNode.MemberID
		}
		id -= 1
	}
	return nil
}

// (master) collect all juice after all tasks completed
func (node *SdfsNode) CollectJuices(job *JuiceJob, outFname string) {
	combinedJuices := ""

	for key, task := range job.DoneTasks {
		juiceFilePath := juiceTempDir + "/" + task.Request.IntermediatePrefix + "_" + key
		// Open juice output files
		file, err := os.Open(juiceFilePath)
		if err != nil {
                        file.Close()
			panic(err)
		}

		// Append to allJuices file
		s := bufio.NewScanner(file)
		for s.Scan() {
			allJuices = allJuices + s.Text() + "\n"
		}
                file.Close()
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

// (master) makes rpc call to worker machine to run maple on specific file block
func (node *SdfsNode) RequestJuiceTask(chosenIp net.IP, req JuiceRequest) {
	mapleClient, err := rpc.DialHTTP("tcp", chosenIp.String()+":"+fmt.Sprint(Configuration.Service.masterPort))
	// Call RpcMaple at chosen IP
	if err != nil {
		fmt.Println(err)
	}

	var res MapleJuiceReply
	err = mapleClient.Call("SdfsNode.RpcJuice", req, &res)
	if err != nil || !res.Completed {
		// Reschedule juicer by sending error
		fmt.Println("Error: ", err, "res.completed = ", res.Completed)
                return error;
	}
        // Complete juice and check if all juices finished
        return nil;
}

// (worker) executes juice call
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
