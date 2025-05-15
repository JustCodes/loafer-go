package sqs

import (
	"encoding/json"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

type sqsMessage struct {
	Timestamp         time.Time                  `json:"Timestamp"`
	MessageAttributes map[string]CustomAttribute `json:"MessageAttributes"`
	Message           string                     `json:"Message"`
}

// message serves as a wrapper for sqs.Message as well as controls the error handling channel
type message struct {
	message         sqsMessage
	backoffChannel  chan time.Duration
	dispatched      chan bool
	originalMessage types.Message
	backedOff       bool
}

func newMessage(m types.Message) *message {
	var msg sqsMessage
	if m.Body != nil {
		_ = json.Unmarshal([]byte(*m.Body), &msg)
	}

	return &message{
		backedOff:       false,
		backoffChannel:  make(chan time.Duration, 1),
		dispatched:      make(chan bool, 1),
		originalMessage: m,
		message:         msg,
	}
}

func (m *message) body() []byte {
	if m.originalMessage.Body != nil {
		return []byte(*m.originalMessage.Body)
	}
	return []byte(``)
}

// Metadata A map of the attributes requested in ReceiveMessage to their respective values.
func (m *message) Metadata() map[string]string {
	attr := map[string]string{}
	for k, v := range m.message.MessageAttributes {
		attr[k] = v.Value
	}
	return attr
}

// Decode will unmarshal the message body into a supplied output using JSON
func (m *message) Decode(out interface{}) error {
	return json.Unmarshal(m.body(), &out)
}

// Attribute will return the custom attribute sent with the request.
func (m *message) Attribute(key string) string {
	id, ok := m.message.MessageAttributes[key]
	if !ok {
		return ""
	}

	return id.Value
}

// Attributes will return the custom attributes sent with the request.
func (m *message) Attributes() map[string]string {
	a := map[string]string{}

	for k, v := range m.message.MessageAttributes {
		a[k] = v.Value
	}

	return a
}

// SystemAttributeByKey will return the system attributing by key.
func (m *message) SystemAttributeByKey(key string) string {
	value, ok := m.originalMessage.Attributes[key]
	if !ok {
		return ""
	}

	return value
}

// SystemAttributes will return the system attributes.
func (m *message) SystemAttributes() map[string]string {
	return m.originalMessage.Attributes
}

// Identifier is an identifier associated with the message ReceiptHandle.
func (m *message) Identifier() string {
	return *m.originalMessage.ReceiptHandle
}

// Message returns the body message
func (m *message) Message() string {
	return m.message.Message
}

// DecodeMessage will unmarshal the message into a supplied output using JSON
func (m *message) DecodeMessage(out any) error {
	return json.Unmarshal([]byte(m.message.Message), &out)
}

// TimeStamp returns the message timestamp
func (m *message) TimeStamp() time.Time {
	return m.message.Timestamp
}

// Dispatch sets dispatched as true
func (m *message) Dispatch() {
	m.dispatched <- true
}

// Backoff sets the visibilityTimeout of the message,
// ignoring the visibilityTimeout defined in the route and
// the default visibilityTimeout of the queue
func (m *message) Backoff(delay time.Duration) {
	m.backedOff = true
	m.backoffChannel <- delay
}

// BackedOff returns true if the message was backed off
func (m *message) BackedOff() bool {
	return m.backedOff
}

// Body returns the message body as []byte
func (m *message) Body() []byte {
	return m.body()
}
