package main

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

// Declare a new DynamoDB instance.
var db = dynamodb.New(session.New(), aws.NewConfig().WithRegion(S3_REGION))

// get item from DynamoDB by ID
func getItem(PhoneNumber string) (*Number, error) {
	// Prepare the input for the query.
	input := &dynamodb.GetItemInput{
		TableName: aws.String(table),
		Key: map[string]*dynamodb.AttributeValue{
			"phone_number": {
				S: aws.String(PhoneNumber),
			},
		},
	}

	// Retrieve the item from DynamoDB. If no matching item is found
	// return nil.
	result, err := db.GetItem(input)
	if err != nil {
		return nil, err
	}

	if result.Item == nil {
		return nil, nil
	}

	number := new(Number)
	err = dynamodbattribute.UnmarshalMap(result.Item, number)
	if err != nil {
		return nil, err
	}

	return number, nil
}
