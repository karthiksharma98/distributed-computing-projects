package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"math/rand"
	"net"
	"time"
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

// Health enum
const (
	Alive = iota
	Failed
	Left
)

// Ticker variables
var (
	disableHeart = make(chan bool)
	ticker       *time.Ticker
	enabledHeart = false
	isGossip     = true
)

// Member constructor
func NewMember(introducer bool) *Member {
	mem := &Member{
		0,
		introducer,
		make(map[uint8]membershipListEntry),
	}
	return mem
}

// membershipListEntry constructor
func NewMembershipListEntry(memberID uint8, address net.IP) membershipListEntry {
	mlEntry := membershipListEntry{
		memberID,
		address,
		0,
		time.Now(),
		Alive,
	}
	return mlEntry
}

// Getter for member entry
func (mem *Member) GetMemberEntry(key uint8) membershipListEntry {
	return mem.membershipList[key]
}

// Setter for member entry
func (mem *Member) SetMemberEntry(key uint8, val membershipListEntry) {
	mem.membershipList[key] = val
	return
}

// Remove fn for member entry
func (mem *Member) RemoveMemberEntry(key uint8) {
	delete(mem.membershipList, key)
}

// Get all members in membership list
func (mem *Member) GetAllMembers() map[uint8]membershipListEntry {
	return mem.membershipList
}

func setTicker() {
	ticker = time.NewTicker(time.Duration(Configuration.Settings.gossipInterval) * 1000 * time.Millisecond)
}

// Timer to schedule heartbeats
func (mem *Member) Tick() {
	if ticker == nil {
		setTicker()
	}

	if enabledHeart {
		Warn.Println("Heartbeating has already started.")
		return
	}

	enabledHeart = true
	for {
		select {
		case <-disableHeart:
			enabledHeart = false
			Warn.Println("Stopped heartbeating.")
			return
		case t := <-ticker.C:
			if isGossip {
				mem.Gossip()
			} else {
				Info.Println("All-to-all ", t)
				//mem.AllToAll()
			}
		}
	}
}

// Switch heartbeating modes (All to All or Gossip)
func SetHeartbeating(flag bool) {
	if ticker == nil {
		setTicker()
	}

	isGossip = flag
	interval := time.Millisecond
	if isGossip {
		Info.Println("Running Gossip at T =", Configuration.Settings.gossipInterval)
		interval = time.Duration(Configuration.Settings.gossipInterval) * 1000 * interval
	} else {
		Info.Println("Running All-to-All at T =", Configuration.Settings.allInterval)
		interval = time.Duration(Configuration.Settings.allInterval) * 1000 * interval
	}

	ticker.Reset(interval)
}

func (mem *Member) Gossip() {
	// TODO: failure detection
	// FailMember()
	// CleanupMember()
	// Select random member
	addr := mem.RandIP()
	Info.Println("Gossiping to " + addr.String())

	// Encode the membership list to send it
	b := new(bytes.Buffer)
	e := gob.NewEncoder(b)
	err := e.Encode(mem.membershipList)
	if err != nil {
		panic(err)
	}

	Send(addr.String()+":"+fmt.Sprint(Configuration.Service.port), HeartbeatMsg, b.Bytes())
}

// Listen function to keep listening for messages
func (mem *Member) Listen(port string) {
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
		n, senderAddr, err := listener.ReadFromUDP(buffer)
		if err != nil {
			return
		}

		msgType := buffer[0]
		switch msgType {
		case TextMsg:
			fmt.Println(string(buffer[1:n]))
		case JoinMsg: // only introducer can accept join messages
			if mem.isIntroducer == true {
				Info.Println(senderAddr.String() + " requests to join.")
				mem.acceptMember(senderAddr.IP)
			}
		case HeartbeatMsg: // handles receipt of heartbeat
			Info.Println("Recieved heartbeat from ", senderAddr.String())
		case AcceptMsg: // handles receipt of membership list from introducer
			Info.Println("Introducer has accepted join request.")
			mem.joinResponse(buffer[1:n])
		default:
			Warn.Println("Invalid message type")
		}
	}
}

// request introducer to join
func (mem *Member) joinRequest() {
	Send(Configuration.Service.introducerIP+":"+fmt.Sprint(Configuration.Service.port), JoinMsg, nil)
}

// receive membership list from introducer and setup
func (mem *Member) joinResponse(membershipListBytes []byte) {
	// First byte received corresponds to assigned memberID
	mem.memberID = uint8(membershipListBytes[0])

	// Decode the rest of the buffer to the membership list
	b := bytes.NewBuffer(membershipListBytes[1:])
	d := gob.NewDecoder(b)
	err := d.Decode(&mem.membershipList)
	if err != nil {
		panic(err)
	}

	Info.Println(mem.membershipList)
}

// modify membership list entry
func (mem *Member) leave() {
	newEntry := mem.membershipList[mem.memberID]
	newEntry.HeartbeatCount++
	newEntry.Health = Left
	newEntry.Timestamp = time.Now()
	mem.membershipList[mem.memberID] = newEntry

	// TODO: kill the leaving process after a certain time/# of heartbeats
}

// for introducer to accept a new member
func (mem *Member) acceptMember(address net.IP) {
	// assign new ID
	newMemberID := GetMaxKey(mem.membershipList) + 1
	mem.membershipList[newMemberID] = NewMembershipListEntry(newMemberID, address)

	// Encode the membership list to send it
	b := new(bytes.Buffer)
	e := gob.NewEncoder(b)
	err := e.Encode(mem.membershipList)
	if err != nil {
		panic(err)
	}

	// Send the memberID by appending it to start of buffer, and the membershiplist
	Send(address.String()+":"+fmt.Sprint(Configuration.Service.port), AcceptMsg, append([]byte{newMemberID}, b.Bytes()...))
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

// Find random IP in membership list
func (mem *Member) RandIP() net.IP {
	i := 0
	randVal := rand.Intn(len(mem.membershipList))
	for _, v := range mem.membershipList {
		if i == randVal {
			return v.IPaddr
		}
		i += 1
	}
	return nil
}
