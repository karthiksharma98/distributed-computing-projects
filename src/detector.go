package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"math/rand"
	"net"
	"strings"
	"text/tabwriter"
	"time"
)

// Member struct to hold member info
// TODO: change membershipList to store pointers of membershipListEntry to make updating entries cleaner
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
        joinAck = make(chan bool)
	disableHeart = make(chan bool)
	ticker       *time.Ticker
	enabledHeart = false
	isGossip     = true
        listener *net.UDPConn
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

// PrintMembershipList pretty-prints all values inside the membership list
func (mem *Member) PrintMembershipList(output io.Writer) {
	writer := tabwriter.NewWriter(output, 0, 8, 1, '\t', tabwriter.AlignRight)
	fmt.Fprintln(writer, "MemberID\tIP\tHeartbeats\tTimestamp\tHealth")
	fmt.Fprintln(writer, "-------\t----------\t----\t-----------\t------")
	for _, v := range mem.membershipList {
		fmt.Fprintf(writer, "%v\t%v\t%v\t%v\t%v\n", v.MemberID, v.IPaddr, v.HeartbeatCount, v.Timestamp.String(), v.Health)
	}
	writer.Flush()
}

// Verifies whether a node has been updated in T_Failed seconds.
func (mem *Member) FailMember(memberId uint8, oldTime time.Time) {
	if memberId == mem.memberID {
		return
	}

	if currEntry, ok := mem.membershipList[memberId]; ok {
		difference := currEntry.Timestamp.Sub(oldTime)
		threshold := time.Duration(Configuration.Settings.failTimeout) * time.Second
		if difference <= threshold && currEntry.Health == Alive {
			mem.membershipList[memberId] = membershipListEntry{
				currEntry.MemberID,
				currEntry.IPaddr,
				currEntry.HeartbeatCount,
				currEntry.Timestamp,
				Failed,
			}

			Info.Println("Marked member failed: ", memberId)
		}
	}

}

// Verifies whether a node has been updated in T_Cleanup seconds.
func (mem *Member) CleanupMember(memberId uint8, oldTime time.Time) {
	if memberId == mem.memberID {
		return
	}

	if currEntry, ok := mem.membershipList[memberId]; ok {
		difference := currEntry.Timestamp.Sub(oldTime)
		threshold := time.Duration(Configuration.Settings.cleanupTimeout) * time.Second
		if difference <= threshold {
			delete(mem.membershipList, memberId)
			Info.Println("Cleaned up member: ", memberId)
		}
	}
}

// Invoked when hearbeat is recieved
func (mem *Member) HeartbeatHandler(membershipListBytes []byte) {
	// grab membership list in order to merge with your own
	// decode the buffer to the membership list, similar to joinResponse()
	b := bytes.NewBuffer(membershipListBytes)
	d := gob.NewDecoder(b)
	rcvdMemList := make(map[uint8]membershipListEntry)

	err := d.Decode(&rcvdMemList)
	if err != nil {
		panic(err)
	}

	for id, rcvdEntry := range rcvdMemList {
		// Dont let anybody else tell u ur a failure
		if id == mem.memberID {
			continue
		}

		newHealth := uint8(Alive)
		// Update if member voluntarily left
		if rcvdEntry.Health == Left {
			newHealth = Left
		}

		oldTime := rcvdMemList[id].Timestamp
		newTime := time.Now()
		newHeartbeatCt := rcvdEntry.HeartbeatCount

		// check that they have the same id in their membership list
		if currEntry, ok := mem.membershipList[id]; ok {
			// No changes to timestamp/heartbeat count if count has not been updated
			if rcvdEntry.HeartbeatCount <= currEntry.HeartbeatCount {
				newHeartbeatCt = currEntry.HeartbeatCount
				newTime = currEntry.Timestamp
			}

			if oldTime.Before(currEntry.Timestamp) {
				oldTime = currEntry.Timestamp
			}
		}

		mem.membershipList[id] = membershipListEntry{
			rcvdEntry.MemberID,
			rcvdEntry.IPaddr,
			newHeartbeatCt,
			newTime,
			newHealth,
		}

		// Cmp most recently updated entry timestamp
		time.AfterFunc(
			time.Duration(Configuration.Settings.failTimeout)*time.Second,
			func() {
				mem.FailMember(id, oldTime)
			})

		time.AfterFunc(
			time.Duration(Configuration.Settings.cleanupTimeout)*time.Second,
			func() {
				mem.CleanupMember(id, oldTime)
			})
	}
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
		// Listen channel to disable heartbeating
		select {
		case <-disableHeart:
			enabledHeart = false
			Warn.Println("Stopped heartbeating.")
			return
		case _ = <-ticker.C:
			// Increment heartbeat counter of self
			entry := mem.membershipList[mem.memberID]
			entry.HeartbeatCount += 1
			mem.membershipList[mem.memberID] = entry
			// Gossip or AllToAll
			if isGossip {
				mem.Gossip()
                                mem.Gossip()
			} else {
				mem.AllToAll()
			}
		}
	}
}

// Stop ticking if enabled
func (mem *Member) StopTick() {
	if !enabledHeart {
		Warn.Println("No process running to stop.")
		return
	}

	disableHeart <- true
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

// Send Gossip to random member in group
func (mem *Member) Gossip() {
	// Select random member
	addr := mem.PickRandMemberIP()
	if addr == nil {
		Info.Println("No other member to gossip to!")
		return
	}

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

// AllToAll heartbeating - sends membership list to all other
func (mem *Member) AllToAll() {
	// Encode the membership list to send it
	b := new(bytes.Buffer)
	e := gob.NewEncoder(b)
	err := e.Encode(mem.membershipList)
	if err != nil {
		panic(err)
	}
	Info.Println("Sending All-to-All.")
	// Send heartbeatmsg and membership list to all members
	mem.SendAll(HeartbeatMsg, b.Bytes())
}

// Broadcast a message to all members in membershiplist
func (mem *Member) SendAll(msgType MessageType, msg []byte) {
	for _, v := range mem.membershipList {
		Send(v.IPaddr.String()+":"+fmt.Sprint(Configuration.Service.port), msgType, msg)
	}
}

// Listen function to keep listening for messages
func (mem *Member) Listen(port string) {
	// UDP buffer 1024 bytes for now
	buffer := make([]byte, 1024)
	addr, err := net.ResolveUDPAddr("udp", ":"+port)
	if err != nil {
		panic(err)
	}

	listener, err = net.ListenUDP("udp", addr)
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
			mem.HeartbeatHandler(buffer[1:n])
			Info.Println("Recieved heartbeat from ", senderAddr.String())
		case AcceptMsg: // handles receipt of membership list from introducer
			Info.Println("Introducer has accepted join request.")
			mem.joinResponse(buffer[1:n])
		case GrepReq: // handles grep request
			ipAddr := senderAddr.String()[:strings.IndexByte(senderAddr.String(), ':')]
			mem.HandleGrepRequest(ipAddr, buffer[1:n])
		case GrepResp: // handles grep response when one is received
			mem.HandleGrepResponse(buffer[1:n])
		case SwitchMsg:
			if buffer[1] == 1 {
				SetHeartbeating(true)
			} else {
				SetHeartbeating(false)
			}
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
        joinAck <- true

	Info.Println(mem.membershipList)
}

// modify membership list entry
func (mem *Member) leave() {
	newEntry := mem.membershipList[mem.memberID]
	newEntry.HeartbeatCount++
	newEntry.Health = Left
	newEntry.Timestamp = time.Now()
	mem.membershipList[mem.memberID] = newEntry
	// Gossip leave status and stop
	mem.Gossip()
	mem.StopTick()
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
func (mem *Member) PickRandMemberIP() net.IP {
	if len(mem.membershipList) == 1 {
		// you are the only process in the list
		return nil
	}

	// loop until you find a member that isn't your own
	for {
		i := 0
		randVal := rand.Intn(len(mem.membershipList))
		var randEntry membershipListEntry
		for _, v := range mem.membershipList {
			if i == randVal {
				randEntry = v
			}

			i += 1
		}

		if randEntry.MemberID != mem.memberID {
			return randEntry.IPaddr
		}
	}

	return nil
}
