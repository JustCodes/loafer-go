package loafergo

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/service/sqs"
)

const (
	defaultVisibilityTimeoutControl = 10
	defaultRetryTimeout             = 10 * time.Second
)

// Route holds the route fields
type Route struct {
	sqs               *sqs.SQS
	queueName         string
	queueURL          string
	handler           Handler
	visibilityTimeout int
	logger            Logger
	maxMessages       int64
	ExtensionLimit    int
	waitTimeSeconds   int64
}

// NewRoute creates a new Route
func NewRoute(queueName string, handler Handler, maxMessages int64, visibilityTimeout, waitTimeSeconds int) *Route {
	if visibilityTimeout <= 0 {
		visibilityTimeout = 30
	}
	if maxMessages <= 0 {
		maxMessages = 10
	}
	if waitTimeSeconds <= 0 {
		waitTimeSeconds = 10
	}
	return &Route{
		queueName:         queueName,
		handler:           handler,
		visibilityTimeout: visibilityTimeout,
		maxMessages:       maxMessages,
		ExtensionLimit:    2,
		waitTimeSeconds:   int64(waitTimeSeconds),
	}
}

func (r *Route) configure(c *Config) error {
	sess, err := newSession(c)
	if err != nil {
		return err
	}
	r.sqs = sqs.New(sess)

	if c.Logger == nil {
		r.logger = &defaultLogger{}
	}

	o, err := r.sqs.GetQueueUrl(&sqs.GetQueueUrlInput{QueueName: &r.queueName})
	if err != nil {
		r.logger.Println("error getting queue url for %s", r.queueName)
		return err
	}
	r.queueURL = *o.QueueUrl

	return nil
}

var (
	all = "All"
)

func (r *Route) run(workerPool int) {
	jobs := make(chan *message)
	for w := 1; w <= workerPool; w++ {
		go r.worker(w, jobs)
	}

	for {
		output, err := r.sqs.ReceiveMessage(
			&sqs.ReceiveMessageInput{
				QueueUrl:              &r.queueURL,
				WaitTimeSeconds:       &r.waitTimeSeconds,
				MaxNumberOfMessages:   &r.maxMessages,
				MessageAttributeNames: []*string{&all},
			},
		)
		if err != nil {
			r.Logger().Println("%s , retrying in 10s", ErrGetMessage.Context(err).Error())
			time.Sleep(defaultRetryTimeout)
			continue
		}

		for _, m := range output.Messages {
			// a message will be sent to the DLQ automatically after 4 tries if it is received but not deleted
			jobs <- newMessage(m)
		}
	}
}

// Logger returns a Logger
// if config logger is a nil so it will set a defaultLogger
func (r *Route) Logger() Logger {
	if r.logger == nil {
		return &defaultLogger{}
	}
	return r.logger
}

func (r *Route) worker(id int, messages <-chan *message) {
	for m := range messages {
		if err := r.dispatch(m); err != nil {
			r.Logger().Println(err.Error())
		}
	}
}

func (r *Route) dispatch(m *message) error {
	ctx := context.Background()

	go r.extend(ctx, m)
	if err := r.handler(ctx, m); err != nil {
		return m.ErrorResponse(ctx, err)
	}

	// finish the extension channel if the message was processed successfully
	_ = m.Success(ctx)

	// deletes message if the handler was successful or if there was no handler with that route
	return r.delete(m) // MESSAGE CONSUMED
}

func (r *Route) extend(ctx context.Context, m *message) {
	var count int
	extension := int64(r.visibilityTimeout)
	for {
		// only allow 1 extension (Default 1m30s)
		if count >= r.ExtensionLimit {
			r.Logger().Println(ErrMessageProcessing.Error(), r.queueName)
			return
		}

		count++
		// allow 10 seconds to process the extension request
		time.Sleep(time.Duration(r.visibilityTimeout-defaultVisibilityTimeoutControl) * time.Second)
		select {
		case <-m.err:
			// goroutine finished
			return
		default:
			// double the allowed processing time
			extension += int64(r.visibilityTimeout)
			_, err := r.sqs.ChangeMessageVisibility(
				&sqs.ChangeMessageVisibilityInput{
					QueueUrl:          &r.queueURL,
					ReceiptHandle:     m.ReceiptHandle,
					VisibilityTimeout: &extension,
				},
			)
			if err != nil {
				r.Logger().Println(ErrUnableToExtend.Error(), err.Error())
				return
			}
		}
	}
}

func (r *Route) delete(m *message) error {
	_, err := r.sqs.DeleteMessage(&sqs.DeleteMessageInput{QueueUrl: &r.queueURL, ReceiptHandle: m.ReceiptHandle})
	if err != nil {
		r.Logger().Println(ErrUnableToDelete.Context(err).Error())
		return ErrUnableToDelete.Context(err)
	}
	return nil
}
