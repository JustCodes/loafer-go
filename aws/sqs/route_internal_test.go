package sqs

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/stretchr/testify/mock"

	loafergo "github.com/justcodes/loafer-go/v2"
	"github.com/justcodes/loafer-go/v2/fake"
)

func stubHandler(ctx context.Context, m loafergo.Message) error {
	fmt.Printf("Message received handler1: %+v\n ", m)
	return nil
}

func TestRouteChangeMessageVisibility(t *testing.T) {
	mockSQSClient := &fake.SQSClient{}
	m := &message{
		dispatched: make(chan bool, 1),
		originalMessage: types.Message{
			Body:          aws.String("body"),
			ReceiptHandle: aws.String("receipt-handler"),
		},
	}
	logger := new(fake.Logger)
	logger.On("Log", mock.Anything).Return()

	t.Run("Should stop changeMessageVisibility when dipatch messages", func(t *testing.T) {
		r := &route{
			sqs:               mockSQSClient,
			handler:           stubHandler,
			queueName:         "queue-name",
			queueURL:          "queue-url",
			extensionLimit:    2,
			visibilityTimeout: 11,
			maxMessages:       10,
			waitTimeSeconds:   10,
			workerPoolSize:    1,
		}

		mockSQSClient.On("ChangeMessageVisibility", context.Background(), &sqs.ChangeMessageVisibilityInput{
			QueueUrl:          aws.String("queue-url"),
			ReceiptHandle:     aws.String("receipt-handler"),
			VisibilityTimeout: 11,
		}).
			Return(&sqs.ChangeMessageVisibilityOutput{}, nil)

		mockSQSClient.On("ChangeMessageVisibility", context.Background(), &sqs.ChangeMessageVisibilityInput{
			QueueUrl:          aws.String("queue-url"),
			ReceiptHandle:     aws.String("receipt-handler"),
			VisibilityTimeout: 22,
		}).
			Return(&sqs.ChangeMessageVisibilityOutput{}, nil)

		ctx := context.Background()
		go r.changeMessageVisibility(ctx, m, logger)
		time.Sleep(1002 * time.Millisecond)
		m.Dispatch()
	})

	t.Run("Should stop changeMessageVisibility when count >= 2", func(t *testing.T) {
		r := &route{
			sqs:               mockSQSClient,
			handler:           stubHandler,
			queueName:         "queue-name",
			queueURL:          "queue-url",
			extensionLimit:    2,
			visibilityTimeout: 11,
			maxMessages:       10,
			waitTimeSeconds:   10,
			workerPoolSize:    1,
		}

		input := &sqs.ChangeMessageVisibilityInput{
			QueueUrl:          aws.String("queue-url"),
			ReceiptHandle:     aws.String("receipt-handler"),
			VisibilityTimeout: 22,
		}

		mockSQSClient.On("ChangeMessageVisibility", context.Background(), input).
			Return(&sqs.ChangeMessageVisibilityOutput{}, nil)

		input = &sqs.ChangeMessageVisibilityInput{
			QueueUrl:          aws.String("queue-url"),
			ReceiptHandle:     aws.String("receipt-handler"),
			VisibilityTimeout: 33,
		}

		mockSQSClient.On("ChangeMessageVisibility", context.Background(), input).
			Return(&sqs.ChangeMessageVisibilityOutput{}, nil)

		ctx := context.Background()
		r.changeMessageVisibility(ctx, m, logger)
	})

}
