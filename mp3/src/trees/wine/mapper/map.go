package main

import (
        "os"
        "bufio"
        "fmt"
)

func RunMap() {
        // MapReduce feeds data line by line
        s := bufio.NewScanner(os.Stdin)
        for s.Scan() {
                // Split strings by tab
                Map("", s.Text())
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
        RunMap()
}
