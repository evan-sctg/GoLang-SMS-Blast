package main

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func LambdaStart(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	batchSize := 1000
	page := 1

	pg := req.QueryStringParameters["page"]
	batch := req.QueryStringParameters["batch_size"]
	if batch != "" {
		batchSize, _ = strconv.Atoi(batch)
	}
	if pg != "" {
		page, _ = strconv.Atoi(pg)
	}

	err := getAuroraDNC(batchSize, batchSize*page)


	if err != nil {
		fmt.Println("failed to update dncs")
		return serverError(err)
	}

	return events.APIGatewayProxyResponse{
		StatusCode:200,
		Body: "SUCCESS!",
	}, nil
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
	dbConn()
	lambda.Start(LambdaStart)
}
