package ddbsync

import (
	"errors"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/zencoder/ddbsync/mocks"
	"github.com/zencoder/ddbsync/models"
)

const (
	VALID_MUTEX_NAME       string        = "mut-test"
	VALID_MUTEX_TTL        int64         = 4
	VALID_MUTEX_CREATED    int64         = 1424385592
	VALID_MUTEX_RETRY_WAIT time.Duration = 1 * time.Millisecond
)

var lockHeldErr = awserr.New(
	dynamodb.ErrCodeConditionalCheckFailedException,
	"Conditional Check Failed",
	errors.New("Conditional Check Failed"),
)

var dynamoInternalErr = awserr.New(
	dynamodb.ErrCodeInternalServerError,
	"Dynamo Internal Server Error",
	errors.New("Dynamo Internal Server Error"),
)

func TestNew(t *testing.T) {
	db := new(mocks.DBer)
	underTest := NewMutex(VALID_MUTEX_NAME, VALID_MUTEX_TTL, db, VALID_MUTEX_RETRY_WAIT)

	require.Equal(t, VALID_MUTEX_NAME, underTest.Name)
	require.Equal(t, VALID_MUTEX_TTL, underTest.TTL)
}

func TestLock(t *testing.T) {
	db := new(mocks.DBer)
	underTest := NewMutex(VALID_MUTEX_NAME, VALID_MUTEX_TTL, db, VALID_MUTEX_RETRY_WAIT)

	db.On("Put", VALID_MUTEX_NAME, mock.AnythingOfType("int64")).Return(nil)
	db.On("Get", VALID_MUTEX_NAME).Return(&models.Item{Name: VALID_MUTEX_NAME, Created: VALID_MUTEX_CREATED}, nil)
	db.On("Delete", VALID_MUTEX_NAME).Return(nil)

	underTest.Lock()
	db.AssertExpectations(t)
}

func TestLockWaitsBeforeRetrying(t *testing.T) {
	db := new(mocks.DBer)
	underTest := NewMutex(VALID_MUTEX_NAME, VALID_MUTEX_TTL, db, 300*time.Millisecond)

	db.On("Get", VALID_MUTEX_NAME).Return(&models.Item{Name: VALID_MUTEX_NAME, Created: VALID_MUTEX_CREATED}, nil)
	db.On("Delete", VALID_MUTEX_NAME).Return(nil)
	db.On("Put", VALID_MUTEX_NAME, mock.AnythingOfType("int64")).Once().Return(lockHeldErr)
	db.On("Put", VALID_MUTEX_NAME, mock.AnythingOfType("int64")).Once().Return(dynamoInternalErr)
	db.On("Put", VALID_MUTEX_NAME, mock.AnythingOfType("int64")).Once().Return(errors.New("Dynamo Glitch"))
	db.On("Put", VALID_MUTEX_NAME, mock.AnythingOfType("int64")).Once().Return(nil)

	before := time.Now()
	underTest.Lock()
	duration := time.Since(before)

	db.AssertExpectations(t)
	require.True(t, duration > (900*time.Millisecond), "Expected to have waited at least 0.3 secs between each retry, total wait time: %s", duration)
}

func TestUnlock(t *testing.T) {
	db := new(mocks.DBer)
	underTest := NewMutex(VALID_MUTEX_NAME, VALID_MUTEX_TTL, db, VALID_MUTEX_RETRY_WAIT)

	db.On("Delete", VALID_MUTEX_NAME).Return(nil)

	underTest.Unlock()
	db.AssertExpectations(t)
}

func TestUnlockGivesUpAfterThreeAttempts(t *testing.T) {
	db := new(mocks.DBer)
	underTest := NewMutex(VALID_MUTEX_NAME, VALID_MUTEX_TTL, db, VALID_MUTEX_RETRY_WAIT)

	db.On("Delete", VALID_MUTEX_NAME).Times(3).Return(errors.New("DynamoDB is down!"))

	underTest.Unlock()
	db.AssertExpectations(t)
}

func TestPruneExpired(t *testing.T) {
	db := new(mocks.DBer)
	underTest := NewMutex(VALID_MUTEX_NAME, VALID_MUTEX_TTL, db, VALID_MUTEX_RETRY_WAIT)

	db.On("Get", VALID_MUTEX_NAME).Return(&models.Item{Name: VALID_MUTEX_NAME, Created: VALID_MUTEX_CREATED}, nil)
	db.On("Delete", VALID_MUTEX_NAME).Return(nil)

	underTest.PruneExpired()
	db.AssertExpectations(t)
}

func TestPruneExpiredError(t *testing.T) {
	db := new(mocks.DBer)
	underTest := NewMutex(VALID_MUTEX_NAME, VALID_MUTEX_TTL, db, VALID_MUTEX_RETRY_WAIT)

	db.On("Get", VALID_MUTEX_NAME).Return((*models.Item)(nil), errors.New("Get Error"))

	underTest.PruneExpired()
	db.AssertExpectations(t)
}
