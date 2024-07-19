package sqs_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsSqs "github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/stretchr/testify/suite"

	loafergo "github.com/justcodes/loafer-go"
	"github.com/justcodes/loafer-go/fake"
	"github.com/justcodes/loafer-go/sqs"
)

func stubHandler(ctx context.Context, m loafergo.Message) error {
	fmt.Printf("Message received handler1: %+v\n ", m)
	return nil
}

type routeSuite struct {
	suite.Suite
	sqsClient *fake.SQSClient
	logger    loafergo.Logger
	route     loafergo.Router
}

func TestRouteSuite(t *testing.T) {
	suite.Run(t, new(routeSuite))
}

func (suite *routeSuite) SetupSuite() {
	suite.sqsClient = fake.NewSQSClient(suite.T())
	suite.route = sqs.NewRoute(&sqs.Config{
		SQSClient: suite.sqsClient,
		Handler:   stubHandler,
		QueueName: "example-1",
	},
		sqs.RouteWithMaxMessages(15),
		sqs.RouteWithWaitTimeSeconds(8),
		sqs.RouteWithVisibilityTimeout(12),
	)
	suite.logger = loafergo.LoggerFunc(func(args ...interface{}) {
		fmt.Println(args...)
	})
}

func (suite *routeSuite) TearDownSuite() {
	suite.SetupSuite()
}

func (suite *routeSuite) TestConfigure() {
	suite.Run("Should configure route", func() {
		param := &awsSqs.GetQueueUrlInput{QueueName: aws.String("example-1")}
		suite.sqsClient.On("GetQueueUrl", context.Background(), param).
			Return(&awsSqs.GetQueueUrlOutput{QueueUrl: aws.String("example-1")}, nil).
			Once()

		err := suite.route.Configure(context.Background())
		suite.NoError(err)
	})

	suite.Run("Should not configure route", func() {
		param := &awsSqs.GetQueueUrlInput{QueueName: aws.String("example-1")}
		suite.sqsClient.On("GetQueueUrl", context.Background(), param).
			Return(nil, fmt.Errorf("got error")).
			Once()

		err := suite.route.Configure(context.Background())
		suite.NotNil(err)
		suite.Equal("got error", err.Error())
	})

	suite.Run("Should return error when sqs client is nil", func() {
		suite.route = sqs.NewRoute(&sqs.Config{
			SQSClient: nil,
			Handler:   stubHandler,
			QueueName: "example-1",
		})

		err := suite.route.Configure(context.Background())
		suite.NotNil(err)
		suite.ErrorIs(err, loafergo.ErrNoSQSClient)
		suite.TearDownSuite()
	})

	suite.Run("Should return error when handler is nil", func() {
		suite.route = sqs.NewRoute(&sqs.Config{
			SQSClient: suite.sqsClient,
			Handler:   nil,
			QueueName: "example-1",
		})

		err := suite.route.Configure(context.Background())
		suite.NotNil(err)
		suite.ErrorIs(err, loafergo.ErrNoHandler)
		suite.TearDownSuite()
	})
}

func (suite *routeSuite) TestGetMessages() {
	suite.Run("Should return the messages", func() {
		cParam := &awsSqs.GetQueueUrlInput{QueueName: aws.String("example-1")}
		suite.sqsClient.On("GetQueueUrl", context.Background(), cParam).
			Return(&awsSqs.GetQueueUrlOutput{QueueUrl: aws.String("example-1-url")}, nil).
			Once()

		err := suite.route.Configure(context.Background())
		suite.NoError(err)
		param := &awsSqs.ReceiveMessageInput{
			QueueUrl:              aws.String("example-1-url"),
			WaitTimeSeconds:       8,
			MaxNumberOfMessages:   15,
			MessageAttributeNames: []string{"All"},
		}
		suite.sqsClient.On("ReceiveMessage", context.Background(), param).
			Return(&awsSqs.ReceiveMessageOutput{
				Messages: []types.Message{{
					Body:          aws.String("hello world"),
					ReceiptHandle: aws.String("receipt-handle"),
				}},
			}, nil).
			Once()

		messages, err := suite.route.GetMessages(context.Background())
		suite.NoError(err)
		suite.Len(messages, 1)
		suite.Equal("hello world", string(messages[0].Body()))
	})

	suite.Run("Should return error when receive message", func() {
		cParam := &awsSqs.GetQueueUrlInput{QueueName: aws.String("example-1")}
		suite.sqsClient.On("GetQueueUrl", context.Background(), cParam).
			Return(&awsSqs.GetQueueUrlOutput{QueueUrl: aws.String("example-1-url")}, nil).
			Once()

		err := suite.route.Configure(context.Background())
		suite.NoError(err)
		param := &awsSqs.ReceiveMessageInput{
			QueueUrl:              aws.String("example-1-url"),
			WaitTimeSeconds:       8,
			MaxNumberOfMessages:   15,
			MessageAttributeNames: []string{"All"},
		}
		suite.sqsClient.On("ReceiveMessage", context.Background(), param).
			Return(nil, fmt.Errorf("got error")).
			Once()

		messages, err := suite.route.GetMessages(context.Background())
		suite.NotNil(err)
		suite.Len(messages, 0)
		suite.Equal("got error", err.Error())
	})
}

func (suite *routeSuite) TestCommit() {
	suite.Run("Should commit commit", func() {
		ctx := context.Background()
		cParam := &awsSqs.GetQueueUrlInput{QueueName: aws.String("example-1")}
		suite.sqsClient.On("GetQueueUrl", context.Background(), cParam).
			Return(&awsSqs.GetQueueUrlOutput{QueueUrl: aws.String("example-1-url")}, nil).
			Once()

		err := suite.route.Configure(context.Background())
		suite.NoError(err)
		param := &awsSqs.ReceiveMessageInput{
			QueueUrl:              aws.String("example-1-url"),
			WaitTimeSeconds:       8,
			MaxNumberOfMessages:   15,
			MessageAttributeNames: []string{"All"},
		}
		suite.sqsClient.On("ReceiveMessage", ctx, param).
			Return(&awsSqs.ReceiveMessageOutput{
				Messages: []types.Message{{
					Body:          aws.String("hello world"),
					ReceiptHandle: aws.String("receipt-handle"),
				}},
			}, nil).
			Once()

		message, err := suite.route.GetMessages(ctx)
		suite.NoError(err)

		commitParam := &awsSqs.DeleteMessageInput{
			QueueUrl:      aws.String("example-1-url"),
			ReceiptHandle: aws.String("receipt-handle"),
		}
		suite.sqsClient.On("DeleteMessage", context.Background(), commitParam).
			Return(nil, nil).
			Once()

		err = suite.route.Commit(ctx, message[0])
		suite.Nil(err)
	})

	suite.Run("Should return error when commit error", func() {
		ctx := context.Background()
		cParam := &awsSqs.GetQueueUrlInput{QueueName: aws.String("example-1")}
		suite.sqsClient.On("GetQueueUrl", context.Background(), cParam).
			Return(&awsSqs.GetQueueUrlOutput{QueueUrl: aws.String("example-1-url")}, nil).
			Once()

		err := suite.route.Configure(context.Background())
		suite.NoError(err)
		param := &awsSqs.ReceiveMessageInput{
			QueueUrl:              aws.String("example-1-url"),
			WaitTimeSeconds:       8,
			MaxNumberOfMessages:   15,
			MessageAttributeNames: []string{"All"},
		}
		suite.sqsClient.On("ReceiveMessage", ctx, param).
			Return(&awsSqs.ReceiveMessageOutput{
				Messages: []types.Message{{
					Body:          aws.String("hello world"),
					ReceiptHandle: aws.String("receipt-handle"),
				}},
			}, nil).
			Once()

		message, err := suite.route.GetMessages(ctx)
		suite.NoError(err)

		commitParam := &awsSqs.DeleteMessageInput{
			QueueUrl:      aws.String("example-1-url"),
			ReceiptHandle: aws.String("receipt-handle"),
		}
		suite.sqsClient.On("DeleteMessage", context.Background(), commitParam).
			Return(nil, fmt.Errorf("got error")).
			Once()

		err = suite.route.Commit(ctx, message[0])
		suite.NotNil(err)
		suite.Equal("got error", err.Error())
	})
}

func (suite *routeSuite) TestHandlerMessage() {
	suite.Run("should handler message", func() {
		ctx := context.Background()
		cParam := &awsSqs.GetQueueUrlInput{QueueName: aws.String("example-1")}
		suite.sqsClient.On("GetQueueUrl", context.Background(), cParam).
			Return(&awsSqs.GetQueueUrlOutput{QueueUrl: aws.String("example-1-url")}, nil).
			Once()

		err := suite.route.Configure(context.Background())
		suite.NoError(err)
		param := &awsSqs.ReceiveMessageInput{
			QueueUrl:              aws.String("example-1-url"),
			WaitTimeSeconds:       8,
			MaxNumberOfMessages:   15,
			MessageAttributeNames: []string{"All"},
		}
		suite.sqsClient.On("ReceiveMessage", ctx, param).
			Return(&awsSqs.ReceiveMessageOutput{
				Messages: []types.Message{{
					Body:          aws.String("hello world"),
					ReceiptHandle: aws.String("receipt-handle"),
				}},
			}, nil).
			Once()

		message, err := suite.route.GetMessages(ctx)
		suite.NoError(err)

		err = suite.route.HandlerMessage(ctx, message[0])
		suite.Nil(err)
	})

	suite.Run("should return error when handler message error", func() {
		ctx := context.Background()
		cParam := &awsSqs.GetQueueUrlInput{QueueName: aws.String("example-1")}
		suite.sqsClient.On("GetQueueUrl", context.Background(), cParam).
			Return(&awsSqs.GetQueueUrlOutput{QueueUrl: aws.String("example-1-url")}, nil).
			Once()

		err := suite.route.Configure(context.Background())
		suite.NoError(err)
		param := &awsSqs.ReceiveMessageInput{
			QueueUrl:              aws.String("example-1-url"),
			WaitTimeSeconds:       8,
			MaxNumberOfMessages:   15,
			MessageAttributeNames: []string{"All"},
		}
		suite.sqsClient.On("ReceiveMessage", ctx, param).
			Return(&awsSqs.ReceiveMessageOutput{
				Messages: []types.Message{{
					Body:          aws.String("hello world"),
					ReceiptHandle: aws.String("receipt-handle"),
				}},
			}, nil).
			Once()

		message, err := suite.route.GetMessages(ctx)
		suite.NoError(err)

		suite.route = sqs.NewRoute(&sqs.Config{
			SQSClient: suite.sqsClient,
			Handler: func(ctx context.Context, message loafergo.Message) error {
				return fmt.Errorf("got error")
			},
			QueueName: "example-1",
		})

		err = suite.route.HandlerMessage(ctx, message[0])
		suite.NotNil(err)
		suite.Equal("got error", err.Error())
	})
}
