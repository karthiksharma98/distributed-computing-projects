package main

import (
	"errors"
	"sync"
)

type LockType int

const (
	SdfsRLock LockType = iota
	SdfsLock
)

type SdfsLockRequest struct {
        NodeId      int
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
                node.Master.fileCond[req.RemoteFname] = sync.NewCond(node.Master.fileLock[req.RemoteFname])
                Info.Println("AcquireLock: Created new lock")
	}

        Info.Println("AcquireLock: Acquiring lock")

	// Acquire RW lock for remoteFname mutex
	// If read, try to lock
	if req.Type == SdfsRLock {
		// If can lock, respond
		node.Master.fileLock[req.RemoteFname].RLock()
	} else if req.Type == SdfsLock {
		// If write (exclusive), can lock, respond
		node.Master.fileLock[req.RemoteFname].Lock()
	}

        Info.Println("AcquireLock: Creating release channel to check lock")

        // Make channel to check if locks been released
        releaseCh := make(chan bool, 1)

	// Lock expiry deadline timer to avoid deadlock
	go func() {
                // Use conditional variable, wakes up upon receipt of unlock
                select {
                case <-releaseCh:
                        Info.Println("Lock released, stop monitor")
                        return
                case id := <-failCh:
                        if int(id) == req.NodeId {
                                Info.Println("Detected lock owner failure, unlocking: ", id)
                                /*
                                if req.Type == SdfsRLock {
                                        node.Master.fileLock[req.RemoteFname].RUnlock()
                                } else if req.Type == SdfsLock {
                                        node.Master.fileLock[req.RemoteFname].Unlock()
                                }*/
                        }
                }
	}()

        // Checks if lock released
        go func() {
                node.Master.fileCond[req.RemoteFname].Wait()
                releaseCh <- true
                return
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

        Info.Println("Releasing lock")

	// If read, try to unlock
	if req.Type == SdfsRLock {
		node.Master.fileLock[req.RemoteFname].RUnlock()
	}

	// If write (exclusive), try to unlock
	if req.Type == SdfsLock {
		node.Master.fileLock[req.RemoteFname].Unlock()
	}

        Info.Println("Lock released")

        node.Master.fileCond[req.RemoteFname].Signal()

	var resp SdfsLockResponse
	resp.RemoteFname = req.RemoteFname
	*reply = resp
	return nil
}
