## Distributed Group Membership/Failure Detector

### Overview

Gossip:

* Infection-style dissemination
* Propagates list information to k random nodes in member list

All-to-all:

* Every member periodically sends heartbeat to all other members
* Declare M_i failed by non-faulty M_j if:
  * M_j does not receive heartbeats from M_i for n consecutive heartbeats

### Membership list update scenarios

1. Machine joins group.
2. Machine voluntarily leaves group.
3. Machine involuntarily leaves group (failure by crash). - check if heartbeat is received by node < 5s

### Requirements

**Features**

* Gossip-style heartbeating
  * <= 3 machines can fail simultaneously
    * What to do if <= 3 machines fail?
* All-to-all heartbeating
  * Any number of machine can fail simultaneously

**Performance**

* Failure must be reflected in one membership list **< 5 s**
  * Heartbeat from A to B timeouts < 5s, do update list
* Update(failure/join/leave) must be reflected in all membership lists **< 6 s**
  * If failure is detected at ~5s, how to ensure failure is in all membership lists ~6s? Broadcast updated state immediately to all nodes?
* Back-to-back failures < 20s
  * What is meant by a "set of failures"? If one failure happens, is that considered a set of failures and no more failures should occur after 20s?
* Bandwidth-efficient

**Other**

* Introducers/leaders cannot remove/fail nodes from list

### Design

* Components:
  * Gossip service
  * All to all service
  * Why not both??
* Gossip protocol:
  * Disseminate updates of members, if any, instead of entire list (bandwidth-efficient)
    * Piggyback state updates on heartbeat messages [1]
    * Low overhead packet (10B): member_id, counter, enumerated update type
    * Merge on heartbeat counter comparison; choose max
    * Update timestamp
    * Reset health of member if needed
  * Heartbeat counting
    * Upon sending heartbeat (rate consistent interval)?
  * Round-robin heartbeat scheduling
    * Disseminate heartbeats at a rate of  (k neighbors)/(5s)
    * Disseminate when (current time - time of last heartbeat) > 4-5s
  * Distribute new updates to all scheduled heartbeat messages sitting in queue
  * Peer selection: Who to infect? Random? Predetermined set?
    * Traditional gossip: send to one random member uniformly[2]
  * Suspicion mechanism similar to SWIM? [1]
  * Failure
    * Consider member failed
    * Local clock - timestamp > Tfail
* All to all protocol:
  * Heartbeat counting
    * Increment periodically
  * Failure
    * Heartbeats not recieved after since
      * Local clock - timestamp > Tfail
      * N heartbeats later
* Introducer mechanism
  * Node sends message to introducer (preconfigured IP)  with desire to join
    * Introducer provides new member with existing membership list
  * Node sends message to introducer to voluntarily leave
    * Introducer gossips randomly or multicasts removal of member to all?
  * Introducer fails
    * Nodes try to send join/remove messages but packets will drop
  * Introducer rejoins
    * Existing nodes will find the introducer eventually but how will the introducer receive the current list?
    * Pull entire list at start up from node
  * Node rejoins
    * Different id
    * Must make request to introducer for list

```go
enum Health {
    ALIVE,
    FAIL,
    LEAVE
}

type Member struct {
    member_id uint8
    ip_addr IP // 'net' package
    heartbeat_count uint64
    timestamp uint64
    health uint8 // -> Health enum
}

// hashtable holding record of all members
var membership_list map[uint8]Member // {uint8 (member_id): Member}
// lock individual member or entire list? Or channels?
```
API

```go
// basic membershiplist API (goroutines?)
func GetMember(member_id uint8) -> (Member)
func CreateOrUpdateMember(member_id uint8, member_data Member)
func RemoveMember(member_id uint8) -> (Member)
func GetAllMembers() -> ([]Member)

// networking
func SendMessage(member_id uint8, message GossipMessage) //goroutine
func SendAll()

// choose random node, check list for failures, send new list to random node
func Gossip()

// heartbeat event handler invoked when connection is recieved
// read inc. message, send ack, perform necessary updates
func HeartbeatHandler() 

// health update mechanism
func FailMember(member_id uint8)
func CleanupMembers()

// introducer mechanism
func Join() -> membership_list map[uint8]Member
func Leave()
func AcceptMember(IPv4)

func Log(message string)
func GetStatus(message string)

// read args commands etc
func main()
```

Configuration (json or yaml?)

```json
{
    "service": {
		"failure_detector": "gossip",
    	"introducer_ip": "10.0.0.1",
    	"port": 1001
    },
    "settings": {
        "gossip_interval": 0.5,
        "fail_timeout": 5,
        "cleanup_timeout": 24
    }
}
```

Directory structure

```
Root
	Src
		main.go
		detector.go
		Config.json
	Tests
		tests.go

```



### Testing

Ctrl+C or kill lol

1. Kill 1 machine
   1. How to synchronize timestamps to check if 5/6 s requirements are met?
2. Kill introducer/leader node
3. Kill 3 >= machines
4. Packet drops



Report:

* Experimental comparison (gossip vs all-to-all)
* Completeness: crash-failure of any group member can be detected by all non-faulty members
* Speed of failure detection: the time interval between a member failure and its detection by some non-fault group member
* Bandwidth: B/sec diseminated packets
* Accuracy: rate of false-positive failures (no report for failures for )



References

[1] SWIM paper

[2] https://dl.acm.org/doi/pdf/10.5555/1659232.1659238

[3] https://www.cs.cornell.edu/home/rvr/papers/GossipFD.pdf

[4] https://research.cs.cornell.edu/projects/Quicksilver/public_pdfs/2007PromiseAndLimitations.pdf

[5] https://stackoverflow.com/questions/31121906/how-to-guarantee-that-all-nodes-get-infected-in-gossip-based-protocols

[6] https://github.com/golang-standards/project-layout