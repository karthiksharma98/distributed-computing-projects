package main

import (
	"fmt"
	"net"
	"math/rand"
)

type MessageType uint8

const (
	JoinMsg = iota
	HeartbeatMsg
	TextMsg
	AcceptMsg
	GrepReq
	GrepResp
	SwitchMsg
)

// Debugging consts
const (
	dropMessage = false
	dropRate    = 20
)

// Send text message over UDP given address and string
func SendMessage(address string, msg string) {
	Send(address, TextMsg, []byte(msg))
}

// Broadcast message over UDP given addresses, messagetype, msg
func SendBroadcast(addresses []string, msgType MessageType, msg []byte) {
	for _, addr := range addresses {
		Send(addr, msgType, msg)
	}
}

// Sends message over UDP given address, messagetype, msg
func Send(address string, msgType MessageType, msg []byte) {
	// Debug purposes: simulate message drop
	if dropMessage && rand.Intn(100) < dropRate {
		return
	}

	addr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		panic(err)
	}

	// Get UDP "connection"
	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		panic(err)
	}

	// Encoded into a byte buffer of the structure:
	// MessageType uint8
	// Message byte[]
	// [0] - MessageType, [1, ...] - message
	buffer := append([]byte{byte(msgType)}, msg...) // TODO: Converting to byte might not be neccessary

	_, err = conn.Write(buffer)
	if err != nil {
		panic(err)
	}
}

// Listener function that listens to a port and prints incoming TextMsg
func Listener(port string) {
	// UDP buffer 1024 bytes for now
	buffer := make([]byte, 1024)
	addr, err := net.ResolveUDPAddr("udp", ":"+port)
	if err != nil {
		panic(err)
	}

	listener, err := net.ListenUDP("udp", addr)
	if err != nil {
		panic(err)
	}

	// listener loop
	for {
		n, _, err := listener.ReadFrom(buffer)
		if err != nil {
			return
		}

		msg_type := buffer[0]
		switch msg_type {
		case TextMsg:
			fmt.Println(string(buffer[1 : n-1]))
		}
	}

}
