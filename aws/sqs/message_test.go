package sqs

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/stretchr/testify/assert"
)

var (
	mockBody = `{"Message": "{\"message\": \"Hello world!\", \"topic\": \"my_topic__test2\", \"id\": 19}", "Timestamp": "2024-09-10T17:37:15.404Z", "MessageAttributes": { "foo": {"Type": "String", "Value": "bar"}}}`
)

func TestMessage_Attribute(t *testing.T) {
	t.Run("should return attribute value", func(t *testing.T) {
		msg := newMessage(types.Message{
			Body: &mockBody,
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
			Body: aws.String("body"),
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

func TestMessage_Attributes(t *testing.T) {
	t.Run("should return attributes value", func(t *testing.T) {
		msg := newMessage(types.Message{
			Body: &mockBody,
			MessageAttributes: map[string]types.MessageAttributeValue{
				"foo": {
					DataType:    aws.String("String"),
					StringValue: aws.String("bar"),
				},
			},
		})
		attrs := msg.Attributes()
		expected := map[string]string{
			"foo": "bar",
		}
		assert.Equal(t, expected, attrs)
	})

	t.Run("should return empty attributes value", func(t *testing.T) {
		msg := newMessage(types.Message{
			Body: aws.String("body"),
		})
		attrs := msg.Attributes()
		expected := map[string]string{}
		assert.Equal(t, expected, attrs)
	})
}

func TestMessage_Body(t *testing.T) {
	t.Run("With body", func(t *testing.T) {
		msg := newMessage(types.Message{
			Body: aws.String(`{"foo": "bar"}`),
		})
		body := msg.Body()
		assert.Equal(t, "{\"foo\": \"bar\"}", string(body))
	})

	t.Run("No body", func(t *testing.T) {
		msg := newMessage(types.Message{})
		body := msg.Body()
		assert.Empty(t, string(body))
	})
}

func TestMessage_Decode(t *testing.T) {
	t.Run("With body", func(t *testing.T) {
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
	})

	t.Run("No body", func(t *testing.T) {
		type data struct {
			Foo string `json:"foo"`
		}
		d := new(data)
		msg := newMessage(types.Message{})
		err := msg.Decode(d)
		assert.Error(t, err)
		assert.ErrorContains(t, err, "unexpected end of JSON input")
	})
}

func TestMessage_Metadata(t *testing.T) {
	msg := newMessage(types.Message{Body: &mockBody})
	got := msg.Metadata()
	assert.Equal(t, "bar", got["foo"])
}

func TestMessage_Identifier(t *testing.T) {
	msg := newMessage(types.Message{
		Body:          &mockBody,
		ReceiptHandle: aws.String("receipt-handle"),
	})
	got := msg.Identifier()
	assert.Equal(t, "receipt-handle", got)
}

func TestMessage_Dispatch(t *testing.T) {
	msg := newMessage(types.Message{Body: &mockBody})
	msg.Dispatch()
	assert.True(t, <-msg.dispatched)
}

func TestMessage_Message(t *testing.T) {
	t.Run("With body", func(t *testing.T) {
		msg := newMessage(types.Message{Body: &mockBody})
		got := msg.Message()
		want := "{\"message\": \"Hello world!\", \"topic\": \"my_topic__test2\", \"id\": 19}"
		assert.Equal(t, want, got)
	})

	t.Run("No body", func(t *testing.T) {
		msg := newMessage(types.Message{})
		got := msg.Message()
		assert.Empty(t, got)
	})
}

func TestMessage_TimeStamp(t *testing.T) {
	t.Run("With body", func(t *testing.T) {
		msg := newMessage(types.Message{Body: &mockBody})
		got := msg.TimeStamp()
		want := time.Date(2024, time.September, 10, 17, 37, 15, 404000000, time.UTC)
		assert.Equal(t, want, got)
	})

	t.Run("No body", func(t *testing.T) {
		msg := newMessage(types.Message{})
		got := msg.TimeStamp()
		want := time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC)
		assert.Equal(t, want, got)
	})
}

func TestMessage_MessageDecode(t *testing.T) {
	t.Run("With body", func(t *testing.T) {
		type data struct {
			Message string `json:"message"`
			Topic   string `json:"topic"`
			ID      int    `json:"id"`
		}
		d := new(data)
		msg := newMessage(types.Message{
			Body: &mockBody,
		})
		err := msg.DecodeMessage(d)
		assert.NoError(t, err)
		assert.Equal(t, "Hello world!", d.Message)
		assert.Equal(t, "my_topic__test2", d.Topic)
		assert.Equal(t, 19, d.ID)
	})

	t.Run("No body", func(t *testing.T) {
		type data struct {
			Message string `json:"message"`
			Topic   string `json:"topic"`
			ID      int    `json:"id"`
		}
		d := new(data)
		msg := newMessage(types.Message{})
		err := msg.DecodeMessage(d)
		assert.Error(t, err)
		assert.ErrorContains(t, err, "unexpected end of JSON input")
	})

	t.Run("Message has no expected contract", func(t *testing.T) {
		type data struct {
			Message string `json:"message"`
			Topic   string `json:"topic"`
			ID      int    `json:"id"`
		}
		d := new(data)
		msg := newMessage(types.Message{
			Body: aws.String(`{"Message": "{\"foo\": \"bar\"}"}`),
		})
		err := msg.DecodeMessage(d)
		assert.NoError(t, err)
		assert.Equal(t, new(data), d)
	})
}

func TestMessage_SystemAttributes(t *testing.T) {
	msg := newMessage(types.Message{Body: &mockBody, Attributes: map[string]string{
		"MessageGroupId": "xpto",
	}})
	got := msg.SystemAttributes()
	want := map[string]string{
		"MessageGroupId": "xpto",
	}
	assert.Equal(t, want, got)
}

func TestMessage_SystemAttributeByKey(t *testing.T) {
	t.Run("should return system attribute value", func(t *testing.T) {
		msg := newMessage(types.Message{
			Body: &mockBody,
			Attributes: map[string]string{
				"MessageGroupId": "xpto",
			},
		})
		attr := msg.SystemAttributeByKey("MessageGroupId")
		assert.Equal(t, "xpto", attr)
	})

	t.Run("should return empty system attribute value", func(t *testing.T) {
		msg := newMessage(types.Message{
			Body: aws.String("body"),
			Attributes: map[string]string{
				"MessageGroupId": "xpto",
			},
		})
		attr := msg.Attribute("stub")
		assert.Equal(t, "", attr)
	})
}
