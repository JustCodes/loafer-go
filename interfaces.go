package loafergo

import (
	"context"
	"time"

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
	VisibilityTimeout(ctx context.Context) int32
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
	// Decode will unmarshal the body message into a supplied output using JSON
	Decode(out interface{}) error
	// Attribute will return the custom attribute sent throughout the request.
	Attribute(key string) string
	// Attributes will return the custom attributes sent with the request.
	Attributes() map[string]string
	// SystemAttributeByKey will return the system attributes by key.
	SystemAttributeByKey(key string) string
	// SystemAttributes will return the system attributes.
	SystemAttributes() map[string]string
	// Metadata will return the metadata sent throughout the request.
	Metadata() map[string]string
	// Identifier will return an identifier associated with the message ReceiptHandle.
	Identifier() string
	// Dispatch used to dispatch a message if necessary
	Dispatch()
	// Backoff used to change the visibilityTimeout of the message
	// when a message is backoff it will not be removed from the queue
	// instead it will extend the visibility timeout of the message
	Backoff(delay time.Duration)
	// BackedOff used to check if the message was backedOff by the handler
	BackedOff() bool
	// Body used to get the message Body
	Body() []byte
	// Message returns the body message
	Message() string
	// TimeStamp returns the message timestamp
	TimeStamp() time.Time
	// DecodeMessage will unmarshal the message into a supplied output using JSON
	DecodeMessage(out any) error
}

// SNSClient represents the aws sns client methods
type SNSClient interface {
	Publish(ctx context.Context, params *sns.PublishInput, optFns ...func(*sns.Options)) (*sns.PublishOutput, error)
	PublishBatch(ctx context.Context, params *sns.PublishBatchInput, optFns ...func(*sns.Options)) (*sns.PublishBatchOutput, error)
}
