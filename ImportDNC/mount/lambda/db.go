package main

import (
    "database/sql"
    "github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/aws/session"
    "github.com/aws/aws-sdk-go/service/dynamodb"
    "github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
    "fmt"
)
import _ "github.com/go-sql-driver/mysql"

const(
    table = "DNC"

    dbName = "blast-sms-live"
    dbUser = "SMSblast"
    dbPassword = "********************"
    dbHost  = "***.***.***.***"
    dbPort = "3308"
    dbDriver = "mysql"
)

type DNC struct {
    PhoneNumber   string `json:"phone_number"`
    Status         string `json:"status"`
}

var sqlDB *sql.DB

// Declare a new DynamoDB instance. 
var db = dynamodb.New(session.New(), aws.NewConfig().WithRegion("us-east-1"))

// AuroraDB connection
func dbConn() {
    var err error
    sqlDB, err = sql.Open(dbDriver, dbUser+":"+dbPassword+"@tcp("+dbHost+":"+dbPort+")/"+dbName)
    if err != nil{
        fmt.Println(err)
    }
}

func getAuroraDNC(limit int, offset int) error {
    //var dncs []DNC

    rows, err := sqlDB.Query(`SELECT dnc_status, phone164c FROM do_not_call LIMIT ? OFFSET ?`, limit, offset)
    if err != nil {
        return err
    }
    for rows.Next() {
        var dnc DNC
        err := rows.Scan(&dnc.Status, &dnc.PhoneNumber)
        if err != nil {
            return err
        }

        err = putItem(&dnc)
        if err != nil {
            return err
        }
    }

    return nil
}

// get item from DynamoDB by ID
func getItem(PhoneNumber string) (*DNC, error) {
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

    CL := new(DNC)
    err = dynamodbattribute.UnmarshalMap(result.Item, CL)
    if err != nil {
        return nil, err
    }

    return CL, nil
}


// Add a record to DynamoDB.
func putItem(CL *DNC) error {

    //create map from struct
    av, errmarsh := dynamodbattribute.MarshalMap(CL)
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