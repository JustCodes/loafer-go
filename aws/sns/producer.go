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

// Producer represents loafer sns producer
type Producer interface {
	Produce(ctx context.Context, input *PublishInput) (string, error)
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
