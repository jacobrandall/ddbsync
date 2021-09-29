package ddbsync

import (
	"errors"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

var (
	ErrLocked = errors.New("key is locked")
)

type Database struct {
	client    AWSDynamoer
	tableName string
}

func NewDatabase(tableName string, region string, endpoint string, disableSSL bool) *Database {
	return &Database{
		client: dynamodb.New(session.New(&aws.Config{
			Endpoint:   &endpoint,
			Region:     &region,
			DisableSSL: &disableSSL,
		})),
		tableName: tableName,
	}
}

var _ DBer = (*Database)(nil) // Forces compile time checking of the interface

var _ AWSDynamoer = (*dynamodb.DynamoDB)(nil) // Forces compile time checking of the interface

type DBer interface {
	Acquire(string, time.Duration) error
	Delete(string) error
}

type AWSDynamoer interface {
	UpdateItem(*dynamodb.UpdateItemInput) (*dynamodb.UpdateItemOutput, error)
	DeleteItem(*dynamodb.DeleteItemInput) (*dynamodb.DeleteItemOutput, error)
}

func (db *Database) Acquire(name string, ttl time.Duration) error {
	now := time.Now()
	_, err := db.client.UpdateItem(&dynamodb.UpdateItemInput{
		TableName: aws.String(db.tableName),
		Key:       key(name),
		ExpressionAttributeNames: map[string]*string{
			"#N": aws.String("Name"),
			"#C": aws.String("Created"),
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":now":    dynamoTime(now),
			":cutoff": dynamoTime(now.Add(-ttl)),
		},
		ConditionExpression: aws.String(`attribute_not_exists(#N) OR #C < :cutoff`),
		UpdateExpression:    aws.String(`SET #C = :now`),
	})
	var awsErr awserr.Error
	if errors.As(err, &awsErr) && awsErr.Code() == dynamodb.ErrCodeConditionalCheckFailedException {
		return ErrLocked
	}
	return err
}

func (db *Database) Delete(name string) error {
	_, err := db.client.DeleteItem(&dynamodb.DeleteItemInput{
		TableName: aws.String(db.tableName),
		Key:       key(name),
	})
	return err
}

func dynamoTime(t time.Time) *dynamodb.AttributeValue {
	return &dynamodb.AttributeValue{
		N: aws.String(strconv.FormatInt(t.UnixMilli(), 10)),
	}
}

func key(name string) map[string]*dynamodb.AttributeValue {
	return map[string]*dynamodb.AttributeValue{
		"Name": {S: aws.String(name)},
	}
}
