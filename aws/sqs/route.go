package sqs

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"

	loafergo "github.com/justcodes/loafer-go/v2"
)

const (
	all                             = "All"
	defaultVisibilityTimeoutControl = 10
	// deleteBufferHeadroom multiplies the delete buffer so a flush in flight does not
	// push Commit onto the single delete fallback.
	deleteBufferHeadroom = 4
)

type route struct {
	sqs                loafergo.SQSClient
	handler            loafergo.Handler
	logger             loafergo.Logger
	batcher            *deleteBatcher
	queueName          string
	queueURL           string
	customGroupFields  []string
	extensionLimit     int
	runMode            loafergo.Mode
	batcherOnce        sync.Once
	visibilityTimeout  int32
	maxMessages        int32
	waitTimeSeconds    int32
	workerPoolSize     int32
	deleteBatchEnabled bool
}

// DoneCtxKey is the context key for the done channel that is optionally passed to the router
// and used inside the doChangeVisibilityTimeout function to signal that the method finished its execution
type DoneCtxKey struct{}

// NewRoute creates a new Route
// By default, the new route will set the followed values:
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
		sqs:                config.SQSClient,
		handler:            config.Handler,
		logger:             cfg.logger,
		queueName:          config.QueueName,
		extensionLimit:     cfg.extensionLimit,
		visibilityTimeout:  cfg.visibilityTimeout,
		maxMessages:        cfg.maxMessages,
		waitTimeSeconds:    cfg.waitTimeSeconds,
		workerPoolSize:     cfg.workerPoolSize,
		runMode:            cfg.runMode,
		customGroupFields:  cfg.customGroupFields,
		deleteBatchEnabled: cfg.deleteBatchEnabled,
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
	r.startDeleteBatcher(ctx)
	return nil
}

// startDeleteBatcher spins up the delete batcher, which lives for as long as ctx does.
// It is a no-op unless RouteWithDeleteBatch was used, which keeps the default route on the
// original one-delete-per-message path.
func (r *route) startDeleteBatcher(ctx context.Context) {
	if !r.deleteBatchEnabled {
		return
	}

	r.batcherOnce.Do(func() {
		// The buffer has to absorb everything the route hands over while the flusher is
		// blocked on a delete round trip, which is more than one receive batch plus one
		// in-flight Commit per worker. Benchmarking 2000 messages showed a buffer that
		// size saturating and falling back to single deletes; the headroom below removes
		// it, and costs only a few hundred pointers.
		capacity := deleteBufferHeadroom * (int(r.maxMessages) + int(r.workerPoolSize) + maxDeleteBatchSize)
		b := newDeleteBatcher(r.sqs, r.logger, r.queueURL, capacity)
		r.batcher = b
		go b.run(ctx)
	})
}

// GetMessages gets messages from queue
func (r *route) GetMessages(ctx context.Context, logger loafergo.Logger) (messages []loafergo.Message, err error) {
	output, err := r.sqs.ReceiveMessage(
		ctx,
		&sqs.ReceiveMessageInput{
			QueueUrl:                    &r.queueURL,
			WaitTimeSeconds:             r.waitTimeSeconds,
			MaxNumberOfMessages:         r.maxMessages,
			VisibilityTimeout:           r.visibilityTimeout,
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
		go r.changeMessageVisibility(ctx, msg, logger)
	}

	return
}

// Commit deletes the message from the queue.
//
// When the route was built with RouteWithDeleteBatch the message is handed to the delete
// batcher and Commit returns nil immediately, without blocking the worker - the delete
// outcome is logged by the batcher instead of being returned here.
func (r *route) Commit(ctx context.Context, m loafergo.Message) error {
	// if the handler backed off the message, we should not delete it
	if m.BackedOff() {
		return nil
	}

	// The batcher refuses when it is shutting down or saturated: keep the delete synchronous
	// then, rather than dropping it or blocking the worker.
	if r.batcher != nil && r.batcher.enqueue(m) {
		return nil
	}
	return r.deleteMessage(ctx, m)
}

// deleteMessage removes a single message from the queue.
func (r *route) deleteMessage(ctx context.Context, m loafergo.Message) error {
	defer m.Dispatch()
	identifier := m.Identifier()
	_, err := r.sqs.DeleteMessage(
		ctx,
		&sqs.DeleteMessageInput{QueueUrl: &r.queueURL, ReceiptHandle: &identifier},
	)
	return err
}

// HandlerMessage consumes the message from the queue
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

// RunMode returns the router run mode
func (r *route) RunMode(ctx context.Context) loafergo.Mode {
	return r.runMode
}

// CustomGroupFields returns the router custom group fields
func (r *route) CustomGroupFields(ctx context.Context) []string {
	return r.customGroupFields
}

// changeMessageVisibility only extends the message visibility timeout when processing
// is still ongoing after it. Since ReceiveMessage already requests VisibilityTimeout
// equal to r.visibilityTimeout, messages committed or dispatched before the first tick
// never trigger a ChangeMessageVisibility call.
func (r *route) changeMessageVisibility(ctx context.Context, m *message, logger loafergo.Logger) {
	var count int
	extension := r.visibilityTimeout
	sleepTime := time.Duration(r.visibilityTimeout-defaultVisibilityTimeoutControl) * time.Second
	ticker := time.NewTicker(sleepTime)
	defer ticker.Stop()

	for {
		// only allow extensionLimit extension (Default 1m30s) beyond the first renewal
		if count > r.extensionLimit {
			return
		}

		select {
		case d := <-m.backoffChannel:
			r.doChangeVisibilityTimeout(ctx, m, int32(d.Seconds()), logger)
			return
		case <-m.dispatched:
			return
		case <-ticker.C:
			// the first tick just renews the original timeout, subsequent ticks double it
			if count > 0 {
				extension += r.visibilityTimeout
			}
			r.doChangeVisibilityTimeout(ctx, m, extension, logger)
			count++
		}
	}
}

func (r *route) doChangeVisibilityTimeout(ctx context.Context, m *message, timeout int32, logger loafergo.Logger) {
	if timeout < 0 {
		timeout = 0
	}

	// https://docs.aws.amazon.com/AWSSimpleQueueService/latest/APIReference/API_ChangeMessageVisibility.html
	maxLimit := int32((12 * time.Hour).Seconds())
	if timeout > maxLimit {
		timeout = maxLimit
	}

	logMsg := fmt.Sprintf(
		"change_visibility_timeout; queueUrl: %s; receipt_hendler: %s; timeout: %ds",
		r.queueURL, *m.originalMessage.ReceiptHandle, timeout,
	)
	logger.Log(logMsg)
	_, err := r.sqs.ChangeMessageVisibility(
		ctx,
		&sqs.ChangeMessageVisibilityInput{
			QueueUrl:          &r.queueURL,
			ReceiptHandle:     m.originalMessage.ReceiptHandle,
			VisibilityTimeout: timeout,
		},
	)
	if err != nil {
		logMsg = fmt.Sprintf(
			"change_visibility_timeout_error: %v; queueUrl: %s; receipt_hendler: %s; timeout: %ds",
			err, r.queueURL, *m.originalMessage.ReceiptHandle, timeout,
		)
		logger.Log(logMsg)
	}

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
