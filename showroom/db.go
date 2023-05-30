package main

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

type Event struct {
	URL       string `dynamodbav:"url"`
	Member    string `dynamodbav:"member"` // Partition Key
	Group     string `dynamodbav:"group"`
	RoomID    uint   `dynamodbav:"room_id"`
	StartAt   uint64 `dynamodbav:"start_at"` // Sort Key
	CreatedAt uint64 `dynamodbav:"created_at"`
}

type DB struct {
	tableName string
	client    *dynamodb.Client
}

func Connect(tableName string) *DB {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(awsRegion))
	if err != nil {
		logger.WithError(err).Error("aws sdk initial error")
	}

	// Using the Config value, create the DynamoDB client
	client := dynamodb.NewFromConfig(cfg)
	return &DB{tableName: tableName, client: client}
}

func (db *DB) Insert(event Event) error {
	item, err := attributevalue.MarshalMap(event)
	if err != nil {
		logger.WithError(err).Error("event marshal error")
		return err
	}
	_, err = db.client.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName: aws.String(db.tableName),
		Item:      item,
	})
	if err != nil {
		logger.WithError(err).Error("dynamodb insert error")
		return err
	}
	return nil
}

func (db *DB) FindByStartTime(startTime uint64) ([]Event, error) {
	var events []Event

	filtEx := expression.Name("start_at").GreaterThan(expression.Value(startTime))
	expr, err := expression.NewBuilder().WithFilter(filtEx).Build()
	if err != nil {
		logger.WithError(err).Error("dynamodb expression builder error")
		return nil, err
	}

	response, err := db.client.Scan(context.TODO(), &dynamodb.ScanInput{
		TableName:                 aws.String(db.tableName),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		FilterExpression:          expr.Filter(),
	})
	if err != nil {
		logger.WithError(err).Error("dynamodb scan error")
		return nil, err
	}

	if err = attributevalue.UnmarshalListOfMaps(response.Items, &events); err != nil {
		logger.WithError(err).Error("dynamodb record umarshall error")
		return nil, err
	}
	return events, nil
}
