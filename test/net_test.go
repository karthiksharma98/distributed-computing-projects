package main

import (
        "net"
        "fmt"
)

type MessageType uint8

const (
        JoinMsg = iota; 
        HeartbeatMsg
        TextMsg
)

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

        // Encoded into a byte buffer of the structure:
        // MessageType uint8
        // Message byte[]
        // [0] - MessageType, [1, ...] - message
        buffer := append([]byte{byte(message_type)}, message...) // TODO: Converting to byte might not be neccessary

        _, err = conn.Write(buffer)
        if err != nil {
                panic(err)
        }
}

func Listener(port string) {
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

                // TODO: Do stuff with data
                // TODO: Need scenario to accept new members if node is introducer
                msg_type := buffer[0]
                switch msg_type {
                case TextMsg:
                        fmt.Println(string(buffer[1:n]))
                }
        }

}

