package main

import (
  "fmt"
  "os"
  "os/exec"
  "io/ioutil"
)

func main() {
  // Args: wrapper <string> <filename> 
  if len(os.Args) != 3 {
    fmt.Fprintf(os.Stderr, "Bad arg", os.Args[0])
    os.Exit(1)
  }

  str := os.Args[1]
  filename := os.Args[2]
  
  // Exec grep with args
  grepCmd := exec.Command("grep", str, filename)
  grepOut, _ := grepCmd.StdoutPipe()
  grepCmd.Start()
  // Read bytes
  grepBytes, _ := ioutil.ReadAll(grepOut)
  fmt.Println(string(grepBytes))
}


