package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/aws/aws-lambda-go/lambda"
	uuid "github.com/satori/go.uuid"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

func main() {
	dbConn()
	lambda.Start(LambdaStart)
}

func LambdaStart(uuid BroadcastUUID) (Bucket, error) {
	var rows PhoneNumbers

	if len(uuid.Uuid) != uuidLength {
		return Bucket{}, errors.New("Broadcast uuid must be " + strconv.Itoa(uuidLength) + " characters long")
	} else {
		rows = getBrodcastList(uuid)
	}

	rows.Uuid = uuid.Uuid
	rows.Message = uuid.Message

	bucket := postToS3(rows)

	return bucket, nil
}

func postToS3(rows PhoneNumbers) Bucket {
	payload, err := json.Marshal(rows)
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
		Bucket:        aws.String(S3_BUCKET),
		Key:           aws.String(bucketKey),
		ACL:           aws.String("private"),
		Body:          bytes.NewReader(payload),
		ContentLength: aws.Int64(int64(len(payload))),
		ContentType:   aws.String(http.DetectContentType(payload)),
	})

	if err != nil {
		fmt.Println(err)
		return Bucket{}
	}

	bucket.Uuid = bucketKey

	return bucket
}

func noDashUUID() string {
	return strings.Replace(uuid.Must(uuid.NewV4()).String(), "-", "", -1)
}
