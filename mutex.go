// Copyright 2012 Ryan Smith. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package ddbsync provides DynamoDB-backed synchronization primitives such
// as mutual exclusion locks. This package is designed to behave like pkg/sync.

package ddbsync

import (
	"errors"
	"log"
	"time"
)

const (
	DefaultReattemptWait = time.Second
	DefaultCutoff        = 2 * time.Minute
)

var (
	ErrReachedCutoff = errors.New("reached cutoff time")
	// Forces compile time checking of the interface
	_ ErrLocker = (*Mutex)(nil)
)

// A Mutex is a mutual exclusion lock.
// Mutexes can be created as part of other structures.
type Mutex struct {
	Name          string
	TTL           time.Duration
	ReattemptWait time.Duration
	Cutoff        time.Duration
	db            DBer
}

// Mutex constructor
func NewMutex(name string, ttl time.Duration, db DBer) *Mutex {
	return &Mutex{
		Name: name,
		TTL:  ttl,
		db:   db,
	}
}

// Lock will write an item in a DynamoDB table if the item does not exist.
// Before writing the lock, we will clear any locks that are expired.
// Calling this function will block until a lock can be acquired.
func (m *Mutex) Lock() error {
	if m.ReattemptWait <= 0 {
		m.ReattemptWait = DefaultReattemptWait
	}
	cutoffTime := time.Now().Add(m.Cutoff)
	for {
		err := m.db.Acquire(m.Name, m.TTL)
		// Early return when we have the lock
		if err == nil {
			return nil
		}
		// Log the error unless it's related to the mutex already being held
		if !errors.Is(err, ErrLocked) {
			log.Printf("Lock. Error: %v", err)
		}
		// Check cutoff
		if m.Cutoff > 0 && time.Now().After(cutoffTime) {
			return ErrReachedCutoff
		}
		// Sleep before retrying
		time.Sleep(m.ReattemptWait)
	}
}

// Unlock will delete an item in a DynamoDB table.
// If for some reason we can't (Dynamo is down / TTL of lock expired and something else deleted it) then
// we give up after a few attempts and let the TTL catch it (if it hasn't already).
func (m *Mutex) Unlock() {
	for i := 0; i < 3; i++ {
		err := m.db.Delete(m.Name)
		if err == nil {
			return
		}
		log.Printf("Unlock. Error: %v", err)
	}
}
