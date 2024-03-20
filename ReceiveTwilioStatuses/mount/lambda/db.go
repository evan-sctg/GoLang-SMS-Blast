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

// get item from DynamoDB by ID
func getItem(sid string) (*DeliveryStatus, error) {
    // Prepare the input for the query.
    input := &dynamodb.GetItemInput{
        TableName: aws.String(table),
        Key: map[string]*dynamodb.AttributeValue{
            "message_id": {
                S: aws.String(sid),
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

    CL := new(DeliveryStatus)
    err = dynamodbattribute.UnmarshalMap(result.Item, CL)
    if err != nil {
        return nil, err
    }

    return CL, nil
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