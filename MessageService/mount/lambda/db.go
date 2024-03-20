package main

import (
    "github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/aws/session"
    "github.com/aws/aws-sdk-go/service/dynamodb"
    "github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

const table = "MessageServices"

// Declare a new DynamoDB instance. 
var db = dynamodb.New(session.New(), aws.NewConfig().WithRegion("us-east-1"))

// get item from DynamoDB by ID
func getItem(PhoneNumber string) (*PhoneNumberSID, error) {
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

    CL := new(PhoneNumberSID)
    err = dynamodbattribute.UnmarshalMap(result.Item, CL)
    if err != nil {
        return nil, err
    }

    return CL, nil
}

func queryItems(sid string) (*[]PhoneNumberSID, error) {
    input := &dynamodb.QueryInput{
        TableName: aws.String(table),
        IndexName: aws.String("service_sid-index"),
        KeyConditionExpression: aws.String("service_sid = :sid"),
        ScanIndexForward: aws.Bool(false),
        ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
            ":sid": {
                S: aws.String(sid),
            },
        },
    }
    //input.SetReturnConsumedCapacity(dynamodb.ReturnConsumedCapacityTotal)

    //if len(startKey) > 0 {
    //    input.SetExclusiveStartKey(startKey[0])
    //}

    output, err := db.Query(input)
    if err != nil {
        return nil, err
    }

    var results []PhoneNumberSID

    err = dynamodbattribute.UnmarshalListOfMaps(output.Items, &results)
    if err != nil {
        return nil, err
    }


    return &results, nil
}


// Add a record to DynamoDB.
func putItem(CL *PhoneNumberSID) error {

    //create map from struct
    av, errmarsh := dynamodbattribute.MarshalMap(CL)
    if errmarsh != nil {
        return errmarsh
    }

    //Setup input Item
    input := &dynamodb.PutItemInput{
        TableName: aws.String(table),
        Item: av,
    }

    //store Item
    _, err := db.PutItem(input)
    return err
}


func deleteItem(phone string) error {

    input := &dynamodb.DeleteItemInput{
        TableName: aws.String(table),
        Key: map[string]*dynamodb.AttributeValue{
            "phone_number": { S: aws.String(phone) },
        },
    }

    _, err := db.DeleteItem(input)

    return err
}