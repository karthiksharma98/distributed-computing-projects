package main

import (
  "net"
  "os"
  "fmt"
  "os/exec"
  "io/ioutil"
)

func main() {
  if len(os.Args) != 2 {
    fmt.Fprintf(os.Stderr, "Bad arg", os.Args[0])
    os.Exit(1)
  }

  tcpAddr, err := net.ResolveTCPAddr("tcp4", os.Args[1])
  checkError(err)

  conn, err = net.DialTCP("tcp", nil, tcpAddr)
  checkError(err)

  result, err := ioutil.ReadAll(conn)
  checkError(err)
  fmt.Println(string(result))

}

func checkError(err error) {
  if err != nil {
    panic(err)
  }
}
