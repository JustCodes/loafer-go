package loafer_go

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-sdk-go/service/sqs"
)

type Message interface {
	// Route returns the event name that is used for routing within a worker, e.g. post_published
	// Decode will unmarshal the message into a supplied output using json
	Decode(out interface{}) error
	// Attribute will return the custom attribute that was sent through out the request.
	Attribute(key string) string
	// Metadata will return the metadata that was sent through out the request.
	Metadata() map[string]*string
}

// message serves as a wrapper for sqs.Message as well as controls the error handling channel
type message struct {
	*sqs.Message
	err chan error
}

func newMessage(m *sqs.Message) *message {
	return &message{m, make(chan error, 1)}
}

func (m *message) body() []byte {
	return []byte(*m.Message.Body)
}

// A map of the attributes requested in ReceiveMessage to their respective values.
// Supported attributes:
//
//   - ApproximateReceiveCount
//
//   - ApproximateFirstReceiveTimestamp
//
//   - MessageDeduplicationId
//
//   - MessageGroupId
//
//   - SenderId
//
//   - SentTimestamp
//
//   - SequenceNumber
//
// ApproximateFirstReceiveTimestamp and SentTimestamp are each returned as an
// integer representing the epoch time (http://en.wikipedia.org/wiki/Unix_time)
// in milliseconds.
func (m *message) Metadata() map[string]*string {
	return m.Message.Attributes
}

// Decode will unmarshal the message into a supplied output using json
func (m *message) Decode(out interface{}) error {
	return json.Unmarshal(m.body(), &out)
}

// ErrorResponse is used to determine for error handling within the handler. When an error occurs,
// this function should be returned.
func (m *message) ErrorResponse(ctx context.Context, err error) error {
	go func() {
		m.err <- err
	}()
	return err
}

// Success is used to determine that a handler was successful in processing the message and the message should
// now be consumed. This will delete the message from the queue
func (m *message) Success(ctx context.Context) error {
	go func() {
		m.err <- nil
	}()

	return nil
}

// Attribute will return the custom attribute that was sent with the request.
// Each message attribute consists of a Name, Type, and Value. For more information,
// see Amazon SQS message attributes (https://docs.aws.amazon.com/AWSSimpleQueueService/latest/SQSDeveloperGuide/sqs-message-metadata.html#sqs-message-attributes)
// in the Amazon SQS Developer Guide.
func (m *message) Attribute(key string) string {
	id, ok := m.MessageAttributes[key]
	if !ok {
		return ""
	}

	return *id.StringValue
}
