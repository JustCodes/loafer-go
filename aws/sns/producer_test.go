package sns_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsSNS "github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sns/types"
	"github.com/stretchr/testify/suite"

	loafergo "github.com/justcodes/loafer-go/v2"
	"github.com/justcodes/loafer-go/v2/aws/sns"
	"github.com/justcodes/loafer-go/v2/fake"
)

type producerSuite struct {
	suite.Suite
	snsCLient *fake.SNSClient
	producer  sns.Producer
}

func TestProducerSuite(t *testing.T) {
	suite.Run(t, new(producerSuite))
}

func (suite *producerSuite) SetupSuite() {
	suite.snsCLient = fake.NewSNSClient(suite.T())
	suite.producer, _ = sns.NewProducer(&sns.Config{
		SNSClient: suite.snsCLient,
	})
}

func (suite *producerSuite) TearDownSuite() {
	suite.SetupSuite()
}

func (suite *producerSuite) TestNewProducerWithError() {
	suite.Run("With Config nil", func() {
		p, err := sns.NewProducer(nil)
		suite.Nil(p)
		suite.NotNil(err)
	})

	suite.Run("With Client nil", func() {
		p, err := sns.NewProducer(&sns.Config{})
		suite.Nil(p)
		suite.NotNil(err)
	})
}

func (suite *producerSuite) TestPublish() {
	ctx := context.Background()
	suite.Run("Should produce with Success", func() {
		topicArn, err := sns.BuildTopicARN("us-east-1", "0000000", "my_topic")
		suite.NoError(err)

		input := sns.PublishInput{
			Message:  "my message",
			TopicARN: topicArn,
		}

		param := &awsSNS.PublishInput{
			Message:   &input.Message,
			TargetArn: &input.TopicARN,
		}

		rID := "id"

		suite.snsCLient.On("Publish", ctx, param).
			Return(
				&awsSNS.PublishOutput{
					MessageId: &rID,
				},
				nil,
			).
			Once()

		got, err := suite.producer.Produce(ctx, &input)
		suite.NoError(err)
		suite.Equal("id", got)

	})

	suite.Run("Should produce with message attributes", func() {
		topicArn, err := sns.BuildTopicARN("us-east-1", "0000000", "my_topic")
		suite.NoError(err)

		input := sns.PublishInput{
			Message:  "my message",
			TopicARN: topicArn,
			Attributes: map[string]string{
				"custom": "custom_value",
			},
		}

		param := &awsSNS.PublishInput{
			Message:   &input.Message,
			TargetArn: &input.TopicARN,
			MessageAttributes: map[string]types.MessageAttributeValue{
				"custom": {DataType: aws.String("String"), StringValue: aws.String("custom_value")},
			},
		}

		rID := "id"

		suite.snsCLient.On("Publish", ctx, param).
			Return(
				&awsSNS.PublishOutput{
					MessageId: &rID,
				},
				nil,
			).
			Once()

		got, err := suite.producer.Produce(ctx, &input)
		suite.NoError(err)
		suite.Equal("id", got)

	})
	suite.Run("Should produce with GroupID", func() {
		topicArn, err := sns.BuildTopicARN("us-east-1", "0000000", "my_topic")
		suite.NoError(err)

		input := sns.PublishInput{
			Message:  "my message",
			GroupID:  "my-group",
			TopicARN: topicArn,
		}

		param := &awsSNS.PublishInput{
			Message:        &input.Message,
			TargetArn:      &input.TopicARN,
			MessageGroupId: aws.String("my-group"),
		}

		rID := "id"

		suite.snsCLient.On("Publish", ctx, param).
			Return(
				&awsSNS.PublishOutput{
					MessageId: &rID,
				},
				nil,
			).
			Once()

		got, err := suite.producer.Produce(ctx, &input)
		suite.NoError(err)
		suite.Equal("id", got)

	})
	suite.Run("Should produce with all fields", func() {
		topicArn, err := sns.BuildTopicARN("us-east-1", "0000000", "my_topic")
		suite.NoError(err)

		input := sns.PublishInput{
			Message:         "my message",
			GroupID:         "my-group",
			DeduplicationID: "dedup-id",
			TopicARN:        topicArn,
			Attributes: map[string]string{
				"custom": "custom_value",
			},
		}

		param := &awsSNS.PublishInput{
			Message:   &input.Message,
			TargetArn: &input.TopicARN,
			MessageAttributes: map[string]types.MessageAttributeValue{
				"custom": {DataType: aws.String("String"), StringValue: aws.String("custom_value")},
			},
			MessageGroupId:         aws.String("my-group"),
			MessageDeduplicationId: aws.String("dedup-id"),
		}

		rID := "id"

		suite.snsCLient.On("Publish", ctx, param).
			Return(
				&awsSNS.PublishOutput{
					MessageId: &rID,
				},
				nil,
			).
			Once()

		got, err := suite.producer.Produce(ctx, &input)
		suite.NoError(err)
		suite.Equal("id", got)

	})

	suite.Run("SNS Publish error", func() {
		topicArn, err := sns.BuildTopicARN("us-east-1", "0000000", "my_topic")
		suite.NoError(err)

		input := sns.PublishInput{
			Message:         "my message",
			GroupID:         "my-group",
			DeduplicationID: "dedup-id",
			TopicARN:        topicArn,
			Attributes: map[string]string{
				"custom": "custom_value",
			},
		}

		param := &awsSNS.PublishInput{
			Message:   &input.Message,
			TargetArn: &input.TopicARN,
			MessageAttributes: map[string]types.MessageAttributeValue{
				"custom": {DataType: aws.String("String"), StringValue: aws.String("custom_value")},
			},
			MessageGroupId:         aws.String("my-group"),
			MessageDeduplicationId: aws.String("dedup-id"),
		}

		suite.snsCLient.On("Publish", ctx, param).
			Return(nil, fmt.Errorf("got error")).
			Once()

		got, err := suite.producer.Produce(ctx, &input)
		suite.Empty(got)
		suite.NotNil(err)
		suite.ErrorContains(err, "got error")

	})

	suite.Run("With Input nil", func() {
		got, err := suite.producer.Produce(ctx, nil)
		suite.Empty(got)
		suite.NotNil(err)
		suite.ErrorIs(err, loafergo.ErrEmptyInput)
	})
}
