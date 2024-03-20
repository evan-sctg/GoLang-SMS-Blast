package main

import (
    "github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/aws/session"
    "github.com/aws/aws-sdk-go/service/dynamodb"
    "github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

const table = "Broadcasts"

type Broadcast struct {
    ID  string              `json:"id"`
    Name string             `json:"name"`
    UnixTime int64          `json:"unix_time"`
    Message string          `json:"message"`
    SendTime int64          `json:"send_time"`
    Processed string        `json:"processed"`
    Successes string        `json:"successes"`
    Failures string         `json:"failures"`
    Sent string             `json:"sent"`
    YearMonth string        `json:"year_month"`
    UserUUID string         `json:"user_uuid"`
    UserDisplayName string  `json:"user_display_name"`
}

type Results struct {
    Broadcasts []Broadcast                      `json:"broadcasts"`
    LastID string                               `json:"last_id"`
    LastUnixTime string                         `json:"last_unix_time"`
    LastYearMonth string                        `json:"last_year_month"`
    ConsumedCapacity dynamodb.ConsumedCapacity  `json:"consumed_capacity"`
}

type PutResults struct {
    Success bool `json:"success"`
    ConsumedCapacity dynamodb.ConsumedCapacity  `json:"consumed_capacity"`
}

// Declare a new DynamoDB instance.
var db = dynamodb.New(session.New(), aws.NewConfig().WithRegion("us-east-1"))

func queryBetween(unixTime string, yearMonth string) (Results, error) {
    input := &dynamodb.QueryInput{
        TableName: aws.String(table),
        IndexName: aws.String("year_month"),
        KeyConditionExpression: aws.String("year_month = :ym AND unix_time <= :ut"),
        FilterExpression: aws.String("#proc = :pr"),
        ScanIndexForward: aws.Bool(false),
        ExpressionAttributeNames: map[string]*string {
            "#proc": aws.String("processed"),
        },
        ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
            ":ut": {
                N: aws.String(unixTime),
            },
            ":ym": {
                S: aws.String(yearMonth),
            },
            ":pr": {
                S: aws.String("false"),
            },
        },
        Limit: aws.Int64(10), // todo:remove this later
    }

    input.SetReturnConsumedCapacity(dynamodb.ReturnConsumedCapacityTotal)

    output, err := db.Query(input)
    if err != nil {
        return Results{}, err
    }

    var bc []Broadcast

    err = dynamodbattribute.UnmarshalListOfMaps(output.Items, &bc)
    if err != nil {
        return Results{}, err
    }

    var results Results

    results.Broadcasts = bc

    results.ConsumedCapacity = *output.ConsumedCapacity

    return results, nil
}

// Add a record to DynamoDB.
func putItem(CL *Broadcast) (PutResults, error) {

    //create map from struct
    av, errmarsh := dynamodbattribute.MarshalMap(CL)
    if errmarsh != nil {
        return PutResults{Success:false}, errmarsh
    }

    //Setup input Item
    input := &dynamodb.PutItemInput{
        TableName: aws.String(table),
        Item: av,
    }

    input.SetReturnConsumedCapacity(dynamodb.ReturnConsumedCapacityTotal)

    //store Item
    output, err := db.PutItem(input)
    if err != nil {
        return PutResults{Success:false}, err
    }

    return PutResults{Success:true, ConsumedCapacity: *output.ConsumedCapacity}, nil
}