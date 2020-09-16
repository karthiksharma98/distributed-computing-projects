package main

import (
	"time"
)

type Member struct {
	memberID       uint8
	addr           string // TODO: should probably store an IP struct instead of the string
	heartbeatCount uint64
	timestamp      int64
	health         uint8
}

// Health status
const (
	Alive = iota
	Failed
	Left
)

var (
	memberList map[uint8]Member
	IDcount    uint8
)

func Initialize() {
	memberList = make(map[uint8]Member)
	IDcount = 0
	// TODO: temporary... mock for now
	NewMember("10.0.0.1")
	NewMember("10.0.0.2")
	NewMember("10.0.0.3")
}

func NewMember(address string) Member {
	// TODO: might merge this into a CreateOrUpdate func
	IDcount += 1
	memberList[IDcount] = Member{
		memberID:       IDcount,
		addr:           address,
		heartbeatCount: 0,
		timestamp:      time.Now().UnixNano(),
		health:         Alive,
	}

	return memberList[IDcount]
}

func GetMember(id uint8) Member {
	return memberList[id]
}

func GetAllMembers() map[uint8]Member {
	return memberList
}
