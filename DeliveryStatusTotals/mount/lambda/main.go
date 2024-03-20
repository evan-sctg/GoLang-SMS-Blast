package main

import (
	"errors"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/jimlawless/whereami"
	"net/http"
	"strconv"
	"time"
)

const twillioAccountSid = "******************"
const twillioAuthToken = "******************"


func LambdaStart(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	if req.HTTPMethod == "GET" {
		broadcastID := req.QueryStringParameters["broadcast_id"]

		if broadcastID != "" {
			broadcast, err := getItem(broadcastID)
			if err != nil {
				return serverError(err, whereami.WhereAmI())
			}

			count, err := getBroadcastStatusCounts(broadcastID)
			if err != nil {
				return serverError(err, whereami.WhereAmI())
			}

			broadcast.Successes = strconv.Itoa(int(count.DeliveredCount))
			broadcast.Failures = strconv.Itoa(int(count.UndeliveredCount))
			broadcast.Sent = strconv.Itoa(int(count.UndeliveredCount + count.DeliveredCount + count.NoStatusCount))
			err = putItem(broadcast)
			return success("updated broadcast")
		} else {
			timeNow := time.Now()

			dayStart := time.Date(timeNow.Year(), timeNow.Month(), timeNow.Day(), 0, 0, 0, 0, timeNow.Location()).Unix()
			dayEnd := dayStart + 86400

			yearMonth := strconv.Itoa(timeNow.Year()) + "_" + timeNow.Month().String()

			bcs, err := queryBetween(strconv.Itoa(int(dayStart)), strconv.Itoa(int(dayEnd)), yearMonth)
			if err != nil {
				return serverError(err, whereami.WhereAmI())
			}
			for _, broadcast := range *bcs {
				// we gotta do some stuff here for sure
				count, err := getBroadcastStatusCounts(broadcast.ID)
				if err != nil {
					return serverError(err, whereami.WhereAmI())
				}
				broadcast.Successes = strconv.Itoa(int(count.DeliveredCount))
				broadcast.Failures = strconv.Itoa(int(count.UndeliveredCount))
				broadcast.Sent = strconv.Itoa(int(count.UndeliveredCount + count.DeliveredCount + count.NoStatusCount))
				err = putItem(&broadcast)
				if err != nil {
					return serverError(err, whereami.WhereAmI())
				}
			}

			return success("updated")
		}

	}

	return serverError(errors.New("invalid request"), whereami.WhereAmI())
}

func success(resp string) (events.APIGatewayProxyResponse, error) {
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       resp,
	}, nil
}

// Add a helper for handling errors.
// returns a 500 Internal Server Error response that the AWS API
// Gateway understands.
func serverError(err error, whereAmI string) (events.APIGatewayProxyResponse, error) {

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusInternalServerError,
		Body:       http.StatusText(http.StatusInternalServerError) + " [" + err.Error() + " | " + whereAmI + "]",
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
