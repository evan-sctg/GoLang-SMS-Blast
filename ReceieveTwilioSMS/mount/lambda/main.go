package main

import (
	"database/sql"
	"github.com/aws/aws-lambda-go/lambda"
	"strings"
	"fmt"
)
import _ "github.com/go-sql-driver/mysql"

type SMSMessage struct {
	ToNumber   string `json:"To"`
	FromNumber string `json:"From"`
	Body 	   string `json:"Body"`
}

const dbName = "SMSBlast"
const dbUser = "SMSBlast"
const dbPassword = "***************"
const dbHost = "***.***.***.***"
const dbPort = "3308"
const dbDriver = "mysql"

var db *sql.DB

func dbConn() {
	var err error
	db, err = sql.Open(dbDriver, dbUser+":"+dbPassword+"@tcp("+dbHost+":"+dbPort+")/"+dbName)
	if err != nil {
		fmt.Println(err)
	}
}

func LambdaStart(sms SMSMessage) (SMSMessage, error) {
	/*var sms SMSMessage

	err := json.Unmarshal([]byte(req.Body), &sms)
	if err != nil {
		return SMSMessage{}, err
	}*/

	sms.Body = strings.ToUpper(sms.Body)
	sms.FromNumber = strings.Replace(sms.FromNumber,"+", "", 1)
	sms.ToNumber = strings.Replace(sms.ToNumber,"+", "", 1)

	_, err := db.Exec(`INSERT INTO twillio_responses(response, phone_164c, to_number) VALUES(?, ?, ?)`, sms.Body, sms.FromNumber, sms.ToNumber)
	if err != nil {
		fmt.Println(err)
	}

	return sms, nil
}

func main() {
	dbConn()
	lambda.Start(LambdaStart)
}
