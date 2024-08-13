package aws

import (
	"context"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"

	loafergo "github.com/justcodes/loafer-go"
)

const (
	defaultRetryCount = 10
	// DataTypeNumber represents the Number datatype, use it when creating custom attributes
	DataTypeNumber = DataType("Number")
	// DataTypeString represents the String datatype, use it when creating custom attributes
	DataTypeString = DataType("String")
)

// Config defines the loafer aws configuration
type Config struct {
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

	// Add custom attributes to the message. This might be a correlationId or client meta information
	// custom attributes will be viewable on the sqs dashboard as metadata
	Attributes []CustomAttribute
}

// ClientConfig defines the loafer aws configuration
type ClientConfig struct {
	Config *Config
	// used to determine how many attempts exponential backoff should use before logging an error
	RetryCount int
	// defines the total amount of goroutines that can be run by the consumer
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
// must use sqs.DataTypeNumber of sqs.DataTypeString for the datatype, the value must match the type provided
func (c *Config) NewCustomAttribute(dataType DataType, title string, value interface{}) error {
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

// ValidateConfig validates client config fields
func ValidateConfig(cfg *ClientConfig) (*ClientConfig, error) {
	if cfg == nil || cfg.Config == nil {
		return nil, loafergo.ErrEmptyParam
	}

	if cfg.Config.Key == "" || cfg.Config.Secret == "" || cfg.Config.Region == "" {
		return nil, loafergo.ErrEmptyRequiredField
	}

	if cfg.RetryCount == 0 {
		cfg.RetryCount = defaultRetryCount
	}

	return cfg, nil
}

// LoadAWSConfig loads aws config
func LoadAWSConfig(ctx context.Context, cfg *ClientConfig, c *aws.CredentialsCache) (aCfg aws.Config, err error) {
	conf := []func(*config.LoadOptions) error{
		config.WithRegion(cfg.Config.Region),
		config.WithCredentialsProvider(c),
		config.WithRetryMaxAttempts(cfg.RetryCount),
	}

	if cfg.Config.Profile != "" {
		conf = append(conf, config.WithSharedConfigProfile(cfg.Config.Profile))
	}

	aCfg, err = config.LoadDefaultConfig(
		ctx,
		conf...,
	)
	if err != nil {
		return
	}

	// if an optional hostname config is provided, then replace the default one
	//
	// This will set the default AWS URL to a hostname of your choice. Perfect for testing, or mocking functionality
	if cfg.Config.Hostname != "" {
		aCfg.BaseEndpoint = aws.String(cfg.Config.Hostname)
	}

	return aCfg, nil
}
