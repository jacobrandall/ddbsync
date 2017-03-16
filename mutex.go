// Copyright 2012 Ryan Smith. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package ddbsync provides DynamoDB-backed synchronization primitives such
// as mutual exclusion locks. This package is designed to behave like pkg/sync.

package ddbsync

import (
	"log"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

// A Mutex is a mutual exclusion lock.
// Mutexes can be created as part of other structures.
type Mutex struct {
	Name              string
	TTL               int64
	LockReattemptWait time.Duration
	db                DBer
}

var _ sync.Locker = (*Mutex)(nil) // Forces compile time checking of the interface

// Mutex constructor
func NewMutex(name string, ttl int64, db DBer, lockReattemptWait time.Duration) *Mutex {
	return &Mutex{
		Name:              name,
		TTL:               ttl,
		db:                db,
		LockReattemptWait: lockReattemptWait,
	}
}

// Lock will write an item in a DynamoDB table if the item does not exist.
// Before writing the lock, we will clear any locks that are expired.
// Calling this function will block until a lock can be acquired.
func (m *Mutex) Lock() {
	for {
		m.PruneExpired()
		err := m.db.Put(m.Name, time.Now().Unix())
		if err == nil {
			return
		}

		// Log the error unless it's related to the mutex already being held
		awsErr, ok := err.(awserr.Error)
		if !ok || awsErr.Code() != dynamodb.ErrCodeConditionalCheckFailedException {
			log.Printf("Lock. Error: %v", err)
		}

		time.Sleep(m.LockReattemptWait)
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

// PruneExpired delete all locks that have lived past their TTL.
// This is to prevent deadlock from processes that have taken locks
// but never removed them after execution. This commonly happens when a
// processor experiences network failure.
func (m *Mutex) PruneExpired() {
	item, err := m.db.Get(m.Name)
	if err != nil {
		if strings.Contains(err.Error(), "No item for Name") {
			// This is normal behaviour, but DynomoDB reports no item as an error
			return
		}
		log.Printf("PruneExpired. error = %v", err)
		return
	}
	if item != nil {
		if item.Created < (time.Now().Unix() - m.TTL) {
			m.Unlock()
		}
	}
	return
}
