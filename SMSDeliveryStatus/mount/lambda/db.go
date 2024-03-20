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

type BroadcastStatusCount struct {
    BroadcastID 	 string `json:"broadcast_id"`
    DeliveredCount 	 int64 `json:"delivered"`
    UndeliveredCount int64 `json:"undelivered"`
    NoStatusCount 	 int64 `json:"no_status"`
}


func getBroadcastStatusCounts(broadcastID string) (BroadcastStatusCount, error) {

    b := BroadcastStatusCount{BroadcastID:broadcastID}

    // Get items with no delivery status
    input := &dynamodb.QueryInput{
        TableName: aws.String(table),
        IndexName: aws.String("delivery_status"),
        KeyConditionExpression: aws.String("delivery_status = :st AND broadcast_id = :id"),
        ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
            ":st": {
                S: aws.String("None"),
            },
            ":id": {
                S: aws.String(broadcastID),
            },
        },
    }

    output, err := db.Query(input)
    if err != nil {
        return b, err
    }

    b.NoStatusCount = *output.Count

    //Get items with delivered status
    input = &dynamodb.QueryInput{
        TableName: aws.String(table),
        IndexName: aws.String("delivery_status"),
        KeyConditionExpression: aws.String("delivery_status = :st AND broadcast_id = :id"),
        ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
            ":st": {
                S: aws.String("delivered"),
            },
            ":id": {
                S: aws.String(broadcastID),
            },
        },
    }

    output, err = db.Query(input)
    if err != nil {
        return b, err
    }

    b.DeliveredCount = *output.Count

    //Get items with undelivered status
    input = &dynamodb.QueryInput{
        TableName: aws.String(table),
        IndexName: aws.String("delivery_status"),
        KeyConditionExpression: aws.String("delivery_status = :st AND broadcast_id = :id"),
        ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
            ":st": {
                S: aws.String("undelivered"),
            },
            ":id": {
                S: aws.String(broadcastID),
            },
        },
    }

    output, err = db.Query(input)
    if err != nil {
        return b, err
    }

    b.UndeliveredCount = *output.Count

    return b, nil

}


func getMessageIDs() ([]DeliveryStatus, error){
    input := &dynamodb.QueryInput{
        TableName: aws.String(table),
        IndexName: aws.String("delivery_status"),
        KeyConditionExpression: aws.String("delivery_status = :st"),
        ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
            ":st": {
                S: aws.String("None"),
            },
        },
        Limit: aws.Int64(2),
    }

    output, err := db.Query(input)
    if err != nil {
        return nil, err
    }

    var d []DeliveryStatus

    err = dynamodbattribute.UnmarshalListOfMaps(output.Items, &d)
    if err != nil {
        return nil, err
    }

    return d, nil

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