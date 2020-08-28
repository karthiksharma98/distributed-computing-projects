package main

import (
  "fmt"
  "syscall"
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

  fmt.Printf("Command spawn")
  spawnGrep(str, filename)
  fmt.Printf("Exec syscall spawn")
  execGrep(str, filename)
}

func spawnGrep(str string, fname string) {
  // Exec grep with args
  grepCmd := exec.Command("grep", str, fname)
  grepOut, _ := grepCmd.StdoutPipe()
  grepCmd.Start()
  // Read bytes
  grepBytes, _ := ioutil.ReadAll(grepOut)
  fmt.Println(string(grepBytes))
}

func execGrep(str string, fname string) {
  bin, lookErr := exec.LookPath("grep")

  if lookErr != nil {
    panic(lookErr)
  }

  args := []string{"grep", str, fname}

  env := os.Environ()

  execErr := syscall.Exec(bin, args, env)

  if execErr != nil {
    panic(execErr)
  }
}
