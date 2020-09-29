# Distributed Group Membership/Failure Detector

### Overview

We implement two classical protocols for distributed group membership and failure detection in Go. Gossip is an infection-style dissemination protocol that propagates list information to k random nodes at a constant rate. All-to-all failure detection periodically sends heartbeat to all other members and declares faulty nodes if a number of consecutive heartbeats are missed.

**Features**

* Gossip-style heartbeating
* All-to-all heartbeating
* Introducer mechanism
  * Memberes can join, voluntarily leave, or crash out of a group
* Command line interface

**Performance requirements**

* Failures are reflected in one membership list within **< 5 s**
* Update(failure/join/leave) must be reflected in all membership lists within **< 6 s**
* Back-to-back failures < 20s
* Bandwidth-efficient

### Usage

To compile and run:

```
go build -o main
./main
```

Configure introducer, port, and intervals in config.json

```json
{
    "service": {
        "failure_detector": "alltoall",
        "introducer_ip": "172.22.156.42",
        "port": 9090
    },
    "settings": {
        "gossip_interval": 1,
        "all_interval": 3,
        "fail_timeout": 5,
        "cleanup_timeout": 24,
        "num_processes_to_gossip": 2
    }
}
```

CLI

```
join					- join the group and start heartbeating
join introducer			- create a group as the introducer
leave					- leave the group and stop heartbeating
status					- get live status of group
whoami					- get self id
get logs				- pull the current local log
grep 					- grep for logs from other group members
stop					- manually stop heartbeating
kill					- exit gracefully
switch <protocol>		- switch system protocol between options of: gossip, alltoall
metrics					- get current failure/bandwidth stats
sim <test>				- debug only; simulation between options of: failtest
```



## Design

### Components

* Gossip service
* All to all service
* Monitor service
* Main process

Directory structure

```
root
	config.json
	src
		main.go			// main program
		net.go			// networking library
		util.go			// utilities (config, etc)
		detector.go		// gossip/all-to-all handlers
		monitor.go
		logs.go
```

### Gossip/All To All protocol

Gossip disseminates by piggybacking his entire list on heartbeat messages. Heartbeats are disseminated at a rate of k peers/Tgossip seconds. Peers are selected at random at each round. Upon delivery of the heartbeat message, the receiving node will take the following steps:

1. Decode the payload of the incoming packet into a readable list.
2. Read each entry in the list individually.
3. Determine if the entry should be merged into the consumer's list by comparing heartbeat counters.
4. If the heartbeat counter of the entry from the received list is greater than the entry the consumer current holds, update the entry's timestamp and heartbeat counter.
5. Invoke a goroutine to sleep for a Tfail period. The routine will be parked until the sleeping period ends, in which it will check if the time the last heartbeat was received surpasses a Tfail threshold.

All to all follows the same membership list handling approach. However, it will disseminate a heartbeat message to the entire group at every round.

```
func (mem *Member) FailMember(memberId uint8, oldTime time.Time)
func (mem *Member) CleanupMember(memberId uint8, oldTime time.Time)
func (mem *Member) HeartbeatHandler(membershipListBytes []byte)

func (mem *Member) Tick()
func (mem *Member) StopTick()
func SetHeartbeating(flag bool)

func (mem *Member) Gossip()
func (mem *Member) AllToAll()
func (mem *Member) SendAll(msgType MessageType, msg []byte)
```

### Membership

```
const (
    Alive = iota
    Failed
    Left
)

type Member struct {
	memberID       uint8
	isIntroducer   bool
	membershipList map[uint8]membershipListEntry // {uint8 (member_id): Member}
}
// lock individual member or entire list? Or channels?

type membershipListEntry struct {
	MemberID       uint8
	IPaddr         net.IP
	HeartbeatCount uint64
	Timestamp      time.Time
	Health         uint8 // -> Health enum
}

func NewMember(introducer bool) *Member
func NewMembershipListEntry(memberID uint8, address net.IP) membershipListEntry
func (mem *Member) PrintMembershipList(output io.Writer)
func (mem *Member) Listen(port string)
```

### Introducer mechanism

A node can send message to introducer (preconfigured IP)  with desire to join. Introducer responds by providing the new member with an existing membership list. Nodes can also voluntarily leave by setting his health to "Left" and sending one final gossip to a random node to announce his departure.

If introducer suffers from a failure or crash, nodes will not be able to join the group. Nodes can attempt to send join/remove messages but their packets will drop. If a node voluntarily left or crashed, he may rejoin as a user occupying a new id. The node will rejoin as if he was joining the group for the first time - by making the proper request to the introducer. 

Introducers may rejoin. Existing nodes will find the introducer eventually but will somehow need to retain his list entry.

```
func (mem *Member) joinRequest()
func (mem *Member) joinResponse(membershipListBytes []byte)
func (mem *Member) leave()
func (mem *Member) acceptMember(address net.IP)
```

### Wire protocol

We use a simple messaging protocol to facilitate a communication standard between nodes. Every UDP packet is sent and received in the following format:

```
[MessageType byte, payload []byte]
```

The MessageType is an enumerated value declared with the name of the message. Before sending any outgoing packet, the sender will prepend a message type to the payload. At the receiving endpoint, the node will strip the first byte of an incoming packet to determine how the message should be handled.

Currently, we use Gob to encode and decode structs and lists into byte arrays during the marshalling/unmarshalling process. They go into bytes 1 to n in the buffer.

```go
type MessageType uint8
const (
	JoinMsg = iota
	HeartbeatMsg
	TextMsg
	AcceptMsg
	GrepReq
	GrepResp
	SwitchMsg
	TestMsg
)
```

```go
// Send text messages
func SendMessage(addr string, msg string)
// Send messages given addr, messsage type, and payload data (byte array)
func Send(addr string, msgType MessageType, msg []byte) // go routine
func SendAll()
// Opens UDP listener connection over user specified port
func Listener(port string)
```

### Logging

* Use Info.Println("message"), Warn.Println("message"), or Err.Println("message")

### Testing

Ctrl+C or kill lol

1. Kill 1 machine
   1. How to synchronize timestamps to check if 5/6 s requirements are met?
2. Kill introducer/leader node
3. Kill 3 >= machines
4. Packet drops

