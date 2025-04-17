package loafergo

import (
	"fmt"
)

// Error defines the error handler for the loafergo package. Error satisfies the error interface and can be
// used safely with other error handlers
type Error struct {
	contextErr error
	Err        string `json:"err"`
}

// Error is used for implementing the error interface and for creating
// a proper error string
func (e *Error) Error() string {
	if e.contextErr != nil {
		return fmt.Sprintf("%s: %s", e.Err, e.contextErr.Error())
	}

	return e.Err
}

// Context is used for creating a new instance of the error with the contextual error attached
func (e *Error) Context(err error) *Error {
	ctxErr := new(Error)
	*ctxErr = *e
	ctxErr.contextErr = err

	return ctxErr
}

// newError creates a new SQS Error
func newError(msg string) *Error {
	e := new(Error)
	e.Err = msg
	return e
}

// ErrInvalidCreds invalid credentials
var ErrInvalidCreds = newError("invalid aws credentials")

// ErrMarshal unable to marshal request
var ErrMarshal = newError("unable to marshal request")

// ErrNoRoute message received without a route
var ErrNoRoute = newError("message received without a route")

// ErrGetMessage fires when a request to retrieve messages from sqs fails
var ErrGetMessage = newError("unable to retrieve message")

// ErrMessageProcessing occurs when a message has exceeded the consumption time limit set by aws SQS
var ErrMessageProcessing = newError("processing time exceeding limit")

// ErrNoSQSClient occurs when the sqs client is nil
var ErrNoSQSClient = newError("sqs client is nil")

// ErrNoHandler occurs when the handler is nil
var ErrNoHandler = newError("handler is nil")

// ErrEmptyParam occurs when the required parameter is missing
var ErrEmptyParam = newError("required parameter is missing")

// ErrEmptyRequiredField occurs when the required field is missing
var ErrEmptyRequiredField = newError("required field is missing")

// ErrEmptyInput occurs when the producer received an empty or nil input
var ErrEmptyInput = newError("input must be filled")
