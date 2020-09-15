package main

import (
        "net"
        "fmt"
)

func SendMessage(address string, message string) {
        addr, err := net.ResolveUDPAddr("udp", address)
        if err != nil {
                panic(err)
        }

        // Get UDP "connection"
        conn, err := net.DialUDP("udp", nil,  addr)
        if err != nil {
                panic(err)
        }

        _, err = conn.Write([]byte(message))
        if err != nil {
                panic(err)
        }
	return;
}

func RecieveAll (port string) {
        // UDP buffer 1024 bytes for now
        buffer := make([]byte, 1024)
        addr, err := net.ResolveUDPAddr("udp", ":" + port)
        if err != nil {
                panic(err)
        }

        listener, err := net.ListenUDP("udp", addr)
        if err != nil {
                panic(err)
        }
        // TODO: goroutine for non-blocking listener
        // listener loop
        for {
                n, _, err := listener.ReadFrom(buffer)
                if err != nil {
                        return
                }
                fmt.Println(string(buffer[0:n]))

                // TODO: do stuff with data
                // TODO: Need scenario to accept new members if node is introducer
        }

}

