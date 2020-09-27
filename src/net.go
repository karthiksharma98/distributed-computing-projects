package main

import (
	"math/rand"
	"net"
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
var (
	dropMessage = false
	dropRate    = 1
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
		memMetrics.Increment(messageDrop, 1)
		return
	}
	memMetrics.Increment(messageSent, 1)

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
