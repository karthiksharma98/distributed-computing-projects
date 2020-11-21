package main

import (
        "os/exec"
        "fmt"
)

/*
const (
        mapleJuiceDirName = ""
)*/

func MockFile() []byte {
        sl := make([]byte, 8 << 20)
        for i := 0; i < cap(sl); i++ {
                sl[i] = 'B'
        }
        return sl
}

/*
// Pull data and shuffle/sort into a single value
func ShuffleSort(key string, prefix string) []string {
        // Get ips of file
        filePath := path.Join([]string{mapleJuiceDirName, prefix + "_" + key}...)
        // download files of key
        //Download(ipAddr.String(), fmt.Sprint(Configuration.Service.filePort), filePath)
        // read and append data by new line
        file, err := os.Open("file")
        if err != nil {
                panic(err)
        }
        defer file.Close()

        // split input into string list and join
        vals := []string{}
        s := bufio.NewScanner(file)
        for s.Scan() {
                append(vals, s.Text())
        }
        // return location of combined data
        return vals
}*/

func Juice() {
        // run partitioner
        // run shuffle sort in local machine
        // run executejuice
}

func ExecuteJuice(exeName string, data []byte) {
        fmt.Println("Executing juice")
        juiceCmd := exec.Command(exeName)
        juiceIn, _ := juiceCmd.StdinPipe()
        juiceCmd.Start()
        juiceIn.Write(data)
        juiceIn.Close()
        juiceCmd.Wait()
        fmt.Println("Complete")
}

func main() {
        data := MockFile()
        fmt.Println("Data created")
        fmt.Println(string(data[:10]))
        ExecuteJuice("./juice", data)
}
