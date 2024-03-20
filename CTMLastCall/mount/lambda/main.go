package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/buger/jsonparser"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"fmt"
	"strings"
)

type SalesForce struct{}

var (
	//All
	baseUrl     = "https://*******.my.salesforce.com/services/"
	apiVersion  = "v44.0"
	redirectURI = "https://***.***.***.***/salesforce/"

	//getSalesForceToken()
	client_id     = "*********************************************************************"
	client_secret = "*********************************************************"
	username      = "sms_blast@******.com"
	password      = "************"
	securityToken = "******************" //can be reset in salesforce // only requeired on an ip not on whitelist

	//getSalesForceContact()
	token = "*********************************************************************"
	id    = "***************"

	//returned from SF
	dynamicAccessToken            = ""
	dynamicAccessTokenExpires     = 0
	dynamicAccessTokenLifespanHrs = 12
	salesforce                    SalesForce
)

func (salesforce *SalesForce) getSalesForceToken() {
	targetURL := baseUrl + "oauth2/token"
	data := url.Values{}
	data.Set("grant_type", "password")
	data.Set("client_id", client_id)
	data.Set("client_secret", client_secret)
	data.Set("username", username)
	data.Set("password", password+securityToken) // securityToken only requeired on an ip not on whitelist
	data.Set("redirect_uri", redirectURI)

	// Create a new request using http
	req, err := http.NewRequest("POST", targetURL, bytes.NewBufferString(data.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	// Send req using http Client
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		//util.logCodeError(whereami.WhereAmI(), err.Error())
	}

	body, _ := ioutil.ReadAll(resp.Body)

	// println("body")
	// fmt.Println(string(body))

	var dat map[string]interface{}
	if err := json.Unmarshal(body, &dat); err != nil {
		//util.logCodeError(whereami.WhereAmI(), err.Error())
	}

	dynamicAccessToken = dat["access_token"].(string)

	issued_at, err := strconv.Atoi(dat["issued_at"].(string))

	if err != nil {
		//util.logCodeError(whereami.WhereAmI(), err.Error())
	}

	dynamicAccessTokenExpires = issued_at/1000 + dynamicAccessTokenLifespanHrs*3600
}

func LambdaStart(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	if req.HTTPMethod == "POST" {
		if req.Body != "" {

			var ctmJson map[string]*json.RawMessage
			err := json.Unmarshal([]byte(req.Body), &ctmJson)
			if err != nil {
				return serverError(errors.New("failed to unmarshal json"))
			}

			if _, ok := ctmJson["caller_number"]; !ok {
				return serverError(errors.New("missing required param"))
			}

			var callerNum string
			err = json.Unmarshal(*ctmJson["caller_number"], &callerNum)
			if err != nil {
				return serverError(errors.New("failed to unmarshal caller json"))
			}

			if callerNum == "" {
				return serverError(errors.New("caller number missing"))
			}

			urlStr := "https://api.calltrackingmetrics.com/api/v1/accounts/*********/calls?auth_token=************&per_page=1&page=1&caller_number=" + strings.Replace(callerNum, "+", "", -1)

			// Create HTTP request client
			client := &http.Client{}
			req, _ := http.NewRequest("GET", urlStr, nil)
			req.Header.Add("Accept", "application/json")

			// Make HTTP GET request and return message SID
			resp, err := client.Do(req)
			if err != nil {
				return serverError(errors.New("ctm request failed"))
			}

			respBytes, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return serverError(errors.New("failed to read response body"))
			}

			//var cr CtmResponse
			//err = json.Unmarshal(respBytes, &cr)
			//if err != nil {
			//	return serverError(errors.New("failed to unmarshal response body"))
			//}
			callId, err := jsonparser.GetInt(respBytes, "calls", "[0]", "id")
			if err != nil {
				return serverError(errors.New("failed to unmarshal response body"))
			}

			fmt.Println(callId)

			callerNum = strings.Replace(callerNum, "+", "E", -1)

			salesforce.getSalesForceToken()

			var bearer = "Bearer " + dynamicAccessToken
			targetURL := baseUrl + "data/" + apiVersion + "/sobjects/Phone_Number__c/Phone_Number_ID__c/" + callerNum

			//
			var jsonStr = []byte(`{"Last_Call_ID__c":"` + strconv.Itoa(int(callId)) + `"}`)
			upsertReq, err := http.NewRequest("PATCH", targetURL, bytes.NewBuffer(jsonStr))
			if err != nil {
				return serverError(errors.New("failed to create new request"))
			}

			// add authorization header to the req
			upsertReq.Header.Add("Authorization", bearer)
			upsertReq.Header.Add("Content-Type", "application/json; charset=UTF-8")
			upsertReq.Header.Add("Accept", "application/json")


			sfClient := &http.Client{}
			sfResp, err := sfClient.Do(upsertReq)
			if err != nil {
				return serverError(errors.New("failed to complete request"))
			}

			sfRespBytes, err := ioutil.ReadAll(sfResp.Body)
			if err != nil {
				return serverError(errors.New("failed to read response body"))
			}

			// now find matching phone number on salesforce and upsert the URL and raw response right here
			return events.APIGatewayProxyResponse{
				StatusCode: 200,
				Body:       string(sfRespBytes),
			}, nil


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
