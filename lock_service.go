package ddbsync

import (
	"time"
)

// ErrLocker is version of sync.Locker but returns an error
type ErrLocker interface {
	Lock() error
	Unlock()
}

type LockServicer interface {
	NewLock(string, time.Duration, time.Duration, time.Duration) ErrLocker
}

type LockService struct {
	db DBer
}

var _ LockServicer = (*LockService)(nil) // Forces compile time checking of the interface

func NewLockService(tableName string, region string, endpoint string, disableSSL bool) *LockService {
	return &LockService{
		db: NewDatabase(tableName, region, endpoint, disableSSL),
	}
}

// Create a new Lock/Mutex with a particular key and timeout
func (l *LockService) NewLock(name string, ttl, reattemptWait, cutoff time.Duration) ErrLocker {
	mutex := NewMutex(name, ttl, l.db)
	mutex.ReattemptWait = reattemptWait
	mutex.Cutoff = cutoff
	return mutex
}
