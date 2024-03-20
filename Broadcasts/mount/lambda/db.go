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
    Count              int64                    `json:"count"`
}

type PutResults struct {
    Success bool `json:"success"`
    ConsumedCapacity dynamodb.ConsumedCapacity  `json:"consumed_capacity"`
}

// Declare a new DynamoDB instance.
var db = dynamodb.New(session.New(), aws.NewConfig().WithRegion("us-east-1"))

func queryBetween(startTime string, endTime string, yearMonth string, limit int64, count bool) (Results, error) {
    input := &dynamodb.QueryInput{
        TableName: aws.String(table),
        IndexName: aws.String("year_month"),
        KeyConditionExpression: aws.String("year_month = :ym AND unix_time BETWEEN :st AND :et"),
        ScanIndexForward: aws.Bool(true),
        ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
            ":st": {
                N: aws.String(startTime),
            },
            ":et": {
                N: aws.String(endTime),
            },
            ":ym": {
                S: aws.String(yearMonth),
            },
        },
    }
    input.SetReturnConsumedCapacity(dynamodb.ReturnConsumedCapacityTotal)
    if count {
        input.SetSelect(dynamodb.SelectCount)
    } else {
        input.SetLimit(limit)
    }

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

    results.Count = *output.Count

    results.Broadcasts = bc

    results.ConsumedCapacity = *output.ConsumedCapacity

    return results, nil
}

func searchItems(unixTime string, yearMonth string, searchQuery string, limit int64, sortForward bool, count bool, startKey ...map[string]*dynamodb.AttributeValue) (Results, error) {
    input := &dynamodb.QueryInput{
        TableName: aws.String(table),
        IndexName: aws.String("year_month"),
        KeyConditionExpression: aws.String("year_month = :ym AND unix_time <= :ut"),
        FilterExpression:aws.String("contains(#nm, :sq) OR contains(message, :sq)"),
        ScanIndexForward: aws.Bool(sortForward),
        ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
            ":ut": {
                N: aws.String(unixTime),
            },
            ":ym": {
                S: aws.String(yearMonth),
            },
            ":sq": {
                S: aws.String(searchQuery),
            },
        },
        ExpressionAttributeNames: map[string]*string{
            "#nm": aws.String("name"),
        },
    }
    //FilterExpression: aws.String(""),
    input.SetReturnConsumedCapacity(dynamodb.ReturnConsumedCapacityTotal)
    if count {
        input.SetSelect(dynamodb.SelectCount)
    } else {
        input.SetLimit(limit)
    }

    if len(startKey) > 0 {
        input.SetExclusiveStartKey(startKey[0])
    }

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



    if _, err := output.LastEvaluatedKey["id"]; err {
        results.LastID = *output.LastEvaluatedKey["id"].S
        results.LastUnixTime = *output.LastEvaluatedKey["unix_time"].N
        results.LastYearMonth = *output.LastEvaluatedKey["year_month"].S
    }

    results.ConsumedCapacity = *output.ConsumedCapacity
    results.Count = *output.Count

    return results, nil
}

func queryItems(unixTime string, yearMonth string, limit int64, sortForward bool, count bool, startKey ...map[string]*dynamodb.AttributeValue) (Results, error) {
    input := &dynamodb.QueryInput{
        TableName: aws.String(table),
        IndexName: aws.String("year_month"),
        KeyConditionExpression: aws.String("year_month = :ym AND unix_time <= :ut"),
        ScanIndexForward: aws.Bool(sortForward),
        ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
            ":ut": {
                N: aws.String(unixTime),
            },
            ":ym": {
                S: aws.String(yearMonth),
            },
        },
    }
    //input.SetSelect(dynamodb.SelectCount)
    //FilterExpression: aws.String(""),
    input.SetReturnConsumedCapacity(dynamodb.ReturnConsumedCapacityTotal)
    if count {
        input.SetSelect(dynamodb.SelectCount)
    } else {
        input.SetLimit(limit)
    }

    if len(startKey) > 0 {
        input.SetExclusiveStartKey(startKey[0])
    }

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



    if _, err := output.LastEvaluatedKey["id"]; err {
        results.LastID = *output.LastEvaluatedKey["id"].S
        results.LastUnixTime = *output.LastEvaluatedKey["unix_time"].N
        results.LastYearMonth = *output.LastEvaluatedKey["year_month"].S
    }

    results.ConsumedCapacity = *output.ConsumedCapacity
    results.Count = *output.Count

    return results, nil
}

// get item from DynamoDB by ID
func getItem(ID string) (*Broadcast, error) {
    // Prepare the input for the query.
    input := &dynamodb.GetItemInput{
        TableName: aws.String(table),
        Key: map[string]*dynamodb.AttributeValue{
            "id": {
                S: aws.String(ID),
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

    CL := new(Broadcast)
    err = dynamodbattribute.UnmarshalMap(result.Item, CL)
    if err != nil {
        return nil, err
    }

    return CL, nil
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