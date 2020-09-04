package main

import (
  "net"
  "os"
  "fmt"
  "os/exec"
  "io/ioutil"
)

func main() {
    service := ":1200"

    tcpAddr, err := net.ResolveTCPAddr("tcp4", service)
    checkError(err)

    listerner, err := net.ListenTCP("tcp", tcpAddr)
    checkError(err)

    for {
        conn, err := listener.Accept()
        if err != nil {
            continue
        }

        conn.Write([]byte("ack"))
        conn.Close()
    }
}


func checkError(err error) {
  if err != nil {
    panic(err)
//    fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
//    os.Exit(1)
  }
}

