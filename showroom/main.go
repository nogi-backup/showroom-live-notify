package main

import (
	"context"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/sirupsen/logrus"
)

var (
	awsRegion         string
	dcWebhookURL      string
	dynamoDBTableName string
	logger            *logrus.Entry
)

func init() {
	l := logrus.New()
	l.SetFormatter(&logrus.JSONFormatter{})
	l.SetOutput(os.Stdout)
	l.SetLevel(logrus.InfoLevel)
	logger = l.WithField("app", "showroom-live-notifiy")

	// DC
	dcWebhookURL = os.Getenv("DISCORD_WEBHOOK_URL")
	if len(dcWebhookURL) == 0 {
		panic("Discord Webhook is missing")
	}

	// DynamoDB
	dynamoDBTableName = os.Getenv("DYNAMODB_TABLE_NAME")
	if len(dynamoDBTableName) == 0 {
		panic("DynamoDB Table Name is missing")
	}
}

func main() {
	lambda.Start(HandleRequest)
}

func HandleRequest(ctx context.Context) error {
	current := time.Now().Unix()

	// Discord
	dc, err := NewDiscord(dcWebhookURL)
	if err != nil {
		panic(err)
	}

	// DynamoDB
	dynamoDB := Connect(dynamoDBTableName)

	// Get Showroom Today Pick
	rooms, err := GetTodayPick()
	if err != nil {
		return err
	}

	logger.WithField("rooms", rooms).Info("N46 rooms")

	// Check Existed Records
	events, err := dynamoDB.FindByStartTime(uint64(current))
	if err != nil {
		return err
	}

	for _, room := range rooms {
		e := room.ParseToEvent()
		if e.Group != "乃木坂46" {
			continue
		}
		if events == nil || len(events) == 0 {
			dynamoDB.Insert(e)
			dc.PostMessage(e)
		} else {
			for _, event := range events {
				if event.Member == e.Member && e.StartAt > event.StartAt {
					dynamoDB.Insert(e)
					dc.PostMessage(e)
				}
			}
		}
	}
	return nil
}
