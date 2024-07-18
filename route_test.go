package loafergo_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/stretchr/testify/suite"

	loafergo "github.com/justcodes/loafer-go"
	"github.com/justcodes/loafer-go/fake"
)

func example1(ctx context.Context, m loafergo.Message) error {
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
	suite.route = loafergo.NewRoute(
		"example-1",
		example1,
		loafergo.RouteWithMaxMessages(15),
		loafergo.RouteWithWaitTimeSeconds(8),
		loafergo.RouteWithVisibilityTimeout(12),
	)
	suite.logger = loafergo.LoggerFunc(func(args ...interface{}) {
		fmt.Println(args...)
	})
}

func (suite *routeSuite) TestConfigure() {
	suite.Run("Should configure route", func() {
		param := &sqs.GetQueueUrlInput{QueueName: aws.String("example-1")}
		suite.sqsClient.On("GetQueueUrl", context.Background(), param).
			Return(&sqs.GetQueueUrlOutput{QueueUrl: aws.String("example-1")}, nil).
			Once()

		err := suite.route.Configure(context.Background(), suite.sqsClient, suite.logger)
		suite.NoError(err)
	})

	suite.Run("Should not configure route", func() {
		param := &sqs.GetQueueUrlInput{QueueName: aws.String("example-1")}
		suite.sqsClient.On("GetQueueUrl", context.Background(), param).
			Return(nil, fmt.Errorf("got error")).
			Once()

		err := suite.route.Configure(context.Background(), suite.sqsClient, suite.logger)
		suite.NotNil(err)
		suite.Equal("got error", err.Error())
	})
}
