package main

import (
        "os/exec"
        "fmt"
)

/*
const (
        mapleJuiceDirName = ""
)*/

type JuiceRequest struct {
        ExeName string
        IntermediatePrefix string
        fileName string
        keys []string
}

func MockFile() []byte {
        sl := make([]byte, 8 << 20)
        for i := 0; i < cap(sl); i++ {
                sl[i] = 'B'
        }
        return sl
}

// Pull data and shuffle/sort into a single value
func ShuffleSort(key string, prefix string) string {
        // Get ips of file
        //filePath := path.Join([]string{mapleJuiceDirName, prefix + "_" + key}...)
        // download files of key
        //Download(ipAddr.String(), fmt.Sprint(Configuration.Service.filePort), filePath)
        // read and append data by new line
        file, err := os.Open("file")
        if err != nil {
                panic(err)
        }
        defer file.Close()

        // split input into string list and join
        sorted := ""
        s := bufio.NewScanner(file)
        for s.Scan() {
                sorted = sorted + s.Text() + "\n"
        }
        // return location of combined data
        return vals
}

func RpcJuice(req JuiceRequest, reply *JuiceReply) {
        keys = []string{"placeholder1", "placeholder2"}
        pref = "placeholder_prefix"
        exeName = "placeholder_exe"
        for _, key := range keys {
                // Run shuffler
                sortedFruits := ShuffleSort(key, pref)
                // Execute juicer
                ExecuteJuice(exeName, []byte(sortedFruits))
        }
        return
}

func Juice(mapleQueueReq MapleJuiceQueueRequest) {
        // run partitioner at juice request
        keys := []string{"key1", "key2"}
        numJuices := 5
        partitions := partitioner(keys, numJuices, false)

        // RequestJuiceTask
        for k, list := range partitions {
                // Get ip list, choose node with key id % numNodes
                var req JuiceRequest
                req.ExeName = mapleQueueReq.ExeName
                req.IntermediatePrefix = mapleQueueReq.IntermediatePrefix
                req.fileName = "placeholder"
                req.key = list
                // Send Juice Request to that partition, with key to reduce
                chosenIp := "10.0.0.1"
                go RequestJuiceTask(chosenIp, req)
                // Add task to runqueue
        }
        return
}

func RequestJuiceTask(chosenIp string, req JuiceRequest) {
        // DialHTTP
        // Call RpcMaple at chosen IP
        return
}

// Execute Juice executable
func ExecuteJuice(exeName string, fname string, fruits []byte) {
        fmt.Println("Executing juice")
        juiceCmd := exec.Command(exeName, fname)
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
