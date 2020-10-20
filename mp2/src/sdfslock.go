package main

import (
	"errors"
	"sync"
	"time"
)

type LockType int

const (
	SdfsRLock LockType = iota
	SdfsLock
)

type SdfsLockRequest struct {
	RemoteFname string
	Type        LockType
}

type SdfsLockResponse struct {
	RemoteFname string
	Type        LockType
}

// Acquire global lock for file
func (node *SdfsNode) AcquireLock(req SdfsLockRequest, reply *SdfsLockResponse) error {
	if node.isMaster == false && node.Master == nil {
		return errors.New("Error: Master not initialized")
	}

	// If lock does not exist, create lock
	if _, ok := node.Master.fileLock[req.RemoteFname]; !ok {
		node.Master.fileLock[req.RemoteFname] = &sync.RWMutex{}
	}

	// Acquire RW lock for remoteFname mutex
	// If read, try to lock
	if req.Type == SdfsRLock {
		// If can lock, respond
		node.Master.fileLock[req.RemoteFname].RLock()
	} else if req.Type == SdfsLock {
		// If write (exclusive), can lock, respond
		node.Master.fileLock[req.RemoteFname].Lock()
	}

	// Lock expiry deadline timer to avoid deadlock
	go func() {
		time.Sleep(60 * time.Second)
		if req.Type == SdfsRLock {
			node.Master.fileLock[req.RemoteFname].RUnlock()
		} else if req.Type == SdfsLock {
			node.Master.fileLock[req.RemoteFname].Unlock()
		}
	}()

	var resp SdfsLockResponse
	resp.RemoteFname = req.RemoteFname
	*reply = resp
	return nil
}

// Release global lock for file
func (node *SdfsNode) ReleaseLock(req SdfsLockRequest, reply *SdfsLockResponse) error {
	if node.isMaster == false && node.Master == nil {
		return errors.New("Error: Master not initialized")
	}

	// If lock does not exist, error
	if _, ok := node.Master.fileLock[req.RemoteFname]; !ok {
		return errors.New("Error: Lock for file does not exist.")
	}

	// If read, try to unlock
	if req.Type == SdfsRLock {
		node.Master.fileLock[req.RemoteFname].RUnlock()
	}

	// If write (exclusive), try to unlock
	if req.Type == SdfsLock {
		node.Master.fileLock[req.RemoteFname].Unlock()
	}
	var resp SdfsLockResponse
	resp.RemoteFname = req.RemoteFname
	*reply = resp
	return nil
}
