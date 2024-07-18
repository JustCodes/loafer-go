package loafergo

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

// Router holds the Route methods to configure and run
type Router interface {
	Configure(ctx context.Context, c SQSClient, l Logger) error
	Run(ctx context.Context, workerPool int)
}

// SQSClient represents the aws sqs client methods
type SQSClient interface {
	ChangeMessageVisibility(ctx context.Context, params *sqs.ChangeMessageVisibilityInput, optFns ...func(*sqs.Options)) (*sqs.ChangeMessageVisibilityOutput, error)
	GetQueueUrl(ctx context.Context, params *sqs.GetQueueUrlInput, optFns ...func(*sqs.Options)) (*sqs.GetQueueUrlOutput, error)
	ReceiveMessage(ctx context.Context, params *sqs.ReceiveMessageInput, optFns ...func(*sqs.Options)) (*sqs.ReceiveMessageOutput, error)
	DeleteMessage(ctx context.Context, params *sqs.DeleteMessageInput, optFns ...func(*sqs.Options)) (*sqs.DeleteMessageOutput, error)
}
