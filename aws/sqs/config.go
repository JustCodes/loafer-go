package sqs

import (
	"strconv"

	loafergo "github.com/justcodes/loafer-go/v2"
)

// A Config provides service configuration for SQS routes.
type Config struct {
	SQSClient loafergo.SQSClient
	Handler   loafergo.Handler
	QueueName string
}

const (
	defaultExtensionLimit    = 2
	defaultVisibilityTimeout = int32(30)
	defaultMaxMessages       = int32(10)
	defaultWaitTimeSeconds   = int32(10)
	defaultWorkerPoolSize    = int32(5)
)

// RouteConfig are a discrete set of route options that are valid for loading the route configuration
type RouteConfig struct {
	visibilityTimeout int32
	maxMessages       int32
	extensionLimit    int
	waitTimeSeconds   int32
	workerPoolSize    int32
}

func loadDefaultRouteConfig() *RouteConfig {
	return &RouteConfig{
		visibilityTimeout: defaultVisibilityTimeout,
		maxMessages:       defaultMaxMessages,
		extensionLimit:    defaultExtensionLimit,
		waitTimeSeconds:   defaultWaitTimeSeconds,
		workerPoolSize:    defaultWorkerPoolSize,
	}
}

// LoadRouteConfigFunc is a type alias for RouteConfig functional config
type LoadRouteConfigFunc func(config *RouteConfig)

// RouteWithVisibilityTimeout is a helper function to construct functional options that sets visibility Timeout value
// on config's Route. If multiple RouteWithVisibilityTimeout calls are made,
// the last call overrides the previous call values.
//
// The minimum value is 11 seconds (defaultVisibilityTimeoutControl + 1)
// This value is used to extend the visibility timeout of the message
//
//	to avoid other consumers from consuming this message while it is being processed.
//
// It will extend it periodically based on the visibility timeout value provided,
// and at each iteration the sleep time will be doubled.
//
// For example,
//
//   - queue visibility timeout = 60 seconds (value defined in aws)
//   - route visibility timeout = 30 seconds
//   - time to process the message = 70 seconds
//   - sleep time = 20 seconds (30 seconds - defaultVisibilityTimeoutControl)
//
// --------------------------------------
//   - sleep 20s
//
// 1st iteration:
//
//   - change visibility timeout = 30 seconds
//
//   - sleep 20s
//
// 2nd iteration:
//
//   - change visibility timeout = 60 seconds
//
//   - handler finishes processing the message
//
//   - error handling the message? Message does not get deleted, and the queue visibility timeout (60s) is used (default aws sqs behavior)
//
//   - success handling the message? Message gets deleted
//
// end
func RouteWithVisibilityTimeout(v int32) LoadRouteConfigFunc {
	return func(rc *RouteConfig) {
		if v <= defaultVisibilityTimeoutControl {
			rc.visibilityTimeout = defaultVisibilityTimeoutControl + 1
			return
		}
		rc.visibilityTimeout = v
	}
}

// RouteWithMaxMessages is a helper function to construct functional options that sets Max Messages value
// on config's Route. If multiple RouteWithMaxMessages calls are made,
// the last call overrides the previous call values.
func RouteWithMaxMessages(v int32) LoadRouteConfigFunc {
	return func(rc *RouteConfig) {
		rc.maxMessages = v
	}
}

// RouteWithWaitTimeSeconds is a helper function to construct functional options that sets Wait Time Seconds value
// on config's Route. If multiple RouteWithWaitTimeSeconds calls are made,
// the last call overrides the previous call values.
func RouteWithWaitTimeSeconds(v int32) LoadRouteConfigFunc {
	return func(rc *RouteConfig) {
		rc.waitTimeSeconds = v
	}
}

// RouteWithWorkerPoolSize is a helper function to construct functional options that sets Worker Pool Size value
// on config's Route. If multiple RouteWithWorkerPoolSize calls are made,
// the last call overrides the previous call values.
func RouteWithWorkerPoolSize(v int32) LoadRouteConfigFunc {
	return func(rc *RouteConfig) {
		rc.workerPoolSize = v
	}
}

// AWSConfig defines the loafer aws configuration
type AWSConfig struct {
	// private key to access aws
	Key string
	// secret to access aws
	Secret string
	// region for aws and used for determining the region
	Region string
	// profile for aws and used for determining the profile
	Profile string
	// provided automatically by aws, but must be set for emulators or local testing
	Hostname string
	// used to determine how many attempts exponential backoff should use before logging an error

	// Add custom attributes to the message. This might be a correlationId or client meta-information
	// custom attributes will be viewable on the sqs dashboard as metadata
	Attributes []CustomAttribute
}

// ClientConfig defines the loafer aws configuration
type ClientConfig struct {
	AwsConfig *AWSConfig
	// used to determine how many attempts exponential backoff should use before logging an error
	RetryCount int
}

// CustomAttribute add custom attributes to SNS and SQS messages.
// This can include correlationIds, or any additional information you would like
// separate from the payload body. These attributes can be easily seen from the SQS console.
type CustomAttribute struct {
	Title string
	// Use sqs.DataTypeNumber or sqs.DataTypeString
	DataType string
	// Value represents the value
	Value string
}

// NewCustomAttribute adds a custom attribute to SNS and SQS messages.
// This can include correlationIds, logIds, or any additional information you would like
// separate from the payload body. These attributes can be easily seen from the SQS console.
//
// Must use sqs.DataTypeNumber of sqs.DataTypeString for the datatype, the value must match the type provided
func (c *AWSConfig) NewCustomAttribute(dataType DataType, title string, value interface{}) error {
	if dataType == DataTypeNumber {
		val, ok := value.(int)
		if !ok {
			return loafergo.ErrMarshal
		}

		c.Attributes = append(c.Attributes, CustomAttribute{title, dataType.String(), strconv.Itoa(val)})
		return nil
	}

	val, ok := value.(string)
	if !ok {
		return loafergo.ErrMarshal
	}
	c.Attributes = append(c.Attributes, CustomAttribute{title, dataType.String(), val})
	return nil
}

// DataType is an alias to string
type DataType string

// String returns DataType as a string
func (dt DataType) String() string {
	return string(dt)
}

// DataTypeNumber represents the Number datatype, use it when creating custom attributes
const DataTypeNumber = DataType("Number")

// DataTypeString represents the String datatype, use it when creating custom attributes
const DataTypeString = DataType("String")
