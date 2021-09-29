package ddbsync

import (
	"errors"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/stretchr/testify/assert"
)

const (
	DB_VALID_TABLE_NAME = "TestLockTable"
	DB_VALID_NAME       = "db-name"
	DB_VALID_TTL        = 30 * time.Second
)

type mockDynamo struct {
	UpdateItemOutput *dynamodb.UpdateItemOutput
	UpdateItemError  error
	DeleteItemOutput *dynamodb.DeleteItemOutput
	DeleteItemError  error
}

func (m mockDynamo) UpdateItem(*dynamodb.UpdateItemInput) (*dynamodb.UpdateItemOutput, error) {
	return m.UpdateItemOutput, m.UpdateItemError
}
func (m mockDynamo) DeleteItem(*dynamodb.DeleteItemInput) (*dynamodb.DeleteItemOutput, error) {
	return m.DeleteItemOutput, m.DeleteItemError
}

func newMockedClient(mocked mockDynamo) *Database {
	return &Database{client: mocked, tableName: DB_VALID_TABLE_NAME}
}

func TestDBPut(t *testing.T) {
	db := newMockedClient(mockDynamo{
		UpdateItemOutput: &dynamodb.UpdateItemOutput{},
	})

	err := db.Acquire(DB_VALID_NAME, DB_VALID_TTL)
	assert.NoError(t, err)
}

func TestDBPutError(t *testing.T) {
	db := newMockedClient(mockDynamo{
		UpdateItemError: errors.New("UpdateItem Error"),
	})

	err := db.Acquire(DB_VALID_NAME, DB_VALID_TTL)
	assert.Error(t, err)
}

func TestDBPutErrorLocked(t *testing.T) {
	db := newMockedClient(mockDynamo{
		UpdateItemError: awserr.New(
			dynamodb.ErrCodeConditionalCheckFailedException,
			"condition check failed",
			errors.New("failed")),
	})

	err := db.Acquire(DB_VALID_NAME, DB_VALID_TTL)
	assert.Error(t, err)
}

func TestDBDelete(t *testing.T) {
	db := newMockedClient(mockDynamo{
		DeleteItemOutput: &dynamodb.DeleteItemOutput{},
	})

	err := db.Delete(DB_VALID_NAME)
	assert.NoError(t, err)
}

func TestDBDeleteError(t *testing.T) {
	db := newMockedClient(mockDynamo{
		DeleteItemError: errors.New("DeleteItem Error"),
	})

	err := db.Delete(DB_VALID_NAME)
	assert.Error(t, err)
}
