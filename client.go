package main

import (
  "net"
  "os"
  "fmt"
  "os/exec"
  "io/ioutil"
)

var peers [0]string = []

func main() {
  if len(os.Args) != 2 {
    fmt.Fprintf(os.Stderr, "Bad arg", os.Args[0])
    os.Exit(1)
  }

  // Open socket connection to each peer
  for i := 0; i < len(peers); i++ {
    connectPeer(peer, os.Args[1])
  }
}

/*
  connectPeer
    Open socket to specific peer
*/
func connectPeer(host string, grepArg string) {
  // Resolve tcp address from hostname(s)
  tcpAddr, err := net.ResolveTCPAddr("tcp4", peer) 

  // Open socket connection with tcpAddr
  conn, err := net.DialTCP("tcp", nil, tcpAddr)
  checkError(err)
  // conn.Write([]byte(grepArg))

  // Read out output
  result, err := ioutil.ReadAll(conn)
  checkError(err)
  fmt.Println(string(result))
}

/*
  Check errors
*/
func checkError(err error) {
  if err != nil {
    panic(err)
  }
}

