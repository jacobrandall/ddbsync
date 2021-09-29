package ddbsync

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNewLock(t *testing.T) {
	const (
		name      = "mut-test"
		ttl       = 4 * time.Second
		retryWait = 5 * time.Second
		cutoff    = 6 * time.Second
	)
	var (
		require = require.New(t)
		ls      = &LockService{}
		m       = ls.NewLock(name, ttl, retryWait, cutoff)
	)
	require.NotNil(m)
	require.IsType(&Mutex{}, m)
	require.Equal(&Mutex{
		Name:          name,
		TTL:           ttl,
		ReattemptWait: retryWait,
		Cutoff:        cutoff,
	}, m)
}
