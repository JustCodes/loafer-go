package sqs

import (
	"encoding/json"

	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

// message serves as a wrapper for sqs.Message as well as controls the error handling channel
type message struct {
	dispatched chan bool
	types.Message
}

func newMessage(m types.Message) *message {
	return &message{
		dispatched: make(chan bool, 1),
		Message:    m,
	}
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
func (m *message) Metadata() map[string]string {
	return m.Message.Attributes
}

// Decode will unmarshal the message into a supplied output using json
func (m *message) Decode(out interface{}) error {
	return json.Unmarshal(m.body(), &out)
}

// Attribute will return the custom attribute that was sent with the request.
// Each message attribute consists of a Name, Type, and Value. For more information,
// see Amazon SQS message attributes
// (https://docs.aws.amazon.com/AWSSimpleQueueService/latest/SQSDeveloperGuide/sqs-message-metadata.html#sqs-message-attributes)
// in the Amazon SQS Developer Guide.
func (m *message) Attribute(key string) string {
	id, ok := m.MessageAttributes[key]
	if !ok {
		return ""
	}

	return *id.StringValue
}

// Identifier An identifier associated with the act of receiving the message.
func (m *message) Identifier() string {
	return *m.ReceiptHandle
}

// Dispatch sets dispatched as true
func (m *message) Dispatch() {
	m.dispatched <- true
}

// Body returns the message body as []byte
func (m *message) Body() []byte {
	return m.body()
}
