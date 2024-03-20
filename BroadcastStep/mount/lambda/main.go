package main

import (
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"net/http"
	"strconv"
	"time"
)

type BroadcastUUID struct {
	Uuid string `json:"uuid"`
	Message string `json:"message"`
}

func LambdaStart() (BroadcastUUID, error) {
	//var bcs []Broadcast
	var bc BroadcastUUID

	date := time.Now().UTC()
	yearMonth := strconv.Itoa(date.Year()) + "_" + date.Month().String()
	unixTime := strconv.FormatInt(date.Unix(), 10)
	fmt.Println(yearMonth)
	fmt.Println(unixTime)

	results, err := queryBetween(unixTime, yearMonth)

	if err != nil {
		fmt.Println(err.Error())
		return bc, err
	}
	if len(results.Broadcasts) > 0 {
		bc.Message = results.Broadcasts[0].Message
		bc.Uuid = results.Broadcasts[0].ID

		var broadcast Broadcast
		broadcast = results.Broadcasts[0]
		broadcast.Processed = "true"
		_, err = putItem(&broadcast)
		if err != nil {
			fmt.Println(err.Error())
			return bc, err
		}
	}

	return bc, nil

	//for _, broadcast := range results.Broadcasts {
	//	broadcast.Processed = "true"
	//	_, err := putItem(&broadcast)
	//	if err != nil {
	//		fmt.Println(err.Error())
	//		continue
	//	}
	//
	//}
	//
	//return bcs, nil
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
