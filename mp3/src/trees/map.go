package main

import (
        "bytes"
        "os"
        "bufio"
        "fmt"
        "strings"
)

var (
        keyValues map[string][]string
)

func ReadStdin() {
        /*
        stdinReader := bufio.NewReader(os.Stdin)
        fruit := bytes.NewBuffer(make([]byte, 0))
        bytes, _ := ioutil.ReadAll(stdinReader)
        fruit.Write(bytes)
        return fruit
        */
        // MapReduce feeds data line by line
        s := bufio.NewScanner(os.Stdin)
        for scanner.Scan() {
                // Split strings by tab
                line := strings.Fields(scanner.Text())
                if len(line) == 2 {
                        if val, ok := keyValues[line[0]]; ok {
                                keyValues[line[0]] = append(keyValues[line[0]], line[1])
                                break
                        }
                        keyValues[line[0]] = []string{line[1]}
                }
        }

        if err := s.Err(); err != nil {
                fmt.Println(err)
        }
}

func Emit(key string, value string) {
        fmt.Printf("%s\t%s\n", key, value)
}

func main() {
        // ReadStdin
        ReadStdin()

        // Split each line into (key, value) pairs
        keyValues[0]

        // Loop over keys run map
}
