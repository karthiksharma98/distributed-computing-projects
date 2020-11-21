package main

import (
        "fmt"
        "os"
        "bufio"
        //"io"
        "io/ioutil"
        "bytes"
)

// run partitioner on keys
// start maple tasks for each key
type IJuice struct {
        keys []string
        values []string
}

// Emit a key/value to juicer
func (j *IJuice) Emit(key string, value string) {
        j.keys = append(j.keys, key)
        j.values = append(j.values, value)
}

// Generate a new file from keys and values emitted
func (j *IJuice) GenerateFile() map[string]string {
        filesGen := make(map[string]string)
        // for every key in list of keys
        for i, key := range j.keys {
                // concatenate values[i] + "\n"
                currStr, ok := filesGen[key]
                if !ok {
                        filesGen[key] = ""
                }
                filesGen[key] = currStr + j.values[i] + "\n"
        }
        return filesGen
}

// Save string to file
func (j *IJuice) SaveToFile(fname string, value string) {
        fileFlags := os.O_CREATE | os.O_WRONLY
        file, err := os.OpenFile(fname, fileFlags, 0777)
        if err != nil {
                fmt.Println(err)
        }
        defer file.Close()

        if _, err := file.WriteString(value); err != nil {
                fmt.Println(err)
        }
}

// Save to files based on key
func (j *IJuice) Save(prefix string) {
        // Parse keys and values into a single string
        // Every key is mapped to a string with new value
        newData := j.GenerateFile()
        for k, v := range newData {
                j.SaveToFile(prefix + "_" + k, v)
        }
}

// Save all keys to one file
func (j *IJuice) SaveAllOutput(outputFname string) {
        newData := j.GenerateFile()
        out := ""
        for k, v := range newData {
                out = out + k + " " + v + "\n"
        }
        j.SaveToFile(outputFname, out)
}

func (j *IJuice) ReadIn() *bytes.Buffer {
        stdinReader := bufio.NewReader(os.Stdin)
        data := bytes.NewBuffer(make([]byte, 0))
        /*
        // Open pipe file
        buf := make([]byte, 4 << 10)
        fmt.Println("Start read")
        for {
                n, err := stdinReader.Read(buf[:cap(buf)])

                if err == io.EOF {
                        fmt.Println("EOF reached")
                        break
                }
                if err != nil {
                        fmt.Println(err)
                        os.Exit(0)
                }
                data.Write(buf[:n])
        }*/
        bytes, _ := ioutil.ReadAll(stdinReader)
        data.Write(bytes)
        return data
}

func main() {
        // Listen to IPC or stdin for juice data
        juicerObj := IJuice{keys: make([]string,0), values: make([]string,0)}

        dat := juicerObj.ReadIn()
        juicerObj.SaveToFile("testout.txt", dat.String())
        return
}
