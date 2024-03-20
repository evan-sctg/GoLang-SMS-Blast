package main

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

const table = "SMSDeliveryStatus"


// Declare a new DynamoDB instance.
var db = dynamodb.New(session.New(), aws.NewConfig().WithRegion("us-east-1"))

type DeliveryStatus struct {
	MessageID 	string `json:"message_id"`
	Status 	  	string `json:"delivery_status"`
	BroadcastID string `json:"broadcast_id"`
}

// Add a record to DynamoDB.
func putItem(item *DeliveryStatus) error {

	//create map from struct
	av, errmarsh := dynamodbattribute.MarshalMap(item)
	if errmarsh != nil {
		return errmarsh
	}

	//alter map to store unix time as a number
	// av["unix_time"]= &dynamodb.AttributeValue{
	//     N: aws.String(CL.Unix_time),
	// }

	//Setup input Item
	input := &dynamodb.PutItemInput{
		TableName: aws.String(table),
		Item: av,
	}
	//store Item
	_, err := db.PutItem(input)
	return err
}