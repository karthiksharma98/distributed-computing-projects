package main

import (
        "net"
        "fmt"
)

type MessageType uint8

const (
        JoinMsg MessageType = iota; 
        HeartbeatMsg
        TextMsg
)

/*
type Message struct {
        message_type MessageType
        buffer []byte
}*/


func SendMessage(address string, message string) {
        Send(address, TextMsg, []byte(message))
}

func Send(address string, message_type MessageType, message []byte) {
        addr, err := net.ResolveUDPAddr("udp", address)
        if err != nil {
                panic(err)
        }

        // Get UDP "connection"
        conn, err := net.DialUDP("udp", nil,  addr)
        if err != nil {
                panic(err)
        }

        buffer := append([]byte{message_type}, message)

        _, err = conn.Write(buffer)
        if err != nil {
                panic(err)
        }
}

func RecieveAll(port string) {
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

