package main

import (
	"encoding/json"
	"errors"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/jimlawless/whereami"
	"net/http"
	"strconv"
)

const twillioAccountSid = "****************"
const twillioAuthToken = "****************"


func LambdaStart(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {


	if req.HTTPMethod == "GET" {

		broadcastID := req.QueryStringParameters["broadcast_id"]

		if broadcastID != "" {

			counts, err := getBroadcastStatusCounts(broadcastID)
			if err != nil {
				return serverError(err, whereami.WhereAmI())
			}

			data, err := json.Marshal(counts)
			if err != nil {
				return serverError(err, whereami.WhereAmI())
			}

			return events.APIGatewayProxyResponse{
				StatusCode: 200,
				Body:       string(data),
			}, nil

		} else {
			messages, err := getMessageIDs()
			if err != nil {
				return serverError(err, whereami.WhereAmI())
			}

			for _, message := range messages {
				//twillioMessageID := req.QueryStringParameters["message_id"]
				twillioMessageID := message.MessageID

				urlStr := "https://api.twilio.com/2010-04-01/Accounts/" + twillioAccountSid + "/Messages/" + twillioMessageID + ".json"

				// Create HTTP request client
				client := &http.Client{}
				req, _ := http.NewRequest("GET", urlStr, nil)
				req.SetBasicAuth(twillioAccountSid, twillioAuthToken)
				req.Header.Add("Accept", "application/json")

				// Make HTTP GET request and return message SID
				resp, err := client.Do(req)
				if err != nil {
					return serverError(err, whereami.WhereAmI())
				}

				var data map[string]interface{}
				decoder := json.NewDecoder(resp.Body)
				err = decoder.Decode(&data)
				if err != nil {
					return serverError(err, whereami.WhereAmI())
				}

				message.Status = data["status"].(string)

				err = putItem(&message)
				if err != nil {
					return serverError(err, whereami.WhereAmI())
				}

			}

			return events.APIGatewayProxyResponse{
				StatusCode: 200,
				Body:       "Updated " + strconv.Itoa(len(messages)) + " Messages!",
			}, nil
		}

	} else if req.HTTPMethod == "POST" {

		var d DeliveryStatus

		if err := json.Unmarshal([]byte(req.Body), &d); err != nil {
			return serverError(err,whereami.WhereAmI())
		}

		if d.MessageID == "" || d.BroadcastID == "" {
			return serverError(errors.New("Missing post body parameter"), whereami.WhereAmI())
		}

		d.Status = "None"

		err := putItem(&d)
		if err != nil {
			return serverError(err, whereami.WhereAmI())
		}

		return events.APIGatewayProxyResponse{
			StatusCode: 200,
			Body:       string("Success!"),
		}, nil


	}

	return serverError(errors.New("invalid request"), whereami.WhereAmI())

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
