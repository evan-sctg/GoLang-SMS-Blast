package main

import (
	"errors"
	"net/http"
	"github.com/ajg/form"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"strings"
)
type TwilioStatus struct {
	MessageStatus string `form:"MessageStatus"`
	MessageSid string  `form:"MessageSid"`
}

type DeliveryStatus struct {
	MessageID 	string `json:"message_id"`
	Status 	  	string `json:"delivery_status"`
	BroadcastID string `json:"broadcast_id"`
}

func LambdaStart(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	if req.HTTPMethod == "POST" {
		if req.Body != "" {



			var PostData map[string]string
			
			


			// var status TwilioStatus

			 d := form.NewDecoder(strings.NewReader(req.Body))
			//  err := d.Decode(&status)
			 err := d.Decode(&PostData)
			 if err != nil {
			 	return serverError(err)
			 }

			 MessageStatus := PostData["MessageStatus"]
			MessageSid := PostData["MessageSid"]

			// messageStatus, err := getItem(status.MessageSid)
			messageStatus, err := getItem(MessageSid)
			if err != nil {
				return serverError(err)
			}
			if messageStatus == nil {
				// err = putItem(&DeliveryStatus{
				// 	Status: status.MessageStatus,
				// 	MessageID:  status.MessageSid,
				// 	BroadcastID: "system",
				// })
				err = putItem(&DeliveryStatus{
					Status: MessageStatus,
					MessageID:  MessageSid,
					BroadcastID: "system",
				})

				if err != nil {					
					return serverError(err)
				}
			} else {
				messageStatus.Status = MessageStatus
				err = putItem(messageStatus)
				if err != nil {
					return serverError(err)
				}
			}
			return success("OK")
		}
	}

	return serverError(errors.New("invalid request"))

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
