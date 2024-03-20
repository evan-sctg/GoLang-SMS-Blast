package main

import (
	"errors"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/buger/jsonparser"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
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
	qURL = "https://sqs.us-east-1.amazonaws.com/**************/SMSQueue.fifo"
)

const (
	twillioAccountSid = "*********************"
	twillioAuthToken = "*********************"
)

var msg_count_semaphore    = make(chan struct{}, 1)
var msgCount = 0

var sendCount = 0
var pullCount = 0
var wg sync.WaitGroup

func SendSMS(toPhoneNumberStr *string, fromPhoneNumberStr *string, message *string) (string, error) {
	urlStr := "https://api.twilio.com/2010-04-01/Accounts/" + twillioAccountSid + "/Messages.json"

	// Pack up the data for our message
	msgData := url.Values{}
	msgData.Set("To", "+"+*toPhoneNumberStr)
	//msgData.Set("From", "+"+*fromPhoneNumberStr)
	msgData.Set("MessagingServiceSid", "**************")
	msgData.Set("StatusCallback", "https://**************.execute-api.us-east-1.amazonaws.com/default/ReceiveTwilioStatuses")
	msgData.Set("Body", *message)
	msgDataReader := *strings.NewReader(msgData.Encode())

	// Create HTTP request client
	client := &http.Client{}
	req, _ := http.NewRequest("POST", urlStr, &msgDataReader)
	req.SetBasicAuth(twillioAccountSid, twillioAuthToken)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	// Send req
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Read response into bytes
	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	//TODO: handle more response codes
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		// Get the sid out of the response from twilio
		sid, err := jsonparser.GetString(respBytes, "sid")
		if err != nil {
			return "", err
		}
		sendCount++

		return sid, nil

	} else {
		// Unhandled error, return the whole response to be logged
		return "", errors.New(string(respBytes))
	}
}

func LambdaStart(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	sendCount = 0
	pullCount = 0
	msgCount = 0

	num, err := strconv.Atoi(req.QueryStringParameters["num"])
	if err != nil {
		return serverError(err)
	}

	var deliveredCount int
	fmt.Println("loop starting")
	for i := 0; i < num; i++ {
		if i % 75 == 0 {
			if msgCount > 0 {
				time.Sleep(1200 * time.Millisecond)
				msgCount = 0
			}
		}
		if i % 400 == 0 {
			wg.Wait()
		}
		wg.Add(1)
		go retrieveFromQueue()
	}

	wg.Wait()
	fmt.Println("loop finished")
	fmt.Println("send count: " + strconv.Itoa(sendCount))
	fmt.Println("pull count: " + strconv.Itoa(pullCount))

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body: "Success, sent " + strconv.Itoa(deliveredCount) + " SMS messages! Number of fails: " + strconv.Itoa(num - deliveredCount),
	}, nil
}

func retrieveFromQueue() {
	defer wg.Done()
	msg_count_semaphore <- struct{}{}
	msgCount ++
	<-msg_count_semaphore
	// Poll for messages from SQS queue
	result, err := svc.ReceiveMessage(&sqs.ReceiveMessageInput{
		AttributeNames: []*string{
			aws.String(sqs.MessageSystemAttributeNameSentTimestamp),
		},
		MessageAttributeNames: []*string{
			aws.String(sqs.QueueAttributeNameAll),
		},
		QueueUrl:            &qURL,
		MaxNumberOfMessages: aws.Int64(int64(1)),
		VisibilityTimeout:   aws.Int64(20), // 20 seconds
		WaitTimeSeconds:     aws.Int64(0),
	})

	if err != nil {
		fmt.Println(err)
		return
	}

	if len(result.Messages) == 0 {
		return
	}
	pullCount ++
	for _, message := range result.Messages {

		// Delete message from SQS queue
		_, err = svc.DeleteMessage(&sqs.DeleteMessageInput{
			QueueUrl:      &qURL,
			ReceiptHandle: message.ReceiptHandle,
		})
		if err != nil {
			fmt.Println("Failed to delete message from sqs queue. Reason: " + err.Error())
			continue
		}
		to := message.MessageAttributes["ToNumber"]
		from := message.MessageAttributes["FromNumber"]
		body := message.MessageAttributes["Body"]

		sid, err := SendSMS(to.StringValue, from.StringValue, body.StringValue)

		if err != nil {
			fmt.Println("Failed to send text message to: " + *to.StringValue + " Reason: " + err.Error())
			continue
		} else {
			// put sid into delivery status dynamo with status of queued
			err := putItem(&DeliveryStatus{
				BroadcastID: *message.MessageAttributes["Broadcast"].StringValue,
				Status: "queued",
				MessageID: sid,
			})
			if err != nil {
				fmt.Println("Failed to insert delivery status into dynamodb: " + err.Error())
				continue
			}
		}

	}
}


func serverError(err error) (events.APIGatewayProxyResponse, error) {
	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusInternalServerError,
		Body:       http.StatusText(http.StatusInternalServerError) + " [" + err.Error() + "]",
	}, nil
}

// Similarly add a helper for send responses relating to client errors.
func clientError(status int) (events.APIGatewayProxyResponse, error) {
	return events.APIGatewayProxyResponse{
		StatusCode: status,
		Body:       http.StatusText(status),
	}, nil
}


func main() {
	lambda.Start(LambdaStart)
}
