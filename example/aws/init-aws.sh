#!/bin/bash

apt-get install jq -y

echo "-------------------------------------Init Script"
ENDPOINT="http://localhost:4566"
REGION="us-east-1"
PROFILE="test-profile"

echo "########### Creating profile ###########"

aws configure set aws_access_key_id dummy --profile $PROFILE
aws configure set aws_secret_access_key dummy --profile $PROFILE
aws configure set region $REGION --profile $PROFILE

echo "########### Listing profile ###########"
aws configure list --profile $PROFILE

echo "########### Creating SQS queues ###########"
EXAMPLE1_QUEUE_URL=$(aws --endpoint-url=$ENDPOINT sqs create-queue --queue-name example-1 --profile $PROFILE --region $REGION | jq -r '.QueueUrl')
EXAMPLE2_QUEUE_URL=$(aws --endpoint-url=$ENDPOINT sqs create-queue --queue-name example-2 --profile $PROFILE --region $REGION | jq -r '.QueueUrl')

echo "########### Send message to the queues ########### Creating SQS queues ###########"

send-message-to-sqs() {
	endPoint=$1
	key=$2
	queue=$3

	if [ -z "$endPoint" ] || [ -z "$key" ] || [ -z "$queue" ]; then
		echo "Queue URL, key and queue are required"
		return
	fi

	aws --endpoint-url=$ENDPOINT sqs send-message --queue-url "$endPoint" --message-body "{\"message\": \"Hello world! $queue $key\"}" --delay-seconds 0 --profile $PROFILE --region $REGION
}

for i in {1..20}
do
   	# do whatever on $i
	echo "Key : $i"
	echo "sending message to sqs queue: $EXAMPLE1_QUEUE_URL"
	send-message-to-sqs "$EXAMPLE1_QUEUE_URL" "$i" "example-1"
done

for i in {1..20}
do
   	# do whatever on $i
	echo "Key : $i"
	echo "sending message to sqs queue: $EXAMPLE2_QUEUE_URL"
	send-message-to-sqs "$EXAMPLE2_QUEUE_URL" "$i" "example-2"
done
