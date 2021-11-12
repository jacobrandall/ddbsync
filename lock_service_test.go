package ddbsync

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNewLock(t *testing.T) {
	t.Parallel()
	const (
		name      = "mut-test"
		ttl       = 4 * time.Second
		retryWait = 5 * time.Second
		cutoff    = 6 * time.Second
	)
	var (
		ls = &LockService{}
		m  = ls.NewLock(name, ttl, retryWait, cutoff)
	)
	require.NotNil(t, m)
	require.IsType(t, &Mutex{}, m)
	require.Equal(t, &Mutex{
		Name:          name,
		TTL:           ttl,
		ReattemptWait: retryWait,
		Cutoff:        cutoff,
	}, m)
}
