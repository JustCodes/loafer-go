package loafergo

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

const (
	defaultRetryTimeout             = 10 * time.Second
	all                             = "All"
	defaultVisibilityTimeoutControl = 10
)

// Route holds the route fields
type Route struct {
	sqs               *sqs.Client
	queueName         string
	queueURL          string
	handler           Handler
	visibilityTimeout int32
	logger            Logger
	maxMessages       int32
	extensionLimit    int
	waitTimeSeconds   int32
}

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
// loafergo.NewRoute(
//
//		"queuename-1",
//		handler1,
//		loafergo.RouteWithVisibilityTimeout(25),
//		loafergo.RouteWithMaxMessages(5),
//		loafergo.RouteWithWaitTimeSeconds(8),
//	)
func NewRoute(queueName string, handler Handler, optConfigFns ...func(config *RouteConfig)) *Route {
	cfg := loadDefaultRouteConfig()
	for _, optFn := range optConfigFns {
		optFn(cfg)
	}

	return &Route{
		queueName:         queueName,
		handler:           handler,
		visibilityTimeout: cfg.visibilityTimeout,
		maxMessages:       cfg.maxMessages,
		extensionLimit:    cfg.extensionLimit,
		waitTimeSeconds:   cfg.waitTimeSeconds,
	}
}

func (r *Route) configure(ctx context.Context, s aws.Config, l Logger) error {
	r.sqs = sqs.NewFromConfig(s)
	r.logger = l

	o, err := r.sqs.GetQueueUrl(ctx, &sqs.GetQueueUrlInput{QueueName: &r.queueName})
	if err != nil {
		r.logger.Log(fmt.Sprintf("error getting queue url for %s", r.queueName))
		return err
	}
	r.queueURL = *o.QueueUrl

	return nil
}

func (r *Route) run(ctx context.Context, workerPool int) {
	jobs := make(chan *message)
	for w := 1; w <= workerPool; w++ {
		go r.worker(ctx, w, jobs)
	}

	for {
		output, err := r.sqs.ReceiveMessage(
			ctx,
			&sqs.ReceiveMessageInput{
				QueueUrl:              &r.queueURL,
				WaitTimeSeconds:       r.waitTimeSeconds,
				MaxNumberOfMessages:   r.maxMessages,
				MessageAttributeNames: []string{all},
			},
		)
		if err != nil {
			r.logger.Log(
				fmt.Sprintf(
					"%s , retrying in %fs",
					ErrGetMessage.Context(err).Error(),
					defaultRetryTimeout.Seconds(),
				),
			)
			time.Sleep(defaultRetryTimeout)
			continue
		}

		for _, m := range output.Messages {
			// a message will be sent to the DLQ automatically after 4 tries if it is received but not deleted
			jobs <- newMessage(m)
		}
	}
}

func (r *Route) worker(ctx context.Context, id int, messages <-chan *message) {
	for m := range messages {
		if err := r.dispatch(ctx, m); err != nil {
			r.logger.Log(err.Error())
		}
	}
}

func (r *Route) dispatch(ctx context.Context, m *message) error {
	go r.extend(ctx, m)
	if err := r.handler(ctx, m); err != nil {
		return m.ErrorResponse(ctx, err)
	}

	// finish the extension channel if the message was processed successfully
	_ = m.Success(ctx)

	return r.commitMessage(ctx, m)
}

func (r *Route) extend(ctx context.Context, m *message) {
	var count int
	extension := r.visibilityTimeout
	for {
		// only allow extensionLimit extension (Default 1m30s)
		if count >= r.extensionLimit {
			r.logger.Log(ErrMessageProcessing.Error(), r.queueName)
			return
		}

		count++
		time.Sleep(time.Duration(r.visibilityTimeout-defaultVisibilityTimeoutControl) * time.Second)
		select {
		case <-m.err:
			return
		default:
			// double the allowed processing time
			extension += r.visibilityTimeout
			_, err := r.sqs.ChangeMessageVisibility(
				ctx,
				&sqs.ChangeMessageVisibilityInput{
					QueueUrl:          &r.queueURL,
					ReceiptHandle:     m.ReceiptHandle,
					VisibilityTimeout: extension,
				},
			)
			if err != nil {
				r.logger.Log(ErrUnableToExtend.Error(), err.Error())
				return
			}
		}
	}
}

func (r *Route) commitMessage(ctx context.Context, m *message) error {
	_, err := r.sqs.DeleteMessage(
		ctx,
		&sqs.DeleteMessageInput{QueueUrl: &r.queueURL, ReceiptHandle: m.ReceiptHandle},
	)
	if err != nil {
		r.logger.Log(ErrUnableToDelete.Context(err).Error())
		return ErrUnableToDelete.Context(err)
	}
	return nil
}
