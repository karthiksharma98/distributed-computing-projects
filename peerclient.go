package main

import (
  "net"
  "os"
  "fmt"
)

func main() {
  // get args
  if len(os.Args) != 2 {
    fmt.Fprintf(os.Stderr, "Bad arg", os.Args[0])
    os.Exit(1)
  }
  service := os.Args[1]
  
  // Resolve tcp address from hostname
  tcpAddr, err := net.ResolveTCPAddr("tcp4", service)

  // Open socket connection with tcpAddr
  //conn, err := net.DialTCP("tcp", nil, tcpAddr)
  checkError(err)

  //addr := net.ParseIP(tcpAddr)

  if tcpAddr == nil {
    fmt.Println("Invalid address")
  } else {
    fmt.Println("The address is ", tcpAddr.String())
  }

  fmt.Println("Hello")
  os.Exit(0)
}

func checkError(err error) {
  if err != nil {
    fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
    os.Exit(1)
  }
}
