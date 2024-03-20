package main

import (
	"encoding/json"
	"errors"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func LambdaStart(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	if req.HTTPMethod == "GET" {

		ID := req.QueryStringParameters["id"]
		unixTime := req.QueryStringParameters["unix_time"]

		startID := req.QueryStringParameters["start_id"]
		startUnixTime := req.QueryStringParameters["start_unix_time"]
		startYearMonth := req.QueryStringParameters["start_year_month"]

		startSentTime := req.QueryStringParameters["start_sent_time"]
		endSentTime := req.QueryStringParameters["end_sent_time"]

		searchQuery := req.QueryStringParameters["search"]

		count := req.QueryStringParameters["count"]
		getCount := false
		if count != "" {
			getCount = true
		}

		limit := req.QueryStringParameters["limit"]
		if limit == "" {
			limit = "20"
		}

		sort := req.QueryStringParameters["sort"]
		if sort == "" {
			sort = "desc"
		}

		lim, _ := strconv.Atoi(limit)
		var sortForward bool
		if strings.ToLower(sort) == "desc" {
			sortForward = false
		} else {
			sortForward = true
		}


		if ID != "" {
			// get one by id
			b, err := getItem(ID)
			if err != nil {
				return serverError(err)
			}

			res, err := json.Marshal(b)
			if err != nil {
				return serverError(err)
			}

			return events.APIGatewayProxyResponse{
				StatusCode: 200,
				Body:       string(res),
			}, nil
		} else if startSentTime != "" && endSentTime != "" {

			utInt, err := strconv.ParseInt(startSentTime, 10, 64)
			if err != nil {
				return serverError(err)
			}
			date := time.Unix(utInt, 0)

			yearMonth := strconv.Itoa(date.Year()) + "_" + date.Month().String()

			b, err := queryBetween(startSentTime, endSentTime, yearMonth, int64(lim), getCount)
			if err != nil {
				return serverError(err)
			}

			res, err := json.Marshal(b)
			if err != nil {
				return serverError(err)
			}

			return events.APIGatewayProxyResponse{
				StatusCode: 200,
				Body:       string(res),
			}, nil

		} else if unixTime != "" {

			utInt, err := strconv.ParseInt(unixTime, 10, 64)
			if err != nil {
				return serverError(err)
			}
			date := time.Unix(utInt, 0)

			yearMonth := strconv.Itoa(date.Year()) + "_" + date.Month().String()

			// get list by unix time return 10 at a time
			if startID != "" && startUnixTime != "" && startYearMonth != ""{

				startKey := map[string]*dynamodb.AttributeValue{
					"id" : {
						S: aws.String(startID),
					},
					"unix_time": {
						N: aws.String(startUnixTime),
					},
					"year_month": {
						S: aws.String(startYearMonth),
					},
				}
				if searchQuery != "" {
					b, err := searchItems(unixTime, yearMonth, searchQuery, int64(lim), sortForward, getCount, startKey)
					if err != nil {
						return serverError(err)
					}

					res, err := json.Marshal(b)
					if err != nil {
						return serverError(err)
					}

					return events.APIGatewayProxyResponse{
						StatusCode: 200,
						Body:       string(res),
					}, nil
				}

				b, err := queryItems(unixTime, yearMonth, int64(lim), sortForward, getCount, startKey)
				if err != nil {
					return serverError(err)
				}

				res, err := json.Marshal(b)
				if err != nil {
					return serverError(err)
				}

				return events.APIGatewayProxyResponse{
					StatusCode: 200,
					Body:       string(res),
				}, nil


			} else {
				if searchQuery != "" {
					b, err := searchItems(unixTime, yearMonth, searchQuery, int64(lim), sortForward, getCount)
					if err != nil {
						return serverError(err)
					}

					res, err := json.Marshal(b)
					if err != nil {
						return serverError(err)
					}

					return events.APIGatewayProxyResponse{
						StatusCode: 200,
						Body:       string(res),
					}, nil
				}

				b, err := queryItems(unixTime, yearMonth, int64(lim), sortForward, getCount)
				if err != nil {
					return serverError(err)
				}

				res, err := json.Marshal(b)
				if err != nil {
					return serverError(err)
				}

				return events.APIGatewayProxyResponse{
					StatusCode: 200,
					Body:       string(res),
				}, nil
			}


		} else {
			return serverError(errors.New("a required field is missing"))
		}

	} else if req.HTTPMethod == "POST" {

		if req.Body != "" {
			var PostData Broadcast
			if err := json.Unmarshal([]byte(req.Body), &PostData); err != nil {
				return serverError(err)
			}

			timeNow := time.Now().UTC()

			PostData.Processed = "false"
			PostData.UnixTime = timeNow.Unix()
			PostData.YearMonth = strconv.Itoa(timeNow.Year()) + "_" + timeNow.Month().String()
			PostData.Successes = "0"
			PostData.Failures = "0"
			PostData.Sent = "0"

			if PostData.ID == "" || PostData.Name == "" || PostData.SendTime == 0 || PostData.Message == "" {
				return serverError(errors.New("a required field is missing"))
			}

			output, err := putItem(&PostData)
			if err != nil {
				return serverError(err)
			}

			res, err := json.Marshal(output)
			if err != nil {
				return serverError(err)
			}

			return events.APIGatewayProxyResponse{
				StatusCode: 200,
				Body:       string(res),
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
