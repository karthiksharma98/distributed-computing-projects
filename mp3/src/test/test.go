package main

import (
        "os/exec"
        "fmt"
)

const (
        mapleJuiceDirName = ""
)

type JuiceRequest struct {
        ExeName string
        IntermediatePrefix string
        fileName string
        keys []string
}

// Pull data and shuffle/sort into a single value
func ShuffleSort(prefix string, key string) string {
        // Get ips of file
        filePath := path.Join([]string{mapleJuiceDirName, prefix + "_" + key}...)

        ipList := []string{"10.0.0.1", "10.0.0.2"} // TODO: Call getIPsForKey
        sorted := ""
        for _, ipAddr := range ipList {
                // download files of key
                err := Download(ipAddr.String(), fmt.Sprint(Configuration.Service.filePort), mapleJuiceDirName, prefix + "_" + key)
                // read and append data by new line
                file, err := os.Open(filePath)
                if err != nil {
                        panic(err)
                }
                defer file.Close()

                // split input into string list and join
                s := bufio.NewScanner(file)
                for s.Scan() {
                        sorted = sorted + s.Text() + "\n"
                }
                // Remove file
                os.Remove(filePath)
        }
        // return location of combined data
        return sorted
}

func RpcJuice(req JuiceRequest, reply *JuiceReply) {
        keys = []string{"placeholder1", "placeholder2"}
        pref = req.IntermediatePrefix
        exeName = "./" + req.ExeName
        for _, key := range keys {
                // Run shuffler
                sortedFruits := ShuffleSort(key, pref)
                // Execute juicer on key
                ExecuteJuice(exeName, pref, key, []byte(sortedFruits))
        }
        return
}

func Juice(juiceQueueReq MapleJuiceQueueRequest) {
        // run partitioner at juice request
        keys := []string{"key1", "key2"}
        numJuices := 5
        partitions := partitioner(keys, numJuices, false)
        outputFname := juiceQueueReq.FileList[0]

        // RequestJuiceTask
        for id, list := range partitions {
                // Get ip list, choose node with key id % numNodes
                var req JuiceRequest
                req.ExeName = juiceQueueReq.ExeName
                req.IntermediatePrefix = juiceQueueReq.IntermediatePrefix
                req.fileName = outputFname
                req.key = list
                // Send Juice Request to that partition, with key to reduce
                // Get ip address of id
                chosenIp := "10.0.0.1"
                go RequestJuiceTask(chosenIp, req)
                // Add task to runqueue
        }
        // TODO: After all tasks complete, create a new file for output by combining all juice outputs
        // TODO: Download juice output from corresponding SDFS file (JuiceCollector)
        // TODO: Open juice output files
        // TODO: Join and append together
        filePath := path.Join([]string{mapleJuiceDirName, outputFname}...)
        CollectJuices(filePath, juiceQueueReq.FileList[0]partitions)
        // TODO: Upload file to SDFS
        if sdfs == nil {
                fmt.Println("Unable to upload juice output to SDFS.")
        }
        sessionId := sdfs.RpcLock(int32(sdfs.Member.memberID), outputFname, SdfsLock)
        sdfs.RpcPut(filePath, outputFname)
        _ := sdfs.RpcUnlock(sessionId, outputFname, SdfsLock)
        return
}

func CollectJuices(outFname string, partitions map[int][]string) {
        allJuices := ""
        for id, list := range partitions {
                // TODO: Get ip address of k
                ipAddr := "10.0.0.1"
                for _, key := range list {
                        filePath := path.Join([]string{mapleJuiceDirName, prefix + "_" + key}...)
                        _ := Download(ipAddr, fmt.Sprint(Configuration.Service.filePort), mapleJuiceDirName, prefix + "_" + key)
                        // TODO: Open file
                        file, err := os.Open(filePath)
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
        }
        // TODO: Save all juices to local folder of collected juice
        fileFlags := os.O_CREATE | os.O_WRONLY
        file, err := os.OpenFile(filePath, fileFlags, 0777)
        if err != nil {
                fmt.Println(err)
        }
        defer file.Close()

        if _, err := file.WriteString(allJuices); err != nil {
                fmt.Println(err)
        }
}

func RequestJuiceTask(chosenIp string, req JuiceRequest) {
        // DialHTTP
        mapleClient, err := rpc.DialHTTP("tcp", chosenIp+":"+fmt.Sprint(Configuration.Service.masterPort))
        // Call RpcMaple at chosen IP
        if err != nil {
                fmt.Println(err)
        }

        var res MapleJuiceReply
        err = mapleClient.Call("SdfsNode.RpcJuice", req, &res)
        if err != nil || !res.Completed {
                fmt.Println(err)
        }

        return
}

// Execute Juice executable
func ExecuteJuice(exeName string, prefix string, key string, fruits []byte) {
        fmt.Println("Executing juice")
        juiceCmd := exec.Command(exeName, prefix, key)
        juiceIn, _ := juiceCmd.StdinPipe()
        juiceCmd.Start()
        juiceIn.Write(fruits)
        juiceIn.Close()
        juiceCmd.Wait()
        fmt.Println("Complete")
}

func main() {
        fruits := MockFile()
        fmt.Println("Fruits created")
        fmt.Println(string(fruits[:10]))
        exeName := "juice"
        ExecuteJuice("./" + exeName, "temp", fruits)
}
