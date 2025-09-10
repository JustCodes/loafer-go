package sqs_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsSqs "github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	loafergo "github.com/justcodes/loafer-go/v2"
	"github.com/justcodes/loafer-go/v2/aws/sqs"
	"github.com/justcodes/loafer-go/v2/fake"
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
	suite.route = suite.setupRouter()
	suite.logger = loafergo.LoggerFunc(func(args ...interface{}) {
		fmt.Println(args...)
	})
}

func (suite *routeSuite) setupRouter(fns ...func(*sqs.RouteConfig)) loafergo.Router {
	var optFns []func(*sqs.RouteConfig)
	optFns = append(optFns, sqs.RouteWithMaxMessages(15))
	optFns = append(optFns, sqs.RouteWithWaitTimeSeconds(8))
	optFns = append(optFns, sqs.RouteWithVisibilityTimeout(12))
	optFns = append(optFns, sqs.RouteWithWorkerPoolSize(11))
	optFns = append(optFns, sqs.RouteWithRunMode(loafergo.PerGroupID))
	optFns = append(optFns, sqs.RouteWithCustomGroupFields([]string{"test_id"}))
	optFns = append(optFns, fns...)

	return sqs.NewRoute(&sqs.Config{
		SQSClient: suite.sqsClient,
		Handler:   stubHandler,
		QueueName: "example-1",
	},
		optFns...,
	)
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
	logger := new(fake.Logger)
	logger.On("Log", mock.Anything).Return()

	suite.Run("Should return the messages", func() {
		suite.route = suite.setupRouter()
		ctx, done := setupContext(1)
		cParam := &awsSqs.GetQueueUrlInput{QueueName: aws.String("example-1")}
		suite.sqsClient.On("GetQueueUrl", ctx, cParam).
			Return(&awsSqs.GetQueueUrlOutput{QueueUrl: aws.String("example-1-url")}, nil).
			Once()

		err := suite.route.Configure(ctx)
		suite.NoError(err)
		param := &awsSqs.ReceiveMessageInput{
			QueueUrl:                    aws.String("example-1-url"),
			WaitTimeSeconds:             8,
			MaxNumberOfMessages:         15,
			MessageAttributeNames:       []string{"All"},
			MessageSystemAttributeNames: []types.MessageSystemAttributeName{types.MessageSystemAttributeNameAll},
		}
		suite.sqsClient.On("ReceiveMessage", ctx, param).
			Return(&awsSqs.ReceiveMessageOutput{
				Messages: []types.Message{{
					Body:          aws.String("hello world"),
					ReceiptHandle: aws.String("receipt-handle"),
				}},
			}, nil).
			Once()

		suite.sqsClient.On("ChangeMessageVisibility", ctx, &awsSqs.ChangeMessageVisibilityInput{
			QueueUrl:          aws.String("example-1-url"),
			ReceiptHandle:     aws.String("receipt-handle"),
			VisibilityTimeout: int32(12),
		}).Return(nil, nil).Once()

		messages, err := suite.route.GetMessages(ctx, logger)
		suite.NoError(err)
		<-done

		suite.Len(messages, 1)
		suite.Equal("hello world", string(messages[0].Body()))

	})

	suite.Run("Should return error when receive message", func() {
		suite.route = suite.setupRouter()
		ctx, done := setupContext(1)
		cParam := &awsSqs.GetQueueUrlInput{QueueName: aws.String("example-1")}
		suite.sqsClient.On("GetQueueUrl", ctx, cParam).
			Return(&awsSqs.GetQueueUrlOutput{QueueUrl: aws.String("example-1-url")}, nil).
			Once()

		err := suite.route.Configure(ctx)
		suite.NoError(err)
		param := &awsSqs.ReceiveMessageInput{
			QueueUrl:                    aws.String("example-1-url"),
			WaitTimeSeconds:             8,
			MaxNumberOfMessages:         15,
			MessageAttributeNames:       []string{"All"},
			MessageSystemAttributeNames: []types.MessageSystemAttributeName{types.MessageSystemAttributeNameAll},
		}
		suite.sqsClient.On("ReceiveMessage", ctx, param).
			Return(nil, fmt.Errorf("got error")).
			Once()

		done <- true // wont call change message visibility
		messages, err := suite.route.GetMessages(ctx, logger)
		suite.NotNil(err)

		suite.Len(messages, 0)
		suite.Equal("got error", err.Error())

	})
}

// The goal of this is to create a channel that will be written on doChangeVisibilityTimeout of router
// to notify that the goroutine has finished execution, otherwise, the test cases may finish before and fail
func setupContext(amountOfVisibilityChangesExpected int) (context.Context, chan bool) {
	done := make(chan bool, amountOfVisibilityChangesExpected)
	k := sqs.DoneCtxKey{}
	return context.WithValue(context.Background(), k, done), done
}

func (suite *routeSuite) TestCommit() {
	logger := new(fake.Logger)
	logger.On("Log", mock.Anything).Return()

	suite.Run("Should commit commit", func() {
		suite.route = suite.setupRouter()
		ctx, done := setupContext(1)
		cParam := &awsSqs.GetQueueUrlInput{QueueName: aws.String("example-1")}
		suite.sqsClient.On("GetQueueUrl", ctx, cParam).
			Return(&awsSqs.GetQueueUrlOutput{QueueUrl: aws.String("example-1-url")}, nil).
			Once()

		err := suite.route.Configure(ctx)
		suite.NoError(err)
		param := &awsSqs.ReceiveMessageInput{
			QueueUrl:                    aws.String("example-1-url"),
			WaitTimeSeconds:             8,
			MaxNumberOfMessages:         15,
			MessageAttributeNames:       []string{"All"},
			MessageSystemAttributeNames: []types.MessageSystemAttributeName{types.MessageSystemAttributeNameAll},
		}
		suite.sqsClient.On("ReceiveMessage", ctx, param).
			Return(&awsSqs.ReceiveMessageOutput{
				Messages: []types.Message{{
					Body:          aws.String("hello world"),
					ReceiptHandle: aws.String("receipt-handle"),
				}},
			}, nil).
			Once()

		suite.sqsClient.On("ChangeMessageVisibility", ctx, &awsSqs.ChangeMessageVisibilityInput{
			QueueUrl:          aws.String("example-1-url"),
			ReceiptHandle:     aws.String("receipt-handle"),
			VisibilityTimeout: int32(12),
		}).Return(nil, nil).Once()

		message, err := suite.route.GetMessages(ctx, logger)
		suite.NoError(err)

		<-done

		commitParam := &awsSqs.DeleteMessageInput{
			QueueUrl:      aws.String("example-1-url"),
			ReceiptHandle: aws.String("receipt-handle"),
		}
		suite.sqsClient.On("DeleteMessage", ctx, commitParam).
			Return(nil, nil).
			Once()

		err = suite.route.Commit(ctx, message[0])
		suite.Nil(err)
	})

	suite.Run("Should return error when commit error", func() {
		suite.route = suite.setupRouter()
		ctx, done := setupContext(1)
		cParam := &awsSqs.GetQueueUrlInput{QueueName: aws.String("example-1")}
		suite.sqsClient.On("GetQueueUrl", ctx, cParam).
			Return(&awsSqs.GetQueueUrlOutput{QueueUrl: aws.String("example-1-url")}, nil).
			Once()

		err := suite.route.Configure(ctx)
		suite.NoError(err)
		param := &awsSqs.ReceiveMessageInput{
			QueueUrl:                    aws.String("example-1-url"),
			WaitTimeSeconds:             8,
			MaxNumberOfMessages:         15,
			MessageAttributeNames:       []string{"All"},
			MessageSystemAttributeNames: []types.MessageSystemAttributeName{types.MessageSystemAttributeNameAll},
		}
		suite.sqsClient.On("ReceiveMessage", ctx, param).
			Return(&awsSqs.ReceiveMessageOutput{
				Messages: []types.Message{{
					Body:          aws.String("hello world"),
					ReceiptHandle: aws.String("receipt-handle"),
				}},
			}, nil).
			Once()

		suite.sqsClient.On("ChangeMessageVisibility", ctx, &awsSqs.ChangeMessageVisibilityInput{
			QueueUrl:          aws.String("example-1-url"),
			ReceiptHandle:     aws.String("receipt-handle"),
			VisibilityTimeout: int32(12),
		}).Return(nil, nil).Once()

		message, err := suite.route.GetMessages(ctx, logger)
		suite.NoError(err)

		<-done

		commitParam := &awsSqs.DeleteMessageInput{
			QueueUrl:      aws.String("example-1-url"),
			ReceiptHandle: aws.String("receipt-handle"),
		}
		suite.sqsClient.On("DeleteMessage", ctx, commitParam).
			Return(nil, fmt.Errorf("got error")).
			Once()

		err = suite.route.Commit(ctx, message[0])
		suite.NotNil(err)
		suite.Equal("got error", err.Error())
	})
}

func (suite *routeSuite) TestHandlerMessage() {
	logger := new(fake.Logger)
	logger.On("Log", mock.Anything).Return()

	suite.Run("should handler message", func() {
		suite.route = suite.setupRouter()
		ctx, done := setupContext(1)
		cParam := &awsSqs.GetQueueUrlInput{QueueName: aws.String("example-1")}
		suite.sqsClient.On("GetQueueUrl", ctx, cParam).
			Return(&awsSqs.GetQueueUrlOutput{QueueUrl: aws.String("example-1-url")}, nil).
			Once()

		err := suite.route.Configure(ctx)
		suite.NoError(err)
		param := &awsSqs.ReceiveMessageInput{
			QueueUrl:                    aws.String("example-1-url"),
			WaitTimeSeconds:             8,
			MaxNumberOfMessages:         15,
			MessageAttributeNames:       []string{"All"},
			MessageSystemAttributeNames: []types.MessageSystemAttributeName{types.MessageSystemAttributeNameAll},
		}
		suite.sqsClient.On("ReceiveMessage", ctx, param).
			Return(&awsSqs.ReceiveMessageOutput{
				Messages: []types.Message{{
					Body:          aws.String("hello world"),
					ReceiptHandle: aws.String("receipt-handle"),
				}},
			}, nil).
			Once()

		suite.sqsClient.On("ChangeMessageVisibility", ctx, &awsSqs.ChangeMessageVisibilityInput{
			QueueUrl:          aws.String("example-1-url"),
			ReceiptHandle:     aws.String("receipt-handle"),
			VisibilityTimeout: int32(12),
		}).Return(nil, nil).Once()

		message, err := suite.route.GetMessages(ctx, logger)
		suite.NoError(err)
		suite.NoError(err)

		<-done

		err = suite.route.HandlerMessage(ctx, message[0])
		suite.Nil(err)
	})

	suite.Run("should return error when handler message error", func() {
		suite.route = suite.setupRouter()
		ctx, done := setupContext(1)
		cParam := &awsSqs.GetQueueUrlInput{QueueName: aws.String("example-1")}
		suite.sqsClient.On("GetQueueUrl", ctx, cParam).
			Return(&awsSqs.GetQueueUrlOutput{QueueUrl: aws.String("example-1-url")}, nil).
			Once()

		err := suite.route.Configure(ctx)
		suite.NoError(err)
		param := &awsSqs.ReceiveMessageInput{
			QueueUrl:                    aws.String("example-1-url"),
			WaitTimeSeconds:             8,
			MaxNumberOfMessages:         15,
			MessageAttributeNames:       []string{"All"},
			MessageSystemAttributeNames: []types.MessageSystemAttributeName{types.MessageSystemAttributeNameAll},
		}
		suite.sqsClient.On("ReceiveMessage", ctx, param).
			Return(&awsSqs.ReceiveMessageOutput{
				Messages: []types.Message{{
					Body:          aws.String("hello world"),
					ReceiptHandle: aws.String("receipt-handle"),
				}},
			}, nil).
			Once()

		suite.sqsClient.On("ChangeMessageVisibility", ctx, &awsSqs.ChangeMessageVisibilityInput{
			QueueUrl:          aws.String("example-1-url"),
			ReceiptHandle:     aws.String("receipt-handle"),
			VisibilityTimeout: int32(12),
		}).Return(nil, nil).Once()

		message, err := suite.route.GetMessages(ctx, logger)
		suite.NoError(err)

		<-done

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

func (suite *routeSuite) TestWorkPoolSize() {
	suite.Run("should work pool size", func() {
		suite.SetupSuite()
		ctx := context.Background()
		got := suite.route.WorkerPoolSize(ctx)
		suite.Equal(int32(11), got)
		suite.TearDownSuite()
	})

	suite.Run("should work pool size default value", func() {
		ctx := context.Background()
		suite.route = sqs.NewRoute(&sqs.Config{
			SQSClient: suite.sqsClient,
			Handler:   stubHandler,
			QueueName: "example-1",
		})
		got := suite.route.WorkerPoolSize(ctx)
		suite.Equal(got, int32(5))
		suite.TearDownSuite()
	})
}

func (suite *routeSuite) TestRunMode() {
	suite.Run("should run mode", func() {
		suite.SetupSuite()
		ctx := context.Background()
		got := suite.route.RunMode(ctx)
		suite.Equal(loafergo.PerGroupID, got)
		suite.TearDownSuite()
	})

	suite.Run("should run mode default value", func() {
		ctx := context.Background()
		suite.route = sqs.NewRoute(&sqs.Config{
			SQSClient: suite.sqsClient,
			Handler:   stubHandler,
			QueueName: "example-1",
		})
		got := suite.route.RunMode(ctx)
		suite.Equal(got, loafergo.Parallel)
		suite.TearDownSuite()
	})
}

func (suite *routeSuite) TestCustomGroupFields() {
	suite.Run("should custom group fields", func() {
		suite.SetupSuite()
		ctx := context.Background()
		got := suite.route.CustomGroupFields(ctx)
		suite.Equal([]string{"test_id"}, got)
		suite.TearDownSuite()
	})

	suite.Run("should custom group fields default value", func() {
		ctx := context.Background()
		suite.route = sqs.NewRoute(&sqs.Config{
			SQSClient: suite.sqsClient,
			Handler:   stubHandler,
			QueueName: "example-1",
		})
		got := suite.route.CustomGroupFields(ctx)
		suite.Equal([]string(nil), got)
		suite.Len(got, 0)
		suite.TearDownSuite()
	})
}

func (suite *routeSuite) TestChangeVisibilityInitially() {
	logger := new(fake.Logger)
	logger.On("Log", mock.Anything).Return()

	suite.Run("should change visibility timeout initially", func() {
		visibilityTimeout := 30
		suite.route = sqs.NewRoute(&sqs.Config{
			SQSClient: suite.sqsClient,
			Handler: func(ctx context.Context, m loafergo.Message) error {
				return nil
			},
			QueueName: "example-1",
		},
			sqs.RouteWithVisibilityTimeout(int32(visibilityTimeout)),
			sqs.RouteWithMaxMessages(10),
			sqs.RouteWithWaitTimeSeconds(10),
		)

		ctx, done := setupContext(1)

		visibilityTimeout = int(suite.route.VisibilityTimeout(ctx))

		receiptHandle := aws.String("receipt-handle")
		queueUrl := aws.String("example-1-url")

		cParam := &awsSqs.GetQueueUrlInput{QueueName: aws.String("example-1")}
		suite.sqsClient.On("GetQueueUrl", ctx, cParam).
			Return(&awsSqs.GetQueueUrlOutput{QueueUrl: queueUrl}, nil).
			Once()

		err := suite.route.Configure(ctx)
		suite.NoError(err)
		param := &awsSqs.ReceiveMessageInput{
			QueueUrl:                    queueUrl,
			WaitTimeSeconds:             10,
			MaxNumberOfMessages:         10,
			MessageAttributeNames:       []string{"All"},
			MessageSystemAttributeNames: []types.MessageSystemAttributeName{types.MessageSystemAttributeNameAll},
		}

		suite.sqsClient.On("ReceiveMessage", ctx, param).
			Return(&awsSqs.ReceiveMessageOutput{
				Messages: []types.Message{{
					Body:          aws.String("hello world"),
					ReceiptHandle: receiptHandle,
				}},
			}, nil).
			Once()

		suite.sqsClient.On("ChangeMessageVisibility", ctx, &awsSqs.ChangeMessageVisibilityInput{
			QueueUrl:          queueUrl,
			ReceiptHandle:     receiptHandle,
			VisibilityTimeout: int32(visibilityTimeout),
		}).Return(nil, nil).Once()

		messages, err := suite.route.GetMessages(ctx, logger)
		suite.NoError(err)

		<-done

		msg := messages[0]

		err = suite.route.HandlerMessage(ctx, msg)
		suite.Nil(err)

		suite.sqsClient.On("DeleteMessage", ctx, &awsSqs.DeleteMessageInput{
			QueueUrl:      queueUrl,
			ReceiptHandle: receiptHandle,
		}).Return(nil, nil).Once()

		err = suite.route.Commit(ctx, msg)
		suite.Nil(err)
	})

}

func (suite *routeSuite) TestChangeVisibilityTimeout() {
	logger := new(fake.Logger)
	logger.On("Log", mock.Anything).Return()

	suite.Run("should change visibility timeout when handler takes time to handle msg", func() {
		visibilityTimeout := 11 // sleepTime of ticker will be 1s
		suite.route = sqs.NewRoute(&sqs.Config{
			SQSClient: suite.sqsClient,
			Handler: func(ctx context.Context, m loafergo.Message) error {
				time.Sleep(1200 * time.Millisecond) // wait 1.5 cycles so that the visibility timeout will be changed on the ticker case
				return nil
			},
			QueueName: "example-1",
		},
			sqs.RouteWithVisibilityTimeout(int32(visibilityTimeout)),
			sqs.RouteWithMaxMessages(10),
			sqs.RouteWithWaitTimeSeconds(10),
		)

		ctx, done := setupContext(2)

		visibilityTimeout = int(suite.route.VisibilityTimeout(ctx))

		receiptHandle := aws.String("receipt-handle")
		queueUrl := aws.String("example-1-url")

		cParam := &awsSqs.GetQueueUrlInput{QueueName: aws.String("example-1")}
		suite.sqsClient.On("GetQueueUrl", ctx, cParam).
			Return(&awsSqs.GetQueueUrlOutput{QueueUrl: queueUrl}, nil).
			Once()

		err := suite.route.Configure(ctx)
		suite.NoError(err)
		param := &awsSqs.ReceiveMessageInput{
			QueueUrl:                    queueUrl,
			WaitTimeSeconds:             10,
			MaxNumberOfMessages:         10,
			MessageAttributeNames:       []string{"All"},
			MessageSystemAttributeNames: []types.MessageSystemAttributeName{types.MessageSystemAttributeNameAll},
		}

		suite.sqsClient.On("ReceiveMessage", ctx, param).
			Return(&awsSqs.ReceiveMessageOutput{
				Messages: []types.Message{{
					Body:          aws.String("hello world"),
					ReceiptHandle: receiptHandle,
				}},
			}, nil).
			Once()

		suite.sqsClient.On("ChangeMessageVisibility", ctx, &awsSqs.ChangeMessageVisibilityInput{
			QueueUrl:          queueUrl,
			ReceiptHandle:     receiptHandle,
			VisibilityTimeout: int32(visibilityTimeout),
		}).Return(nil, nil).Once()

		suite.sqsClient.On("ChangeMessageVisibility", ctx, &awsSqs.ChangeMessageVisibilityInput{
			QueueUrl:          queueUrl,
			ReceiptHandle:     receiptHandle,
			VisibilityTimeout: int32(visibilityTimeout) * 2,
		}).Return(nil, nil).Once()

		messages, err := suite.route.GetMessages(ctx, logger)
		suite.NoError(err)

		<-done

		msg := messages[0]

		err = suite.route.HandlerMessage(ctx, msg)
		suite.Nil(err)

		suite.sqsClient.On("DeleteMessage", ctx, &awsSqs.DeleteMessageInput{
			QueueUrl:      queueUrl,
			ReceiptHandle: receiptHandle,
		}).Return(nil, nil).Once()

		err = suite.route.Commit(ctx, msg)
		suite.Nil(err)
	})

}

func (suite *routeSuite) TestBackoff() {
	logger := new(fake.Logger)
	logger.On("Log", mock.Anything).Return()

	suite.Run("should change visibility timeout initially and when backoff is called", func() {
		visibilityTimeout := 30
		backoffTimeout := 10

		suite.route = sqs.NewRoute(&sqs.Config{
			SQSClient: suite.sqsClient,
			Handler: func(ctx context.Context, m loafergo.Message) error {
				m.Backoff(time.Duration(backoffTimeout) * time.Second)
				return nil
			},
			QueueName: "example-1",
		},
			sqs.RouteWithVisibilityTimeout(int32(visibilityTimeout)),
			sqs.RouteWithMaxMessages(10),
			sqs.RouteWithWaitTimeSeconds(10),
		)

		ctx, done := setupContext(2)

		visibilityTimeout = int(suite.route.VisibilityTimeout(ctx))

		receiptHandle := aws.String("receipt-handle")
		queueUrl := aws.String("example-1-url")

		cParam := &awsSqs.GetQueueUrlInput{QueueName: aws.String("example-1")}
		suite.sqsClient.On("GetQueueUrl", ctx, cParam).
			Return(&awsSqs.GetQueueUrlOutput{QueueUrl: queueUrl}, nil).
			Once()

		err := suite.route.Configure(ctx)
		suite.NoError(err)
		param := &awsSqs.ReceiveMessageInput{
			QueueUrl:                    queueUrl,
			WaitTimeSeconds:             10,
			MaxNumberOfMessages:         10,
			MessageAttributeNames:       []string{"All"},
			MessageSystemAttributeNames: []types.MessageSystemAttributeName{types.MessageSystemAttributeNameAll},
		}

		suite.sqsClient.On("ReceiveMessage", ctx, param).
			Return(&awsSqs.ReceiveMessageOutput{
				Messages: []types.Message{{
					Body:          aws.String("hello world"),
					ReceiptHandle: receiptHandle,
				}},
			}, nil).
			Once()

		suite.sqsClient.On("ChangeMessageVisibility", ctx, &awsSqs.ChangeMessageVisibilityInput{
			QueueUrl:          queueUrl,
			ReceiptHandle:     receiptHandle,
			VisibilityTimeout: int32(visibilityTimeout),
		}).Return(nil, nil).Once()

		suite.sqsClient.On("ChangeMessageVisibility", ctx, &awsSqs.ChangeMessageVisibilityInput{
			QueueUrl:          queueUrl,
			ReceiptHandle:     receiptHandle,
			VisibilityTimeout: int32(backoffTimeout),
		}).Return(nil, nil).Once()

		messages, err := suite.route.GetMessages(ctx, logger)
		suite.NoError(err)

		msg := messages[0]

		err = suite.route.HandlerMessage(ctx, msg)
		suite.Nil(err)

		<-done
		// should not delete the message if it was backedoff inside the handler
		suite.sqsClient.AssertNotCalled(suite.T(), "DeleteMessage")

		err = suite.route.Commit(ctx, msg)
		suite.Nil(err)
	})
}
