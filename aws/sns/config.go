package sns

import (
	loafergo "github.com/justcodes/loafer-go"
)

// A Config provides service configuration for SNS producer.
type Config struct {
	SNSClient loafergo.SNSClient
}

func validateConfig(c *Config) error {
	if c == nil {
		return loafergo.ErrEmptyParam
	}

	if c.SNSClient == nil {
		return loafergo.ErrEmptyRequiredField
	}

	return nil
}
