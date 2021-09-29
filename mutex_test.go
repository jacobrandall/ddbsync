package ddbsync

import (
	"errors"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const (
	VALID_MUTEX_NAME       = "mut-test"
	VALID_MUTEX_TTL        = 4 * time.Second
	VALID_MUTEX_CREATED    = 1424385592
	VALID_MUTEX_RETRY_WAIT = 1 * time.Millisecond
)

var dynamoInternalErr = awserr.New(
	dynamodb.ErrCodeInternalServerError,
	"Dynamo Internal Server Error",
	errors.New("Dynamo Internal Server Error"),
)

type mockDB struct{ mock.Mock }

func (m *mockDB) OnAcquire() *mock.Call {
	return m.On("Acquire", VALID_MUTEX_NAME, VALID_MUTEX_TTL)
}
func (m *mockDB) OnDelete() *mock.Call {
	return m.On("Delete", VALID_MUTEX_NAME)
}
func (m *mockDB) Acquire(name string, ttl time.Duration) error {
	return m.Called(name, ttl).Error(0)
}
func (m *mockDB) Delete(name string) error {
	return m.Called(name).Error(0)
}

func newMockedMutex() (*Mutex, *mockDB) {
	db := &mockDB{}
	mutex := NewMutex(VALID_MUTEX_NAME, VALID_MUTEX_TTL, db)
	mutex.ReattemptWait = VALID_MUTEX_RETRY_WAIT
	return mutex, db
}

func TestNew(t *testing.T) {
	underTest, _ := newMockedMutex()
	require.Equal(t, VALID_MUTEX_NAME, underTest.Name)
	require.Equal(t, VALID_MUTEX_TTL, underTest.TTL)
}

func TestLock(t *testing.T) {
	underTest, db := newMockedMutex()
	defer db.AssertExpectations(t)

	db.OnAcquire().Return(nil)

	require.NoError(t, underTest.Lock())
}

func TestLockWaitsBeforeRetrying(t *testing.T) {
	underTest, db := newMockedMutex()
	defer db.AssertExpectations(t)
	underTest.ReattemptWait = 300 * time.Millisecond

	db.OnAcquire().Once().Return(ErrLocked)
	db.OnAcquire().Once().Return(dynamoInternalErr)
	db.OnAcquire().Once().Return(errors.New("Dynamo Glitch"))
	db.OnAcquire().Once().Return(nil)

	before := time.Now()
	require.NoError(t, underTest.Lock())
	duration := time.Since(before)

	require.True(t, duration > 900*time.Millisecond, "Expected to have waited at least 0.3 secs between each retry, total wait time: %s", duration)
}

func TestLockCutoff(t *testing.T) {
	underTest, db := newMockedMutex()
	defer db.AssertExpectations(t)
	underTest.ReattemptWait = 300 * time.Millisecond
	underTest.Cutoff = 100 * time.Millisecond

	db.OnAcquire().Twice().Return(ErrLocked)

	before := time.Now()
	err := underTest.Lock()
	duration := time.Since(before)

	require.EqualError(t, err, "reached cutoff time")
	require.True(t, duration > 300*time.Millisecond, "Expected to have waited at least 0.3 secs between each retry, total wait time: %s", duration)
}

func TestUnlock(t *testing.T) {
	underTest, db := newMockedMutex()
	defer db.AssertExpectations(t)

	db.OnDelete().Return(nil)

	underTest.Unlock()
}

func TestUnlockGivesUpAfterThreeAttempts(t *testing.T) {
	underTest, db := newMockedMutex()
	defer db.AssertExpectations(t)

	db.OnDelete().Times(3).Return(errors.New("DynamoDB is down!"))

	underTest.Unlock()
}
