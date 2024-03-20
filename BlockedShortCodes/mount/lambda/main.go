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

var (
	sess, _ = session.NewSession(&aws.Config{
		Region: aws.String(S3_REGION)},
	)

	downloader = s3manager.NewDownloader(sess)
)

var numbers []string
var wg sync.WaitGroup

func LambdaStart(bucket Bucket) (Bucket, error) {
	numbers = make([]string, 0)
	//receive bucket id of phone numbers
	//get bucket
	buf := aws.NewWriteAtBuffer([]byte{})

	_, err := downloader.Download(buf,
		&s3.GetObjectInput{
			Bucket: aws.String(S3_BUCKET),
			Key:    aws.String(bucket.Uuid),
		})

	if err != nil {
		fmt.Println(err)
		return Bucket{}, err
	}

	var messages Message
	err = json.Unmarshal(buf.Bytes(), &messages)
	if err != nil {
		fmt.Println(err)
		return Bucket{}, err
	}

	for i, phone := range messages.PhoneNumbers{
		if i%650 == 0 {
			wg.Wait()
		}
		wg.Add(1)
		go checkIfBlocked(phone)
 	}
	wg.Wait()
	messages.PhoneNumbers = numbers

	//put the new numbers in a bucket and return the id.
	return postToS3(&messages), nil
}

func checkIfBlocked(phone string) {
	defer wg.Done()
	//look in blocked shorts codes tbl
	number, err := getItem(phone)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	//remove numbers in bucket which are listed in blocked short codes
	if number == nil {
		numbers = append(numbers, phone)
	}
}

func postToS3(msgs *Message) (Bucket) {
	payload, err := json.Marshal(&msgs)
	if err != nil {
		fmt.Println(err)
		return Bucket{}
	}

	s, err := session.NewSession(&aws.Config{Region: aws.String(S3_REGION)})
	if err != nil {
		fmt.Println(err)
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
		fmt.Println(err)
		return Bucket{}
	}

	bucket.Uuid = bucketKey

	return bucket
}

func main() {
	lambda.Start(LambdaStart)
}

func noDashUUID() string {
	return strings.Replace(uuid.Must(uuid.NewV4()).String(), "-", "", -1)
}