package main

import (
  "net"
  "os"
  "fmt"
  "os/exec"
  "io/ioutil"
)

func main() {
  // Args: peerclient <hostname>
  if len(os.Args) != 2 {
    fmt.Fprintf(os.Stderr, "Bad arg", os.Args[0])
    os.Exit(1)
  }

  // Open listeners for specific port? ex. ":1200"?
  service := os.Args[1] //":1200"
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
  pingPeer
  Ping a peer from list
*/
func pingPeers() {
  // Need list of distributed peers (ip addr, etc)
}

/*
  connectPeer
  Connect to specific peer
*/
func connectPeer(host string, port string) {
  // Resolve tcp address from hostname(s)
  peer := host + ":" + port
  tcpAddr, err := net.ResolveTCPAddr("tcp4", service)

  // Open socket connection with tcpAddr
  conn, err := net.DialTCP("tcp", nil, tcpAddr)
  checkError(err)

  // Get ouput
  result, err := ioutil.ReadAll(conn)
  checkError(err)
  fmt.Println(string(result))
}

/*
  validate tcp
*/
func checkTcp(tcpAddr string) {
  addr := net.ParseIP(tcpAddr)
  if tcpAddr == nil {
    fmt.Println("Invalid address")
  } else {
    fmt.Println("The address is ", tcpAddr.String())
  }
  return
}

/*
  execGrep
  Run grep and return output given log file name and string
func execGrep() {

}
*/

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
