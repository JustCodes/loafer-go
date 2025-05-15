package loafergo_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	loafergo "github.com/justcodes/loafer-go/v2"
)

func TestError_Context(t *testing.T) {
	base := loafergo.ErrEmptyParam
	wrapped := base.Context(errors.New("wrapped reason"))

	assert.Equal(t, "required parameter is missing: wrapped reason", wrapped.Error())
	assert.Equal(t, "required parameter is missing", wrapped.Context(nil).Error()) // should not panic with nil
}

func TestPredefinedErrors(t *testing.T) {
	assert.Equal(t, "no routes registered", loafergo.ErrNoRoute.Error())
	assert.Equal(t, "failed to receive messages", loafergo.ErrGetMessage.Error())
	assert.Equal(t, "invalid aws credentials", loafergo.ErrInvalidCreds.Error())
	assert.Equal(t, "unable to marshal request", loafergo.ErrMarshal.Error())
	assert.Equal(t, "sqs client is nil", loafergo.ErrNoSQSClient.Error())
	assert.Equal(t, "handler is nil", loafergo.ErrNoHandler.Error())
	assert.Equal(t, "required parameter is missing", loafergo.ErrEmptyParam.Error())
	assert.Equal(t, "required field is missing", loafergo.ErrEmptyRequiredField.Error())
	assert.Equal(t, "input must be filled", loafergo.ErrEmptyInput.Error())
}
