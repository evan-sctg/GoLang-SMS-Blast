package main

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func LambdaStart(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	if req.HTTPMethod == "GET" {
		//return status of phone number
		phone := req.QueryStringParameters["phone_number"]
		if phone != "" {
			dnc, err := getItem(phone)
			if err != nil {
				return serverError(err)
			}

			if dnc == nil {
				var d DNC
				d.PhoneNumber = phone
				d.Status = "0"
				dnc = &d
			}


			res, err := json.Marshal(dnc)
			if err != nil {
				return serverError(err)
			}

			return events.APIGatewayProxyResponse{
				StatusCode: 200,
				Body:       string(res),
			}, nil
		} else {
			return serverError(errors.New("phone_number field is empty"))
		}

	} else if req.HTTPMethod == "POST" {
		bulkStatus := req.QueryStringParameters["bulk_status"]
		if req.Body != "" {
			if bulkStatus != "" {
				var dbs DNCBulkStatus

				if err := json.Unmarshal([]byte(req.Body), &dbs); err != nil {
					return serverError(err)
				}

				//var DNCResults []DNC

				var PhoneNumbers []string

				for _, phone := range dbs.PhoneNumbers {
					dnc, getErr := getItem(phone)
					if getErr != nil {
						return serverError(getErr)
					}
					if dnc == nil {
						PhoneNumbers = append(PhoneNumbers, phone)
						continue

					}
					if dnc.Status == "0" {
						PhoneNumbers = append(PhoneNumbers, dnc.PhoneNumber)
					}
				}

				res, err := json.Marshal(&PhoneNumbers)
				if err != nil {
					return serverError(err)
				}

				return events.APIGatewayProxyResponse{
					StatusCode: 200,
					Body:       string(res),
				}, nil
			}

			var d []DNC

			if err := json.Unmarshal([]byte(req.Body), &d); err != nil {
				return serverError(err)
			}

			for _, dncSingle := range d {
				if dncSingle.PhoneNumber == "" || dncSingle.Status == "" {
					return serverError(errors.New("required field is missing"))
				}

				if err := putItem(&dncSingle); err != nil {
					return serverError(err)
				}
			}

			return events.APIGatewayProxyResponse{
				StatusCode: 200,
				Body:       string("Success!"),
			}, nil

		} else {
			return serverError(errors.New("required field is missing"))
		}

	}

	return serverError(errors.New("invalid request"))

}

// Add a helper for handling errors.
// returns a 500 Internal Server Error response that the AWS API
// Gateway understands.
func serverError(err error) (events.APIGatewayProxyResponse, error) {

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusInternalServerError,
		Body:       http.StatusText(http.StatusInternalServerError) + " [" + err.Error() + "]",
	}, nil
}

// Similarly add a helper for send responses relating to client errors.
func clientError(status int) (events.APIGatewayProxyResponse, error) {
	return events.APIGatewayProxyResponse{
		StatusCode: status,
		Body:       http.StatusText(status),
	}, nil
}

func main() {
	lambda.Start(LambdaStart)
}
