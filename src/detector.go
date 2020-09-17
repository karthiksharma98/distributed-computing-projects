package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"net"
	"time"
)

// Health enum
const (
	Alive = iota
	Failed
	Left
)

// Member struct to hold member info
type Member struct {
	memberID       uint8
	isIntroducer   bool
	membershipList map[uint8]membershipListEntry // {uint8 (member_id): Member}
}

// holds one entry of the membership list
type membershipListEntry struct {
	MemberID       uint8
	IPaddr         net.IP
	HeartbeatCount uint64
	Timestamp      time.Time
	Health         uint8 // -> Health enum
}

// Listen function to keep listening for messages
func (mem Member) Listen(port string) {
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

	// TODO: goroutine for non-blocking listener
	// listener loop
	for {
		n, senderAddr, err := listener.ReadFromUDP(buffer)
		if err != nil {
			return
		}

		msgType := buffer[0]
		switch msgType {
		case TextMsg:
			fmt.Println(string(buffer[1:n]))
		case JoinMsg:
			// only introducer can accept join messages
			if mem.isIntroducer == true {
				mem.acceptMember(senderAddr.IP)
			}
		case HeartbeatMsg: // handles receipt of heartbeat

		case AcceptMsg: // handles receipt of membership list from introducer
			mem.joinResponse(buffer[1:n])

		default:
			fmt.Println("Invalid message type")
		}
	}

}

// request introducer to join
func (mem Member) joinRequest() {
	Send(ServiceInfo["introducer_ip"].(string)+":"+fmt.Sprint(ServiceInfo["port"]), JoinMsg, nil)
}

// receive membership list from introducer and setup
func (mem Member) joinResponse(membershipListBytes []byte) {
	// First byte received corresponds to assigned memberID
	mem.memberID = uint8(membershipListBytes[0])

	// Decode the rest of the buffer to the membership list
	b := bytes.NewBuffer(membershipListBytes[1:])
	d := gob.NewDecoder(b)
	err := d.Decode(&mem.membershipList)
	if err != nil {
		panic(err)
	}

	fmt.Println(mem.membershipList)
}

// modify membership list entry
func (mem Member) leave() {
	newEntry := mem.membershipList[mem.memberID]
	newEntry.HeartbeatCount++
	newEntry.Health = Left
	newEntry.Timestamp = time.Now()
	mem.membershipList[mem.memberID] = newEntry

	// TODO: kill the leaving process after a certain time/# of heartbeats
}

// for introducer to accept a new member
func (mem Member) acceptMember(address net.IP) {
	// assign new ID
	newMemberID := GetMaxKey(mem.membershipList) + 1
	mem.membershipList[newMemberID] = membershipListEntry{newMemberID, address, 0, time.Now(), Alive}

	// Encode the membership list to send it
	b := new(bytes.Buffer)
	e := gob.NewEncoder(b)
	err := e.Encode(mem.membershipList)
	if err != nil {
		panic(err)
	}

	// Send the memberID by appending it to start of buffer, and the membershiplist
	Send(address.String()+":"+fmt.Sprint(ServiceInfo["port"]), AcceptMsg, append([]byte{newMemberID}, b.Bytes()...))
}

// GetMaxKey to get the maximum of all memberIDs
func GetMaxKey(list map[uint8]membershipListEntry) uint8 {
	var result uint8
	for result = range list {
		break
	}
	for n := range list {
		if n > result {
			result = n
		}
	}
	return result
}
