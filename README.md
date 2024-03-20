# GoLang-SMS-Blast
Ultra-fast shortcode SMS blasts via serverless AWS &amp; Twilio. Real-time Salesforce, NoSQL DB, opt-out &amp; DNC written in GoLang

# Ultra-scalable SMS Lead Gen for Call Centers
This project provides a highly scalable SMS blast solution built with GoLang and serverless AWS services. It integrates with Twilio for message sending, DynamoDB for NoSQL data storage, Salesforce for real-time data synchronization, and utilizes AWS Lambda functions for efficient message processing.

# Features:

- Ultra-fast Shortcode SMS Blasts: Send SMS messages to large contact lists using shortcodes or traditional longcode phone numbers.
- Real-time Salesforce Integration: Synchronize data between Salesforce and your SMS campaigns for seamless lead management.
- Do Not Call (DNC) List Management: Ensure compliance by excluding DNC numbers from your blasts.
- Opt-out Management: Allow recipients to opt out of future messages.
- Serverless Scalability: Leverage AWS Lambda functions for automatic scaling based on traffic.
- NoSQL Database (DynamoDB): Efficiently store and manage large datasets of contact information.


# SMS-Blast lambdas
## This project utilizes a variety of GoLang Lambda functions to manage different aspects of the SMS blast workflow:

- AddToSMSQueue: Adds messages to the SMS queue for processing.
- AuroraDBContactListByBroadcastid: Retrieves contacts from AuroraDB for a specific broadcast.
- BatchDNCStep: Performs batch checks against the Do Not Call (DNC) list.
- BlockedShortCodes: Manages a list of blocked shortcodes or longcodes.
- Broadcasts: Manages and stores information about broadcasts.
- BroadcastStep: Processes a step within a broadcast.
- CTMLastCall: Tracks the last call made for a specific contact.
- DeliveryStatusTotals: Maintains tallies for message delivery statuses.
- DNC: Handles Do Not Call list management.
- ImportDNC: Imports a Do Not Call list.
- MessageService: Manages and stores phone numbers and SIDs stored in DynamoDB.
- ReceieveTwilioSMS: Receives incoming SMS messages from Twilio.
- ReceieveTwilioSMSNew: Sends outgoing SMS messages using Twilio.
- ReceiveTwilioStatuses: Receives and processes delivery status updates from Twilio.
- RetrieveSMSQueue: Retrieves messages from the SMS queue for processing.
- S3TOSQS: Moves messages from Amazon S3 to Amazon SQS.
- SMSDeliveryStatus: Updates the delivery status of messages.
- SMSTODNC: Adds phone numbers to the Do Not Call list.
- StateMachines: Manages state machines for orchestrating the SMS blast workflow (example provided below).

## This project utilizes Docker Compose for managing dependencies.  Ensure you have Docker installed and running on your system.

Clone the repository:
```
git clone https://github.com/evan-sctg/GoLang-SMS-Blast.git
```

Navigate to the project directory:
```
cd GoLang-SMS-Blast
```

## Building Lambdas
#### ***Create each Lambda in the project and connect to Step Functions, SQS, S3, DynamoDB, Aurora or SQL
#### ***Replace shortcode or longcode sending phone numbers, twillio, salesforce connection details 
Navigate to each Lambda directory:
```
cd AddToSMSQueue
```

start container:
```
docker-compose up -d
```

(Optional) Run the GoLang lambda functions inside the appropriate container for development purposes:
```
docker exec -it go_lambda_dynamodb /bin/bash
```




# State Machine for Sending SMS Blasts:

## The project utilizes a Step Functions state machine to orchestrate the SMS blast workflow. Here's a breakdown of the provided state machine definition:

```
{
  "StartAt": "BroadcastStep",
  "States": {
    "BroadcastStep": {
      "Type": "Task",
      "Resource": "arn:aws:lambda:us-east-1:**********:function:BroadcastStep",
      "Next": "BroadcastsFound"
    },
    "BroadcastsFound": {
      "Type": "Choice",
      "Choices": [
        {
          "Not": {
            "Variable": "$.uuid",
            "StringEquals": ""
          },
          "Next": "SelectContacts"
        }
      ],
      "Default": "NoBroadcasts"
    },
    "SelectContacts": {
      "Type": "Task",
      "Resource": "arn:aws:lambda:us-east-1:**********:function:AuroraDBContactListByBroadcastId",
      "Next": "BlockedShortcodes"
    },
    "BlockedShortcodes": {
      "Type": "Task",
      "Resource": "arn:aws:lambda:us-east-1:**********:function:BlockedShortCodes",
      "Next": "BatchDNCStep"
    },
    "BatchDNCStep": {
      "Type": "Task",
      "Resource": "arn:aws:lambda:us-east-1:**********:function:BatchDNCStep",
      "Next": "S3ToSQS"
    },
    "S3ToSQS": {
      "Type": "Task",
      "Resource": "arn:aws:lambda:us-east-1:**********:function:S3ToSQS",
      "Next": "CheckBroadcastsStep"
    },
    "CheckBroadcastsStep": {
      "Type": "Task",
      "Resource": "arn:aws:lambda:us-east-1:**********:function:BroadcastStep",
      "Next": "BroadcastsFound"
    },
```
