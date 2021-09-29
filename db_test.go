package ddbsync

import (
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/zencoder/ddbsync/mocks"
	"github.com/zencoder/ddbsync/models"
)

const (
	DB_VALID_TABLE_NAME     string = "TestLockTable"
	DB_VALID_NAME           string = "db-name"
	DB_VALID_CREATED        int64  = 1424385592
	DB_VALID_CREATED_STRING string = "1424385592"
)

func newMockedClient() (*database, *mocks.AWSDynamoer) {
	mocked := &mocks.AWSDynamoer{}
	db := &database{client: mocked, tableName: DB_VALID_TABLE_NAME}
	return db, mocked
}

func TestDBPut(t *testing.T) {
	db, client := newMockedClient()
	defer client.AssertExpectations(t)

	client.On("PutItem", mock.AnythingOfType("*dynamodb.PutItemInput")).Return(&dynamodb.PutItemOutput{}, nil)
	err := db.Put(DB_VALID_NAME, DB_VALID_CREATED)
	assert.NoError(t, err)
}

func TestDBPutError(t *testing.T) {
	db, client := newMockedClient()
	defer client.AssertExpectations(t)

	client.On("PutItem", mock.AnythingOfType("*dynamodb.PutItemInput")).Return((*dynamodb.PutItemOutput)(nil), errors.New("PutItem Error"))
	err := db.Put(DB_VALID_NAME, DB_VALID_CREATED)
	assert.Error(t, err)
}

func TestDBGet(t *testing.T) {
	db, client := newMockedClient()
	defer client.AssertExpectations(t)

	one := int64(1)
	qo := &dynamodb.QueryOutput{
		Count: &one,
		Items: []map[string]*dynamodb.AttributeValue{
			{
				"Name": &dynamodb.AttributeValue{
					S: aws.String(DB_VALID_NAME),
				},
				"Created": &dynamodb.AttributeValue{
					N: aws.String(DB_VALID_CREATED_STRING),
				},
			},
		},
	}

	client.On("Query", mock.AnythingOfType("*dynamodb.QueryInput")).Return(qo, nil)

	i, err := db.Get(DB_VALID_NAME)

	assert.NotNil(t, i)
	assert.NoError(t, err)
	assert.Equal(t, &models.Item{Name: DB_VALID_NAME, Created: DB_VALID_CREATED}, i)
}

func TestDBGetErrorNoQueryOutput(t *testing.T) {
	db, client := newMockedClient()
	defer client.AssertExpectations(t)

	client.On("Query", mock.AnythingOfType("*dynamodb.QueryInput")).Return((*dynamodb.QueryOutput)(nil), errors.New("Query Error"))

	i, err := db.Get(DB_VALID_NAME)

	assert.Nil(t, i)
	assert.Error(t, err)
}

func TestDBGetErrorNilCount(t *testing.T) {
	db, client := newMockedClient()
	defer client.AssertExpectations(t)

	qo := &dynamodb.QueryOutput{
		Count: nil,
	}

	client.On("Query", mock.AnythingOfType("*dynamodb.QueryInput")).Return(qo, nil)

	i, err := db.Get(DB_VALID_NAME)

	assert.Nil(t, i)
	assert.EqualError(t, err, "Count not returned")
}

func TestDBGetErrorZeroCount(t *testing.T) {
	db, client := newMockedClient()
	defer client.AssertExpectations(t)

	zero := int64(0)
	qo := &dynamodb.QueryOutput{
		Count: &zero,
	}

	client.On("Query", mock.AnythingOfType("*dynamodb.QueryInput")).Return(qo, nil)

	i, err := db.Get(DB_VALID_NAME)

	assert.Nil(t, i)
	assert.EqualError(t, err, fmt.Sprintf("No item for Name, %s", DB_VALID_NAME))
}

func TestDBGetErrorCountTooHigh(t *testing.T) {
	db, client := newMockedClient()
	defer client.AssertExpectations(t)

	two := int64(2)
	qo := &dynamodb.QueryOutput{
		Count: &two,
	}

	client.On("Query", mock.AnythingOfType("*dynamodb.QueryInput")).Return(qo, nil)

	i, err := db.Get(DB_VALID_NAME)

	assert.Nil(t, i)
	assert.EqualError(t, err, "Expected only 1 item returned from Dynamo, got 2")
}

func TestDBGetErrorCountSetNoItems(t *testing.T) {
	db, client := newMockedClient()
	defer client.AssertExpectations(t)

	one := int64(1)
	qo := &dynamodb.QueryOutput{
		Count: &one,
	}

	client.On("Query", mock.AnythingOfType("*dynamodb.QueryInput")).Return(qo, nil)

	i, err := db.Get(DB_VALID_NAME)

	assert.Nil(t, i)
	assert.EqualError(t, err, "No item returned, count is invalid.")
}

func TestDBDelete(t *testing.T) {
	db, client := newMockedClient()
	defer client.AssertExpectations(t)

	client.On("DeleteItem", mock.AnythingOfType("*dynamodb.DeleteItemInput")).Return(&dynamodb.DeleteItemOutput{}, nil)

	err := db.Delete(DB_VALID_NAME)

	assert.NoError(t, err)
}

func TestDBDeleteError(t *testing.T) {
	db, client := newMockedClient()
	defer client.AssertExpectations(t)

	client.On("DeleteItem", mock.AnythingOfType("*dynamodb.DeleteItemInput")).Return((*dynamodb.DeleteItemOutput)(nil), errors.New("Delete Error"))

	err := db.Delete(DB_VALID_NAME)

	assert.Error(t, err)
}
