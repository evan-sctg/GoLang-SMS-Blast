package main

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/viki-org/dnscache"
	"strconv"
	"sync"
	"time"
)

type PhoneNumbers struct {
	Numbers []string `json:"numbers"`
	Uuid string `json:"uuid"`
	Message string `json:"message"`
}

type Bucket struct {
	Uuid   	 string   `json:"uuid"`
}

const (
 	S3_REGION = "us-east-1"
 	S3_BUCKET = "aurora-db-contact-list-by-broadcast-id-live"
 	FromNumber = "*****"
)

var (
	sess, _ = session.NewSession(&aws.Config{
		Region: aws.String(S3_REGION)},
	)
	// s3 bucket session
	downloader = s3manager.NewDownloader(sess)
	// sqs session
	svc = sqs.New(sess)
	// URL to our queue
	//qURL = "https://sqs.us-east-1.amazonaws.com/*********/SMSQueue_Debug.fifo"
)

//refresh items every 45 minutes
var resolver = dnscache.New(time.Minute * 45)

var ip_SQS, _ = resolver.FetchOneString("sqs.us-east-1.amazonaws.com")
// var qURL = "https://" + ip_SQS + "//*********//SMSQueue.fifo"
var qURL = "https://" + ip_SQS + "//*********//SMSQueue_Debug.fifo"

var wg sync.WaitGroup

func LambdaStart(bkt Bucket) (string, error) {

	buf := aws.NewWriteAtBuffer([]byte{})

	_, err := downloader.Download(buf,
		&s3.GetObjectInput{
			Bucket: aws.String(S3_BUCKET),
			Key:    aws.String(bkt.Uuid),
		})
	if err != nil {
		fmt.Println("Problem reading item from bucket: " + err.Error())
		return "", err
	}

	var messages PhoneNumbers
	err = json.Unmarshal(buf.Bytes(), &messages)
	if err != nil {
		fmt.Println("failed to unmarshal bucket data: " + err.Error())
		return "", err
	}

	//limit := limiter.NewConcurrencyLimiter(MaxThreads)
	for i, phone := range messages.Numbers {
		if i%300 == 0 {
			wg.Wait()
		}
		wg.Add(1)
		go addToSQS(phone, FromNumber, &messages.Uuid, &messages.Message)
		//time.Sleep(3 * time.Millisecond)
	}
	wg.Wait()

	return "", nil

}

func addToSQS(toNumber string, fromNumber string, broadcast *string, message *string) {
	defer wg.Done()
	// Create a unique key for message deduplication
	dedupeID := toNumber + ":" + fromNumber + ":" + strconv.Itoa(int(time.Now().UnixNano()))
	// Send to SQS Queue
	_, err := svc.SendMessage(&sqs.SendMessageInput{
		MessageAttributes: map[string]*sqs.MessageAttributeValue{
			"ToNumber": {
				DataType:    aws.String("String"),
				StringValue: aws.String(toNumber),
			},
			"FromNumber": {
				DataType:    aws.String("String"),
				StringValue: aws.String(fromNumber),
			},
			"Body": {
				DataType:    aws.String("String"),
				StringValue: aws.String(*message  + " STOP to cancel"),
			},
			"Broadcast": {
				DataType:    aws.String("String"),
				StringValue: aws.String(*broadcast),
			},
		},
		MessageBody: aws.String("SMS Message"),
		QueueUrl:    &qURL,
		MessageGroupId: aws.String(dedupeID),
		MessageDeduplicationId: aws.String(dedupeID),
	})

	if err != nil {
		fmt.Println(err)
	}
}


func main() {
	lambda.Start(LambdaStart)
}
