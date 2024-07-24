package sqs

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sqs"

	loafergo "github.com/justcodes/loafer-go"
)

const defaultRetryCount = 10

// NewSQSClient instantiates a new sqs client to be used on the sqs route
func NewSQSClient(ctx context.Context, cfg *ClientConfig) (client loafergo.SQSClient, err error) {
	cfg, err = validateConfig(cfg)
	if err != nil {
		return nil, err
	}

	c := aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider(cfg.AwsConfig.Key, cfg.AwsConfig.Secret, ""))
	_, err = c.Retrieve(ctx)
	if err != nil {
		return client, loafergo.ErrInvalidCreds.Context(err)
	}

	aCfg, err := loadAWSConfig(ctx, cfg, c)
	if err != nil {
		return nil, err
	}
	client = sqs.NewFromConfig(aCfg)
	return
}

func loadAWSConfig(ctx context.Context, cfg *ClientConfig, c *aws.CredentialsCache) (aCfg aws.Config, err error) {
	conf := []func(*config.LoadOptions) error{
		config.WithRegion(cfg.AwsConfig.Region),
		config.WithCredentialsProvider(c),
		config.WithRetryMaxAttempts(cfg.RetryCount),
	}

	if cfg.AwsConfig.Profile != "" {
		conf = append(conf, config.WithSharedConfigProfile(cfg.AwsConfig.Profile))
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
	if cfg.AwsConfig.Hostname != "" {
		aCfg.BaseEndpoint = aws.String(cfg.AwsConfig.Hostname)
	}

	return aCfg, nil
}

func validateConfig(cfg *ClientConfig) (*ClientConfig, error) {
	if cfg == nil || cfg.AwsConfig == nil {
		return nil, loafergo.ErrEmptyParam
	}

	if cfg.AwsConfig.Key == "" || cfg.AwsConfig.Secret == "" || cfg.AwsConfig.Region == "" {
		return nil, loafergo.ErrEmptyRequiredField
	}

	if cfg.RetryCount == 0 {
		cfg.RetryCount = defaultRetryCount
	}

	return cfg, nil
}
