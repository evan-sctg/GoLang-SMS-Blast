package main

import (
	"github.com/aws/aws-lambda-go/lambda"
)

type SMSMessage struct {
	ToNumber   string `json:"To"`
	FromNumber string `json:"From"`
	Body 	   string `json:"Body"`
}

type SMSMessageOut struct {
	ToNumber   string `json:"To"`
	FromNumber string `json:"From"`
	Body 	   string `json:"Body"`
	Broadcast  string `json:"Broadcast"`
}

type DNC struct {
	PhoneNumber   string `json:"phone_number"`
	Status         string `json:"status"`
}

func LambdaStart(event SMSMessage) (SMSMessageOut, error) {
	var dnc DNC
	dnc.PhoneNumber = event.FromNumber

	var smsOut SMSMessageOut

	smsOut.ToNumber = event.FromNumber
	smsOut.FromNumber = event.ToNumber
	smsOut.Broadcast = "System"



	if event.Body == "STOP" {
		dnc.Status = "1"
		smsOut.Body = "You are unsubscribed from SMS Blast Alerts. No more messages will be sent. Reply HELP for help or 555-555-5555."
	} else {
		dnc.Status = "0"
		smsOut.Body = "You have reactivated messages. Reply HELP for help, STOP to cancel. Msg&data rates may apply."
	}

	err := putItem(&dnc)
	if err != nil {
		return SMSMessageOut{}, err
	}

	return smsOut, nil

}

func main() {
	lambda.Start(LambdaStart)
}
