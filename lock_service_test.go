package ddbsync

import (
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

const LOCK_SERVICE_VALID_MUTEX_NAME string = "mut-test"
const LOCK_SERVICE_VALID_MUTEX_TTL int64 = 4
const LOCK_SERVICE_VALID_MUTEX_RETRY_WAIT time.Duration = 5 * time.Second

func TestNewLock(t *testing.T) {
	ls := &LockService{}
	m := ls.NewLock(LOCK_SERVICE_VALID_MUTEX_NAME, LOCK_SERVICE_VALID_MUTEX_TTL, LOCK_SERVICE_VALID_MUTEX_RETRY_WAIT)

	require.NotNil(t, ls)
	require.NotNil(t, m)
	require.IsType(t, &LockService{}, ls)
	require.IsType(t, &Mutex{}, m)
	require.Equal(t, &Mutex{Name: LOCK_SERVICE_VALID_MUTEX_NAME, TTL: LOCK_SERVICE_VALID_MUTEX_TTL, LockReattemptWait: LOCK_SERVICE_VALID_MUTEX_RETRY_WAIT}, m)
}
