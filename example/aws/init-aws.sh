#!/bin/bash

apt-get install jq -y

echo "-------------------------------------Init Script"
ENDPOINT="http://localhost:4566"
REGION="us-east-1"
PROFILE="test-profile"
TOPIC_ONE="my_topic__test"
TOPIC_TWO="my_topic__test2"
QUEUE_ONE="example-1"
QUEUE_TWO="example-2"

echo "########### Creating profile ###########"

aws configure set aws_access_key_id dummy --profile $PROFILE
aws configure set aws_secret_access_key dummy --profile $PROFILE
aws configure set region $REGION --profile $PROFILE

echo "########### Listing profile ###########"
aws configure list --profile $PROFILE

echo "########### Creating SNS  topics ###########"
aws --endpoint-url=$ENDPOINT sns create-topic --name $TOPIC_ONE --profile $PROFILE --region $REGION --output table | cat

aws --endpoint-url=$ENDPOINT sns create-topic --name $TOPIC_TWO --profile $PROFILE --region $REGION --output table | cat

echo "########### Creating SQS queues ###########"
aws --endpoint-url=$ENDPOINT sqs create-queue --queue-name example-1 --profile $PROFILE --region $REGION --output table | cat

aws --endpoint-url=$ENDPOINT sqs create-queue --queue-name example-2 --profile $PROFILE --region $REGION --output table | cat

echo "########### Subscribing the topics"
aws --endpoint-url=$ENDPOINT sns subscribe --topic-arn arn:aws:sns:us-east-1:000000000000:$TOPIC_ONE --protocol sqs --notification-endpoint arn:aws:sqs:us-east-1:000000000000:$QUEUE_ONE --profile $PROFILE --region $REGION --output table | cat

aws --endpoint-url=$ENDPOINT sns subscribe --topic-arn arn:aws:sns:us-east-1:000000000000:$TOPIC_TWO --protocol sqs --notification-endpoint arn:aws:sqs:us-east-1:000000000000:$QUEUE_TWO --profile $PROFILE --region $REGION --output table | cat
