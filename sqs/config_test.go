package sqs_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	loafergo "github.com/justcodes/loafer-go"
	"github.com/justcodes/loafer-go/sqs"
)

func TestDataType_String(t *testing.T) {
	d := sqs.DataType("Number")
	assert.Equal(t, "Number", d.String())
}

func TestConfig_NewCustomAttribute(t *testing.T) {
	t.Run("With data type string", func(t *testing.T) {
		got := &sqs.AWSConfig{}
		want := []sqs.CustomAttribute{{
			Title:    "title",
			DataType: "String",
			Value:    "my title",
		}}
		err := got.NewCustomAttribute(sqs.DataTypeString, "title", "my title")
		assert.NoError(t, err)
		assert.Equal(t, want, got.Attributes)
	})

	t.Run("With data type string error", func(t *testing.T) {
		got := sqs.AWSConfig{}
		err := got.NewCustomAttribute(sqs.DataTypeString, "title", 1.6)
		assert.NotNil(t, err)
		assert.ErrorIs(t, loafergo.ErrMarshal, err)
	})

	t.Run("With data type number", func(t *testing.T) {
		got := &sqs.AWSConfig{}
		want := []sqs.CustomAttribute{{
			Title:    "title",
			DataType: "Number",
			Value:    "1",
		}}
		err := got.NewCustomAttribute(sqs.DataTypeNumber, "title", 1)
		assert.NoError(t, err)
		assert.Equal(t, want, got.Attributes)
	})

	t.Run("With data type number error", func(t *testing.T) {
		got := &sqs.AWSConfig{}
		err := got.NewCustomAttribute(sqs.DataTypeNumber, "title", 1.6)
		assert.NotNil(t, err)
		assert.ErrorIs(t, loafergo.ErrMarshal, err)
	})
}
