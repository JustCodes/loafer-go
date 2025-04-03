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

func (suite *producerSuite) TestPublishBatch() {
	ctx := context.Background()
	suite.Run("Should produce with Success", func() {
		topicArn, err := sns.BuildTopicARN("us-east-1", "0000000", "my_topic")
		suite.NoError(err)

		input := sns.PublishBatchInput{
			Messages: []*sns.PublishBatchEntry{
				{
					ID:      "id",
					Message: "my message",
				},
			},
			TopicARN: topicArn,
		}

		param := &awsSNS.PublishBatchInput{
			TopicArn: aws.String(input.TopicARN),
			PublishBatchRequestEntries: []types.PublishBatchRequestEntry{
				{
					Id:      aws.String(input.Messages[0].ID),
					Message: aws.String(input.Messages[0].Message),
				},
			},
		}

		rID := "id"

		suite.snsCLient.On("PublishBatch", ctx, param).
			Return(
				&awsSNS.PublishBatchOutput{
					Successful: []types.PublishBatchResultEntry{
						{
							Id:        aws.String(input.Messages[0].ID),
							MessageId: aws.String(rID),
						},
					},
				},
				nil,
			).
			Once()

		got, err := suite.producer.ProduceBatch(ctx, &input)
		suite.NoError(err)
		suite.Equal("id", got.Successful[0].MessageID)

	})

	suite.Run("Should produce many messages", func() {
		topicArn, err := sns.BuildTopicARN("us-east-1", "0000000", "my_topic")
		suite.NoError(err)

		input := sns.PublishBatchInput{
			Messages: []*sns.PublishBatchEntry{
				{
					ID:      "id1",
					Message: "my message 1",
				},
				{
					ID:      "id2",
					Message: "my message 2",
				},
				{
					ID:      "id3",
					Message: "my message 3",
				},
			},
			TopicARN: topicArn,
		}
		param := &awsSNS.PublishBatchInput{
			TopicArn: aws.String(input.TopicARN),
			PublishBatchRequestEntries: []types.PublishBatchRequestEntry{
				{
					Id:      aws.String(input.Messages[0].ID),
					Message: aws.String(input.Messages[0].Message),
				},
				{
					Id:      aws.String(input.Messages[1].ID),
					Message: aws.String(input.Messages[1].Message),
				},
				{
					Id:      aws.String(input.Messages[2].ID),
					Message: aws.String(input.Messages[2].Message),
				},
			},
		}
		rID1 := "id1"
		rID2 := "id2"
		rID3 := "id3"

		suite.snsCLient.On("PublishBatch", ctx, param).
			Return(
				&awsSNS.PublishBatchOutput{
					Successful: []types.PublishBatchResultEntry{
						{
							Id:        aws.String(input.Messages[0].ID),
							MessageId: aws.String(rID1),
						},
						{
							Id:        aws.String(input.Messages[1].ID),
							MessageId: aws.String(rID2),
						},
						{
							Id:        aws.String(input.Messages[2].ID),
							MessageId: aws.String(rID3),
						},
					},
				},
				nil,
			).
			Once()
		got, err := suite.producer.ProduceBatch(ctx, &input)
		suite.NoError(err)
		suite.Equal("id1", got.Successful[0].MessageID)
		suite.Equal("id2", got.Successful[1].MessageID)
		suite.Equal("id3", got.Successful[2].MessageID)
	})

	suite.Run("Should produce and all messages fail", func() {
		topicArn, err := sns.BuildTopicARN("us-east-1", "0000000", "my_topic")
		suite.NoError(err)
		input := sns.PublishBatchInput{
			Messages: []*sns.PublishBatchEntry{
				{
					ID:      "id1",
					Message: "my message 1",
				},
				{
					ID:      "id2",
					Message: "my message 2",
				},
			},
			TopicARN: topicArn,
		}

		param := &awsSNS.PublishBatchInput{
			TopicArn: aws.String(input.TopicARN),
			PublishBatchRequestEntries: []types.PublishBatchRequestEntry{
				{
					Id:      aws.String(input.Messages[0].ID),
					Message: aws.String(input.Messages[0].Message),
				},
				{
					Id:      aws.String(input.Messages[1].ID),
					Message: aws.String(input.Messages[1].Message),
				},
			},
		}

		suite.snsCLient.On("PublishBatch", ctx, param).
			Return(
				&awsSNS.PublishBatchOutput{
					Failed: []types.BatchResultErrorEntry{
						{
							Id:      aws.String(input.Messages[0].ID),
							Message: aws.String("error"),
						},
						{
							Id:      aws.String(input.Messages[1].ID),
							Message: aws.String("error"),
						},
					},
				}, nil).
			Once()
		got, err := suite.producer.ProduceBatch(ctx, &input)
		suite.NoError(err)
		suite.Equal(2, len(got.Failed))
		suite.Equal("id1", got.Failed[0].EntryID)
		suite.Equal("failed to publish message; error: error", got.Failed[0].Err.Error())
		suite.Equal("id2", got.Failed[1].EntryID)
		suite.Equal("failed to publish message; error: error", got.Failed[1].Err.Error())
	})

	suite.Run("Should produce and some messages fail", func() {
		topicArn, err := sns.BuildTopicARN("us-east-1", "0000000", "my_topic")
		suite.NoError(err)
		input := sns.PublishBatchInput{
			Messages: []*sns.PublishBatchEntry{
				{
					ID:      "id1",
					Message: "my message 1",
				},
				{
					ID:      "id2",
					Message: "my message 2",
				},
			},
			TopicARN: topicArn,
		}

		param := &awsSNS.PublishBatchInput{
			TopicArn: aws.String(input.TopicARN),
			PublishBatchRequestEntries: []types.PublishBatchRequestEntry{
				{
					Id:      aws.String(input.Messages[0].ID),
					Message: aws.String(input.Messages[0].Message),
				},
				{
					Id:      aws.String(input.Messages[1].ID),
					Message: aws.String(input.Messages[1].Message),
				},
			},
		}

		suite.snsCLient.On("PublishBatch", ctx, param).
			Return(
				&awsSNS.PublishBatchOutput{
					Successful: []types.PublishBatchResultEntry{
						{
							Id:        aws.String(input.Messages[0].ID),
							MessageId: aws.String("id123"),
						},
					},
					Failed: []types.BatchResultErrorEntry{
						{
							Id:      aws.String(input.Messages[1].ID),
							Message: aws.String("error"),
						},
					},
				}, nil).
			Once()
		got, err := suite.producer.ProduceBatch(ctx, &input)
		suite.NoError(err)
		suite.Equal(1, len(got.Successful))
		suite.Equal(1, len(got.Failed))
		suite.Equal("id1", got.Successful[0].EntryID)
		suite.Equal("id123", got.Successful[0].MessageID)
		suite.Equal("id2", got.Failed[0].EntryID)
		suite.Equal("failed to publish message; error: error", got.Failed[0].Err.Error())
	})

	suite.Run("Should produce with message attributes", func() {
		topicArn, err := sns.BuildTopicARN("us-east-1", "0000000", "my_topic")
		suite.NoError(err)
		input := sns.PublishBatchInput{
			Messages: []*sns.PublishBatchEntry{
				{
					ID:      "id",
					Message: "my message",
					Attributes: map[string]string{
						"custom": "custom_value",
					},
				},
			},
			TopicARN: topicArn,
		}
		param := &awsSNS.PublishBatchInput{
			TopicArn: aws.String(input.TopicARN),
			PublishBatchRequestEntries: []types.PublishBatchRequestEntry{
				{
					Id:      aws.String(input.Messages[0].ID),
					Message: aws.String(input.Messages[0].Message),
					MessageAttributes: map[string]types.MessageAttributeValue{
						"custom": {DataType: aws.String("String"), StringValue: aws.String("custom_value")},
					},
				},
			},
		}

		rID := "id-123"
		suite.snsCLient.On("PublishBatch", ctx, param).
			Return(
				&awsSNS.PublishBatchOutput{
					Successful: []types.PublishBatchResultEntry{
						{
							Id:        aws.String(input.Messages[0].ID),
							MessageId: aws.String(rID),
						},
					},
				},
				nil,
			).
			Once()

		got, err := suite.producer.ProduceBatch(ctx, &input)
		suite.NoError(err)
		suite.Equal("id-123", got.Successful[0].MessageID)
	})

	suite.Run("Should produce with GroupID", func() {
		topicArn, err := sns.BuildTopicARN("us-east-1", "0000000", "my_topic")
		suite.NoError(err)

		input := sns.PublishBatchInput{
			Messages: []*sns.PublishBatchEntry{
				{
					ID:      "id",
					Message: "my message",
					GroupID: "my-group",
				},
			},
			TopicARN: topicArn,
		}

		param := &awsSNS.PublishBatchInput{
			TopicArn: aws.String(input.TopicARN),
			PublishBatchRequestEntries: []types.PublishBatchRequestEntry{
				{
					Id:             aws.String(input.Messages[0].ID),
					Message:        aws.String(input.Messages[0].Message),
					MessageGroupId: aws.String("my-group"),
				},
			},
		}

		rID := "id"
		suite.snsCLient.On("PublishBatch", ctx, param).
			Return(
				&awsSNS.PublishBatchOutput{
					Successful: []types.PublishBatchResultEntry{
						{
							Id:        aws.String(input.Messages[0].ID),
							MessageId: aws.String(rID),
						},
					},
				},
				nil,
			).
			Once()

		got, err := suite.producer.ProduceBatch(ctx, &input)
		suite.NoError(err)
		suite.Equal("id", got.Successful[0].MessageID)
	})

	suite.Run("Should produce with all fields", func() {
		topicArn, err := sns.BuildTopicARN("us-east-1", "0000000", "my_topic")
		suite.NoError(err)

		input := sns.PublishBatchInput{
			Messages: []*sns.PublishBatchEntry{
				{
					ID:              "id",
					Message:         "my message",
					GroupID:         "my-group",
					DeduplicationID: "dedup-id",
					Attributes: map[string]string{
						"custom": "custom_value",
					},
				},
			},
			TopicARN: topicArn,
		}

		param := &awsSNS.PublishBatchInput{
			TopicArn: aws.String(input.TopicARN),
			PublishBatchRequestEntries: []types.PublishBatchRequestEntry{
				{
					Id:      aws.String(input.Messages[0].ID),
					Message: aws.String(input.Messages[0].Message),
					MessageAttributes: map[string]types.MessageAttributeValue{
						"custom": {DataType: aws.String("String"), StringValue: aws.String("custom_value")},
					},
					MessageGroupId:         aws.String("my-group"),
					MessageDeduplicationId: aws.String("dedup-id"),
				},
			},
		}

		rID := "id"
		suite.snsCLient.On("PublishBatch", ctx, param).
			Return(
				&awsSNS.PublishBatchOutput{
					Successful: []types.PublishBatchResultEntry{
						{
							Id:        aws.String(input.Messages[0].ID),
							MessageId: aws.String(rID),
						},
					},
				},
				nil,
			).
			Once()

		got, err := suite.producer.ProduceBatch(ctx, &input)
		suite.NoError(err)
		suite.Equal("id", got.Successful[0].MessageID)
	})

	suite.Run("SNS PublishBatch error", func() {
		topicArn, err := sns.BuildTopicARN("us-east-1", "0000000", "my_topic")
		suite.NoError(err)

		input := sns.PublishBatchInput{
			Messages: []*sns.PublishBatchEntry{
				{
					ID:              "id",
					Message:         "my message",
					GroupID:         "my-group",
					DeduplicationID: "dedup-id",
					Attributes: map[string]string{
						"custom": "custom_value",
					},
				},
			},
			TopicARN: topicArn,
		}

		param := &awsSNS.PublishBatchInput{
			TopicArn: aws.String(input.TopicARN),
			PublishBatchRequestEntries: []types.PublishBatchRequestEntry{
				{
					Id:      aws.String(input.Messages[0].ID),
					Message: aws.String(input.Messages[0].Message),
					MessageAttributes: map[string]types.MessageAttributeValue{
						"custom": {DataType: aws.String("String"), StringValue: aws.String("custom_value")},
					},
					MessageGroupId:         aws.String("my-group"),
					MessageDeduplicationId: aws.String("dedup-id"),
				},
			},
		}

		suite.snsCLient.On("PublishBatch", ctx, param).
			Return(nil, fmt.Errorf("got error")).
			Once()

		got, err := suite.producer.ProduceBatch(ctx, &input)
		suite.Empty(got)
		suite.NotNil(err)
		suite.ErrorContains(err, "got error")
	})

	suite.Run("With Input nil", func() {
		got, err := suite.producer.ProduceBatch(ctx, nil)
		suite.Empty(got)
		suite.NotNil(err)
		suite.ErrorIs(err, loafergo.ErrEmptyInput)
	})

	suite.Run("With emtpy Messages", func() {
		topicArn, err := sns.BuildTopicARN("us-east-1", "0000000", "my_topic")
		suite.NoError(err)

		input := sns.PublishBatchInput{
			Messages: []*sns.PublishBatchEntry{},
			TopicARN: topicArn,
		}

		got, err := suite.producer.ProduceBatch(ctx, &input)
		suite.Empty(got)
		suite.NotNil(err)
		suite.ErrorIs(err, loafergo.ErrEmptyInput)
	})

	suite.Run("Too many batch messages", func() {
		topicArn, err := sns.BuildTopicARN("us-east-1", "0000000", "my_topic")
		suite.NoError(err)

		input := sns.PublishBatchInput{
			Messages: []*sns.PublishBatchEntry{},
			TopicARN: topicArn,
		}

		for i := 0; i < sns.DefaultMaxBatchSize+1; i++ {
			input.Messages = append(input.Messages, &sns.PublishBatchEntry{
				ID:      fmt.Sprintf("id-%d", i),
				Message: fmt.Sprintf("my message %d", i),
			})
		}

		got, err := suite.producer.ProduceBatch(ctx, &input)
		suite.Empty(got)
		suite.NotNil(err)
		suite.ErrorContains(err, fmt.Sprintf("maximum batch size is %d", sns.DefaultMaxBatchSize))
	})
}
