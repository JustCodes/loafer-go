package loafergo

import (
	"fmt"
)

// Error represents a typed application error with context.
type Error struct {
	err     error
	message string
}

// Error returns the composed error message.
func (e Error) Error() string {
	if e.err != nil {
		return fmt.Sprintf("%s: %s", e.message, e.err.Error())
	}

	return e.message
}

// Context wraps a base error with context.
func (e Error) Context(err error) Error {
	return Error{
		message: e.message,
		err:     err,
	}
}

// Predefined errors.
var (
	ErrNoRoute            = Error{message: "no routes registered"}
	ErrGetMessage         = Error{message: "failed to receive messages"}
	ErrInvalidCreds       = Error{message: "invalid aws credentials"}
	ErrMarshal            = Error{message: "unable to marshal request"}
	ErrNoSQSClient        = Error{message: "sqs client is nil"}
	ErrNoHandler          = Error{message: "handler is nil"}
	ErrEmptyParam         = Error{message: "required parameter is missing"}
	ErrEmptyRequiredField = Error{message: "required field is missing"}
	ErrEmptyInput         = Error{message: "input must be filled"}
)
