package sns

import (
	"context"
	"fmt"
	"reflect"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sns/types"

	loafergo "github.com/justcodes/loafer-go/v2"
)

const (
	// DefaultMaxBatchSize is the default max batch size for sns
	DefaultMaxBatchSize = 10
)

// Producer represents loafer sns producer
type Producer interface {
	Produce(ctx context.Context, input *PublishInput) (string, error)
	ProduceBatch(ctx context.Context, input *PublishBatchInput) (*PublishBatchOutput, error)
}

// PublishBatchInput holds the sns batch publish attributes
type PublishBatchInput struct {
	Messages []*PublishBatchEntry
	TopicARN string
}

// PublishBatchEntry holds the sns batch publish attributes
// Each entry must have a unique ID to identify it in the request and response
type PublishBatchEntry struct {
	ID              string
	Message         string
	GroupID         string
	DeduplicationID string
	Attributes      map[string]string
}

// PublishBatchOutput holds the sns batch publish response
type PublishBatchOutput struct {
	Failed     []*PublishBatchEntryFailed
	Successful []*PublishBatchEntrySuccessful
}

// PublishBatchEntrySuccessful holds the sns batch publish response
type PublishBatchEntrySuccessful struct {
	EntryID   string
	MessageID string
}

// PublishBatchEntryFailed holds the sns batch publish response
type PublishBatchEntryFailed struct {
	EntryID string
	Err     error
}

// PublishInput has the sns event attributes
type PublishInput struct {
	Message         string
	GroupID         string
	DeduplicationID string
	TopicARN        string
	Attributes      map[string]string
}

type producer struct {
	sns loafergo.SNSClient
}

// NewProducer creates a new Producer
// It encapsulates the Amazon Simple Notification Service client
func NewProducer(config *Config) (Producer, error) {
	err := validateConfig(config)
	if err != nil {
		return nil, err
	}

	return &producer{
		sns: config.SNSClient,
	}, nil
}

// Produce publishes a message to an Amazon SNS topic. The message is then sent to all
// subscribers. When the topic is a FIFO topic, the message must also contain a group ID
// and, when ID-based deduplication is used, a deduplication ID. An optional key-value
// filter attribute can be specified so that the message can be filtered according to
// a filter policy.
func (p *producer) Produce(ctx context.Context, input *PublishInput) (string, error) {
	if input == nil || reflect.DeepEqual(input, &PublishInput{}) {
		return "", loafergo.ErrEmptyInput
	}

	pubInp := &sns.PublishInput{
		Message:   &input.Message,
		TargetArn: &input.TopicARN,
	}

	if input.GroupID != "" {
		pubInp.MessageGroupId = aws.String(input.GroupID)
	}

	if input.DeduplicationID != "" {
		pubInp.MessageDeduplicationId = aws.String(input.DeduplicationID)
	}

	if len(input.Attributes) > 0 {
		pubInp.MessageAttributes = p.messageAttributes(input.Attributes)
	}

	result, err := p.sns.Publish(ctx, pubInp)
	if err != nil {
		return "", fmt.Errorf("failed to publish message; topic: %s  error: %w", input.TopicARN, err)
	}

	return *result.MessageId, nil
}

// ProduceBatch publishes a batch of messages to an Amazon SNS topic. The messages are then sent to all
// subscribers. Each entry must have a unique ID to identify it in the request and response
// When the topic is a FIFO topic, the messages must also contain a group ID
// and, when ID-based deduplication is used, a deduplication ID. An optional key-value
// filter attribute can be specified so that the messages can be filtered according to
// a filter policy.
func (p *producer) ProduceBatch(ctx context.Context, input *PublishBatchInput) (*PublishBatchOutput, error) {
	if input == nil || reflect.DeepEqual(input, &PublishBatchInput{}) || len(input.Messages) == 0 {
		return nil, loafergo.ErrEmptyInput
	}

	if len(input.Messages) > DefaultMaxBatchSize {
		return nil, fmt.Errorf("maximum batch size is %d", DefaultMaxBatchSize)
	}

	pubInp := &sns.PublishBatchInput{
		TopicArn: aws.String(input.TopicARN),
	}

	for _, msg := range input.Messages {
		message := types.PublishBatchRequestEntry{
			Id:      aws.String(msg.ID),
			Message: aws.String(msg.Message),
		}

		if msg.GroupID != "" {
			message.MessageGroupId = aws.String(msg.GroupID)
		}

		if msg.DeduplicationID != "" {
			message.MessageDeduplicationId = aws.String(msg.DeduplicationID)
		}

		if len(msg.Attributes) > 0 {
			message.MessageAttributes = p.messageAttributes(msg.Attributes)
		}

		pubInp.PublishBatchRequestEntries = append(pubInp.PublishBatchRequestEntries, message)
	}

	result, err := p.sns.PublishBatch(ctx, pubInp)
	if err != nil {
		return nil, fmt.Errorf("failed to publish messages; topic: %s  error: %w", input.TopicARN, err)
	}

	return &PublishBatchOutput{
		Failed:     p.getFailedEntries(result.Failed),
		Successful: p.getSuccessfulEntries(result.Successful),
	}, nil
}

func (p *producer) getSuccessfulEntries(entries []types.PublishBatchResultEntry) []*PublishBatchEntrySuccessful {
	successful := make([]*PublishBatchEntrySuccessful, len(entries))
	for i, entry := range entries {
		successful[i] = &PublishBatchEntrySuccessful{
			EntryID:   *entry.Id,
			MessageID: *entry.MessageId,
		}
	}
	return successful
}

func (p *producer) getFailedEntries(entries []types.BatchResultErrorEntry) []*PublishBatchEntryFailed {
	failed := make([]*PublishBatchEntryFailed, len(entries))
	for i, entry := range entries {
		failed[i] = &PublishBatchEntryFailed{
			EntryID: *entry.Id,
			Err:     fmt.Errorf("failed to publish message; error: %s", *entry.Message),
		}
	}
	return failed
}

func (p *producer) messageAttributes(attr map[string]string) map[string]types.MessageAttributeValue {
	ma := make(map[string]types.MessageAttributeValue)
	for k, v := range attr {
		ma[k] = types.MessageAttributeValue{
			DataType:    aws.String("String"),
			StringValue: aws.String(v),
		}
	}
	return ma
}
