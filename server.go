package main

import (
  "net"
  "os"
  "fmt"
  "os/exec"
  "io/ioutil"
)

func main() {
  // Open listeners for specific port? ex. ":1200"?
  service := ":1200"
  // Resolve tcp address from hostname(s)
  tcpAddr, err := net.ResolveTCPAddr("tcp4", service)
  checkError(err)
  // Open listener
  listener, err := net.ListenTCP("tcp", tcpAddr)
  checkError(err)

  // Accept connection loop
  for {
      conn, err := listener.Accept()
      if err != nil {
        continue
      }

      // Handle grep shabang here
      // execGrep(string, filename)
      conn.Write([]byte("ack"))
      conn.Close()
  }
}

/*
  Check errors
*/
func checkError(err error) {
  if err != nil {
    panic(err)
//    fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
//    os.Exit(1)
  }
}

