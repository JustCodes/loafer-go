package loafergo

import (
	"fmt"
)

// SQSError defines the error handler for the loafergo package. SQSError satisfies the error interface and can be
// used safely with other error handlers
type SQSError struct {
	contextErr error
	Err        string `json:"err"`
}

// Error is used for implementing the error interface, and for creating
// a proper error string
func (e *SQSError) Error() string {
	if e.contextErr != nil {
		return fmt.Sprintf("%s: %s", e.Err, e.contextErr.Error())
	}

	return e.Err
}

// Context is used for creating a new instance of the error with the contextual error attached
func (e *SQSError) Context(err error) *SQSError {
	ctxErr := new(SQSError)
	*ctxErr = *e
	ctxErr.contextErr = err

	return ctxErr
}

// newSQSErr creates a new SQS Error
func newSQSErr(msg string) *SQSError {
	e := new(SQSError)
	e.Err = msg
	return e
}

// ErrInvalidCreds invalid credentials
var ErrInvalidCreds = newSQSErr("invalid aws credentials")

// ErrMarshal unable to marshal request
var ErrMarshal = newSQSErr("unable to marshal request")

// ErrNoRoute message received without a route
var ErrNoRoute = newSQSErr("message received without a route")

// ErrGetMessage fires when a request to retrieve messages from sqs fails
var ErrGetMessage = newSQSErr("unable to retrieve message")

// ErrMessageProcessing occurs when a message has exceeded the consumption time limit set by aws SQS
var ErrMessageProcessing = newSQSErr("processing time exceeding limit")

// ErrNoSQSClient occurs when the sqs client is nil
var ErrNoSQSClient = newSQSErr("sqs client is nil")

// ErrNoHandler occurs when the handler is nil
var ErrNoHandler = newSQSErr("handler is nil")

// ErrEmptyParam occurs when the required parameter is missing
var ErrEmptyParam = newSQSErr("required parameter is missing")

// ErrEmptyRequiredField occurs when the required field is missing
var ErrEmptyRequiredField = newSQSErr("required field is missing")
