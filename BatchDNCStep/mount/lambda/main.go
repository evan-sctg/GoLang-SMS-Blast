package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	uuid "github.com/satori/go.uuid"
	"net/http"
	"strings"
	"sync"
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
)

var (
	sess, _ = session.NewSession(&aws.Config{
		Region: aws.String(S3_REGION)},
	)
	// s3 bucket session
	downloader = s3manager.NewDownloader(sess)
)



var semaphore_cnt_err    = make(chan struct{}, 1)
var semaphore_cnt_dnc    = make(chan struct{}, 1)
var semaphore_cnt    = make(chan struct{}, 1)
var semaphore_numbers    = make(chan struct{}, 1)


var wg sync.WaitGroup



var numbers []string
var DNCCount = 0
var ErrorCount = 0
var OKCount = 0
var checkedNums = 0

func LambdaStart(bkt Bucket) (Bucket, error) {
	numbers = make([]string, 0)
	DNCCount = 0
	ErrorCount = 0
	OKCount = 0
	checkedNums = 0

	buf := aws.NewWriteAtBuffer([]byte{})

	_, err := downloader.Download(buf,
		&s3.GetObjectInput{
			Bucket: aws.String(S3_BUCKET),
			Key:    aws.String(bkt.Uuid),
		})
	if err != nil {
		fmt.Println("Problem reading item from bucket: " + err.Error())
		return Bucket{}, err
	}

	var messages PhoneNumbers
	err = json.Unmarshal(buf.Bytes(), &messages)
	if err != nil {
		fmt.Println("failed to unmarshal bucket data: " + err.Error())
		return Bucket{}, err
	}


	var StartedGoRoutines = 0
	for i, phone := range messages.Numbers {
		if i%650 == 0 {
			wg.Wait()
		}
		wg.Add(1)
		StartedGoRoutines++
		go getAndAdd(phone)
	}
	wg.Wait()



	fmt.Println("messages.Numbers: ", len(messages.Numbers))
	fmt.Println("StartedGoRoutines: ", StartedGoRoutines)
	fmt.Println("checkedNums: ", checkedNums)
	fmt.Println("DNCCount: ", DNCCount)
	fmt.Println("ErrorCount: ", ErrorCount)
	fmt.Println("OKCount: ", OKCount)
	fmt.Println("verify OK: ", (checkedNums - DNCCount - ErrorCount))
	fmt.Println("output number count: ", len(numbers))

	messages.Numbers = numbers
	newBkt := postToS3(&messages)
	return newBkt, nil

}



func getAndAdd(phone string) {
	semaphore_cnt <- struct{}{}
	checkedNums++
	<-semaphore_cnt
	defer wg.Done()
	dnc, getErr := getItem(phone)
	if getErr != nil {
		semaphore_cnt_err <- struct{}{}
		ErrorCount++
		<-semaphore_cnt_err
		fmt.Println(getErr.Error())
		return
	}

	if dnc == nil {
		semaphore_numbers <- struct{}{}
		OKCount++
		numbers = append(numbers, phone)
		<-semaphore_numbers
	} else if dnc.Status == "0" {
		semaphore_numbers <- struct{}{}
		OKCount++
		numbers = append(numbers, phone)
		<-semaphore_numbers
	}else{
		semaphore_cnt_dnc <- struct{}{}
		DNCCount++
		<-semaphore_cnt_dnc
	}
}

func postToS3(msgs *PhoneNumbers) (Bucket) {
	payload, err := json.Marshal(&msgs)
	if err != nil {
		fmt.Println("Error: postToS3 json.Marshal")
		fmt.Println(err.Error())
		return Bucket{}
	}

	s, err := session.NewSession(&aws.Config{Region: aws.String(S3_REGION)})
	if err != nil {
		fmt.Println("Error: postToS3 session.NewSession")
		fmt.Println(err.Error())
		return Bucket{}
	}

	var bucket Bucket
	bucketKey := noDashUUID()

	_, err = s3.New(s).PutObject(&s3.PutObjectInput{
		Bucket:               aws.String(S3_BUCKET),
		Key:                  aws.String(bucketKey),
		ACL:                  aws.String("private"),
		Body:                 bytes.NewReader(payload),
		ContentLength:        aws.Int64(int64(len(payload))),
		ContentType:          aws.String(http.DetectContentType(payload)),
	})

	if err != nil{
		fmt.Println("Error: postToS3 PutObject")
		fmt.Println(err.Error())
		return Bucket{}
	}

	bucket.Uuid = bucketKey

	return bucket
}

func noDashUUID() string {
	return strings.Replace(uuid.Must(uuid.NewV4()).String(), "-", "", -1)
}

func main() {
	lambda.Start(LambdaStart)
}
