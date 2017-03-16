package ddbsync

import (
	"github.com/stretchr/testify/require"
	"testing"
)

const LOCK_SERVICE_VALID_MUTEX_NAME string = "mut-test"
const LOCK_SERVICE_VALID_MUTEX_TTL int64 = 4

func TestNewLock(t *testing.T) {
	ls := &LockService{}
	m := ls.NewLock(LOCK_SERVICE_VALID_MUTEX_NAME, LOCK_SERVICE_VALID_MUTEX_TTL)

	require.NotNil(t, ls)
	require.NotNil(t, m)
	require.IsType(t, &LockService{}, ls)
	require.IsType(t, &Mutex{}, m)
	require.Equal(t, &Mutex{Name: LOCK_SERVICE_VALID_MUTEX_NAME, TTL: LOCK_SERVICE_VALID_MUTEX_TTL}, m)
}
