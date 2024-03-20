package main

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

type PhoneNumberSID struct {
	PhoneNumber string `json:"phone_number"`
	SID string	`json:"service_sid"`
}

func LambdaStart(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	if req.HTTPMethod == "GET" {
		phone := req.QueryStringParameters["phone"]
		sid := req.QueryStringParameters["sid"]

		if phone != "" {
			output, err := getItem(phone)
			if err != nil {
				return serverError(err)
			}

			jsonBytes, err := json.Marshal(&output)
			if err != nil {
				return serverError(err)
			}

			return success(string(jsonBytes))

		} else if sid != "" {
			output, err := queryItems(sid)
			if err != nil {
				return serverError(err)
			}

			jsonBytes, err := json.Marshal(&output)
			if err != nil {
				return serverError(err)
			}

			return success(string(jsonBytes))
		}

	} else if req.HTTPMethod == "POST" {
		if req.Body != "" {
			var postData PhoneNumberSID

			if err := json.Unmarshal([]byte(req.Body), &postData); err != nil {
				return serverError(err)
			}

			err := putItem(&postData)
			if err != nil {
				return serverError(err)
			}

			return success("success!")

		}
	} else if req.HTTPMethod == "DELETE" {
		phone := req.QueryStringParameters["phone"]
		if phone != "" {
			err := deleteItem(phone)
			if err != nil {
				return serverError(err)
			}

			return success("deleted item!")
		}

	}

	return serverError(errors.New("invalid request"))

}

// Add a helper for handling errors.
// returns a 500 Internal Server Error response that the AWS API
// Gateway understands.

func success(resp string) (events.APIGatewayProxyResponse, error) {
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       resp,
	}, nil
}

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
