package sqs

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"

	loafergo "github.com/justcodes/loafer-go/v2"
)

const (
	all                             = "All"
	defaultVisibilityTimeoutControl = 10
)

type route struct {
	sqs               loafergo.SQSClient
	handler           loafergo.Handler
	queueName         string
	queueURL          string
	extensionLimit    int
	visibilityTimeout int32
	maxMessages       int32
	waitTimeSeconds   int32
	workerPoolSize    int32
}

// DoneCtxKey is the context key for the done channel that is optionally passed to the router
// and used inside the doChangeVisibilityTimeout function to signal that the method finished its execution
type DoneCtxKey struct{}

// NewRoute creates a new Route
// By default the new route will set the followed values:
//
// Visibility timeout: 30 seconds
// Max message: 10 unit
// Wait time: 10 seconds
//
// Use the Route method to modify these values.
// Example:
//
// sqs.NewRoute(
//
//		&sqs.Config{
//			SQSClient: sqsClient,
//			Handler:   handler1,
//			QueueName: "example-1",
//		},
//		sqs.RouteWithVisibilityTimeout(25),
//		sqs.RouteWithMaxMessages(5),
//		sqs.RouteWithWaitTimeSeconds(8),
//	)
func NewRoute(config *Config, optFns ...func(config *RouteConfig)) loafergo.Router {
	cfg := loadDefaultRouteConfig()
	for _, optFn := range optFns {
		optFn(cfg)
	}

	return &route{
		sqs:               config.SQSClient,
		handler:           config.Handler,
		queueName:         config.QueueName,
		extensionLimit:    cfg.extensionLimit,
		visibilityTimeout: cfg.visibilityTimeout,
		maxMessages:       cfg.maxMessages,
		waitTimeSeconds:   cfg.waitTimeSeconds,
		workerPoolSize:    cfg.workerPoolSize,
	}
}

// Configure sets the queue url to route
func (r *route) Configure(ctx context.Context) error {
	err := r.checkRequiredFields()
	if err != nil {
		return err
	}

	o, err := r.sqs.GetQueueUrl(ctx, &sqs.GetQueueUrlInput{QueueName: &r.queueName})
	if err != nil {
		return err
	}

	r.queueURL = *o.QueueUrl
	return nil
}

// GetMessages gets messages from queue
func (r *route) GetMessages(ctx context.Context) (messages []loafergo.Message, err error) {
	output, err := r.sqs.ReceiveMessage(
		ctx,
		&sqs.ReceiveMessageInput{
			QueueUrl:                    &r.queueURL,
			WaitTimeSeconds:             r.waitTimeSeconds,
			MaxNumberOfMessages:         r.maxMessages,
			MessageAttributeNames:       []string{all},
			MessageSystemAttributeNames: []types.MessageSystemAttributeName{types.MessageSystemAttributeNameAll},
		},
	)
	if err != nil {
		return
	}

	for _, m := range output.Messages {
		msg := newMessage(m)
		messages = append(messages, msg)
		// change the message visibility
		go r.changeMessageVisibility(ctx, msg)
	}

	return
}

// Commit deletes the message from queue
func (r *route) Commit(ctx context.Context, m loafergo.Message) error {
	// if the handler  backed off the message, we should not delete it
	if m.BackedOff() {
		return nil
	}

	defer m.Dispatch()
	identifier := m.Identifier()
	_, err := r.sqs.DeleteMessage(
		ctx,
		&sqs.DeleteMessageInput{QueueUrl: &r.queueURL, ReceiptHandle: &identifier},
	)
	if err != nil {
		return err
	}
	return err
}

// HandlerMessage consumes the message from queue
func (r *route) HandlerMessage(ctx context.Context, msg loafergo.Message) error {
	err := r.handler(ctx, msg)
	if err != nil {
		msg.Dispatch()
		return err
	}
	return nil
}

// WorkerPoolSize returns the router worker pool size
func (r *route) WorkerPoolSize(ctx context.Context) int32 {
	return r.workerPoolSize
}

// VisibilityTimeout returns the router visibility timeout
func (r *route) VisibilityTimeout(ctx context.Context) int32 {
	return r.visibilityTimeout
}

func (r *route) changeMessageVisibility(ctx context.Context, m *message) {
	var count int
	extension := r.visibilityTimeout
	sleepTime := time.Duration(r.visibilityTimeout-defaultVisibilityTimeoutControl) * time.Second
	ticker := time.NewTicker(sleepTime)
	defer ticker.Stop()

	r.doChangeVisibilityTimeout(ctx, m, extension)

	for {
		// only allow extensionLimit extension (Default 1m30s)
		if count >= r.extensionLimit {
			break
		}

		select {
		case d := <-m.backoffChannel:
			r.doChangeVisibilityTimeout(ctx, m, int32(d.Seconds()))
			return
		case <-m.dispatched:
			return
		case <-ticker.C:
			count++
			// double the allowed processing time
			extension += r.visibilityTimeout
			r.doChangeVisibilityTimeout(ctx, m, extension)
		}
	}
}

func (r *route) doChangeVisibilityTimeout(ctx context.Context, m *message, timeout int32) {
	if timeout <= 0 {
		timeout = 0
	}

	//https://docs.aws.amazon.com/AWSSimpleQueueService/latest/APIReference/API_ChangeMessageVisibility.html
	maxLimit := int32((12 * time.Hour).Seconds())
	if timeout > maxLimit {
		timeout = maxLimit
	}

	_, _ = r.sqs.ChangeMessageVisibility(
		ctx,
		&sqs.ChangeMessageVisibilityInput{
			QueueUrl:          &r.queueURL,
			ReceiptHandle:     m.originalMessage.ReceiptHandle,
			VisibilityTimeout: timeout,
		},
	)

	done, ok := ctx.Value(DoneCtxKey{}).(chan bool)
	if ok {
		done <- true
	}
}

func (r *route) checkRequiredFields() error {
	if r.sqs == nil {
		return loafergo.ErrNoSQSClient
	}

	if r.handler == nil {
		return loafergo.ErrNoHandler
	}
	return nil
}
