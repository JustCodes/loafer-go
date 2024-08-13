package loafergo

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

// Router holds the Route methods to configure and run
type Router interface {
	Configure(ctx context.Context) error
	GetMessages(ctx context.Context) ([]Message, error)
	HandlerMessage(ctx context.Context, msg Message) error
	Commit(ctx context.Context, m Message) error
	WorkerPoolSize(ctx context.Context) int32
}

// SQSClient represents the aws sqs client methods
type SQSClient interface {
	ChangeMessageVisibility(
		ctx context.Context,
		params *sqs.ChangeMessageVisibilityInput,
		optFns ...func(*sqs.Options)) (*sqs.ChangeMessageVisibilityOutput, error)
	GetQueueUrl(ctx context.Context, params *sqs.GetQueueUrlInput, optFns ...func(*sqs.Options)) (*sqs.GetQueueUrlOutput, error)
	ReceiveMessage(ctx context.Context, params *sqs.ReceiveMessageInput, optFns ...func(*sqs.Options)) (*sqs.ReceiveMessageOutput, error)
	DeleteMessage(ctx context.Context, params *sqs.DeleteMessageInput, optFns ...func(*sqs.Options)) (*sqs.DeleteMessageOutput, error)
}

// Message represents the message interface methods
type Message interface {
	// Decode will unmarshal the message into a supplied output using json
	Decode(out interface{}) error
	// Attribute will return the custom attribute that was sent throughout the request.
	Attribute(key string) string
	// Metadata will return the metadata that was sent throughout the request.
	Metadata() map[string]string
	// Identifier will return a message identifier
	Identifier() string
	// Dispatch used to dispatch message if necessary
	Dispatch()
	// Body used to get the message Body
	Body() []byte
}

// SNSClient represents the aws sns client methods
type SNSClient interface {
	Publish(ctx context.Context, params *sns.PublishInput, optFns ...func(*sns.Options)) (*sns.PublishOutput, error)
	PublishBatch(ctx context.Context, params *sns.PublishBatchInput, optFns ...func(*sns.Options)) (*sns.PublishBatchOutput, error)
}
