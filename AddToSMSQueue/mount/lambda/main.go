package main

import (
	"fmt"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"strconv"
	"time"
)


type SMSMessage struct {
	ToNumber   string `json:"To"`
	FromNumber string `json:"From"`
	Body 	   string `json:"Body"`
	Broadcast  string `json:"Broadcast"`
}

var (
	sess, err = session.NewSession(&aws.Config{
		Region: aws.String("us-east-1")},
	)

	svc = sqs.New(sess)

	// URL to our queue
	qURL = "https://sqs.us-east-1.amazonaws.com/*************/SMSQueue.fifo"
)

func LambdaStart(sms SMSMessage) (SMSMessage, error) {

	// Create a unique key for message deduplication
	dedupeID := sms.ToNumber + ":" + sms.FromNumber + ":" + strconv.Itoa(int(time.Now().Unix()))

	// Send to SQS Queue
	result, err := svc.SendMessage(&sqs.SendMessageInput{
		MessageAttributes: map[string]*sqs.MessageAttributeValue{
			"ToNumber": {
				DataType:    aws.String("String"),
				StringValue: aws.String(sms.ToNumber),
			},
			"FromNumber": {
				DataType:    aws.String("String"),
				StringValue: aws.String(sms.FromNumber),
			},
			"Body": {
				DataType:    aws.String("String"),
				StringValue: aws.String(sms.Body),
			},
			"Broadcast": {
				DataType:    aws.String("String"),
				StringValue: aws.String(sms.Broadcast),
			},
		},
		MessageBody: aws.String("SMS Message"),
		QueueUrl:    &qURL,
		MessageGroupId: aws.String("metals"),
		MessageDeduplicationId: aws.String(dedupeID),
	})

	if err != nil {
		fmt.Println("Error", err)
		return SMSMessage{}, err
	}

	fmt.Println("Success", *result.MessageId)


	return SMSMessage{}, nil
}

func main() {
	lambda.Start(LambdaStart)
}
