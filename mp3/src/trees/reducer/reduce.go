package main

import (
        "fmt"
        "bufio"
        "strings"
        "os"
)

var (
        keyValues map[string][]string
)

func CollectReduce() {
        // MapReduce feeds data line by line
        s := bufio.NewScanner(os.Stdin)
        for s.Scan() {
                // Split strings by tab
                line := strings.Fields(s.Text())
                if len(line) == 2 {
                        if val, ok := keyValues[line[0]]; ok {
                                keyValues[line[0]] = append(val, line[1])
                                break
                        }
                        keyValues[line[0]] = []string{line[1]}
                }
        }

        if err := s.Err(); err != nil {
                fmt.Println(err)
        }
}

func RunReduce() {
        CollectReduce()
        for key, values := range keyValues {
                Reduce(key, values)
        }
}

func Emit(key string, value string) {
        fmt.Printf("%s\t%s\n", key, value)
}

func main() {
        RunReduce()
}
