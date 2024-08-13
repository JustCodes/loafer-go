package sns_test

import (
	"context"
	"testing"

	awsSNS "github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/stretchr/testify/suite"

	"github.com/justcodes/loafer-go/aws/sns"
	"github.com/justcodes/loafer-go/fake"
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
}
