package sqs

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/stretchr/testify/assert"
)

func TestMessage_Attribute(t *testing.T) {
	t.Run("should return attribute value", func(t *testing.T) {
		msg := newMessage(types.Message{
			MessageAttributes: map[string]types.MessageAttributeValue{
				"foo": {
					DataType:    aws.String("String"),
					StringValue: aws.String("bar"),
				},
			},
		})
		attr := msg.Attribute("foo")
		assert.Equal(t, "bar", attr)
	})

	t.Run("should return empty attribute value", func(t *testing.T) {
		msg := newMessage(types.Message{
			MessageAttributes: map[string]types.MessageAttributeValue{
				"foo": {
					DataType:    aws.String("String"),
					StringValue: aws.String("bar"),
				},
			},
		})
		attr := msg.Attribute("stub")
		assert.Equal(t, "", attr)
	})
}

func TestMessage_Body(t *testing.T) {
	msg := newMessage(types.Message{
		Body: aws.String(`{"foo": "bar"}`),
	})
	body := msg.Body()
	assert.Equal(t, "{\"foo\": \"bar\"}", string(body))
}

func TestMessage_Decode(t *testing.T) {
	type data struct {
		Foo string `json:"foo"`
	}
	d := new(data)
	msg := newMessage(types.Message{
		Body: aws.String(`{"foo": "bar"}`),
	})
	err := msg.Decode(d)
	assert.NoError(t, err)
	assert.Equal(t, "bar", d.Foo)
}

func TestMessage_Metadata(t *testing.T) {
	msg := newMessage(types.Message{
		Attributes: map[string]string{
			"foo": "bar",
		},
	})
	got := msg.Metadata()
	assert.Equal(t, "bar", got["foo"])
}

func TestMessage_Identifier(t *testing.T) {
	msg := newMessage(types.Message{
		ReceiptHandle: aws.String("receipt-handle"),
	})
	got := msg.Identifier()
	assert.Equal(t, "receipt-handle", got)
}

func TestMessage_Dispatch(t *testing.T) {
	msg := newMessage(types.Message{})
	msg.Dispatch()
	assert.True(t, <-msg.dispatched)
}
