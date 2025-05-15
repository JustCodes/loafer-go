#!/bin/bash
set -euo pipefail

# Constants
ENDPOINT="http://localhost:4566"
REGION="us-east-1"
PROFILE="test-profile"
TOPIC_ONE="my_topic__test"
TOPIC_ONE_FIFO="my_topic__test_f.fifo"
TOPIC_TWO="my_topic__test2"
QUEUE_ONE="example-1"
QUEUE_TWO="example-2"
QUEUE_FIFO="example.fifo"

echo "🚀 Initializing LocalStack AWS resources..."
echo "------------------------------------------"

# Ensure dependencies
if ! command -v jq &> /dev/null; then
  echo "Installing jq..."
  apt-get update && apt-get install -y jq
fi

echo "🔐 Configuring AWS CLI profile: $PROFILE"
aws configure set aws_access_key_id dummy --profile "$PROFILE"
aws configure set aws_secret_access_key dummy --profile "$PROFILE"
aws configure set region "$REGION" --profile "$PROFILE"

echo "📋 AWS CLI Profile Details:"
aws configure list --profile "$PROFILE"

echo "📣 Creating SNS Topics..."
TOPIC_ONE_FIFO_ARN=$(aws --endpoint-url="$ENDPOINT" sns create-topic \
  --name "$TOPIC_ONE_FIFO" \
  --attributes "FifoTopic=true,ContentBasedDeduplication=true,DisplayName=Auto Insurance Claims Topic" \
  --region "$REGION" \
  --profile "$PROFILE" | jq -r '.TopicArn')

aws --endpoint-url="$ENDPOINT" sns create-topic --name "$TOPIC_ONE" --profile "$PROFILE" --region "$REGION" --output table
aws --endpoint-url="$ENDPOINT" sns create-topic --name "$TOPIC_TWO" --profile "$PROFILE" --region "$REGION" --output table

echo "📬 Creating SQS Queues..."
aws --endpoint-url="$ENDPOINT" sqs create-queue --queue-name "$QUEUE_ONE" --profile "$PROFILE" --region "$REGION" --output table
aws --endpoint-url="$ENDPOINT" sqs create-queue --queue-name "$QUEUE_TWO" --profile "$PROFILE" --region "$REGION" --output table

QUEUE_FIFO_URL=$(aws --endpoint-url="$ENDPOINT" sqs create-queue \
  --queue-name "$QUEUE_FIFO" \
  --attributes "FifoQueue=true" \
  --profile "$PROFILE" \
  --region "$REGION" | jq -r '.QueueUrl')

QUEUE_FIFO_ARN=$(aws --endpoint-url="$ENDPOINT" sqs get-queue-attributes \
  --queue-url "$QUEUE_FIFO_URL" \
  --attribute-names QueueArn \
  --profile "$PROFILE" --region "$REGION" | jq -r '.Attributes.QueueArn')

echo "📎 FIFO Queue URL: $QUEUE_FIFO_URL"
echo "📎 FIFO Queue ARN: $QUEUE_FIFO_ARN"

echo "🔗 Subscribing SQS Queues to SNS Topics..."
aws --endpoint-url="$ENDPOINT" sns subscribe \
  --topic-arn "arn:aws:sns:$REGION:000000000000:$TOPIC_ONE" \
  --protocol sqs \
  --notification-endpoint "arn:aws:sqs:$REGION:000000000000:$QUEUE_ONE" \
  --profile "$PROFILE" \
  --region "$REGION" \
  --output table

aws --endpoint-url="$ENDPOINT" sns subscribe \
  --topic-arn "arn:aws:sns:$REGION:000000000000:$TOPIC_TWO" \
  --protocol sqs \
  --notification-endpoint "arn:aws:sqs:$REGION:000000000000:$QUEUE_TWO" \
  --profile "$PROFILE" \
  --region "$REGION" \
  --output table

aws --endpoint-url="$ENDPOINT" sns subscribe \
  --topic-arn "$TOPIC_ONE_FIFO_ARN" \
  --protocol sqs \
  --notification-endpoint "$QUEUE_FIFO_ARN" \
  --profile "$PROFILE" \
  --region "$REGION" \
  --output table

echo "✅ Setup complete!"
