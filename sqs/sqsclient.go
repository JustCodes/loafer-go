package sqs

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sqs"

	loafergo "github.com/justcodes/loafer-go"
)

// NewSQSClient instantiates a new sqs client to be used on the sqs route
func NewSQSClient(ctx context.Context, cfg *ClientConfig) (client loafergo.SQSClient, err error) {
	c := aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider(cfg.AwsConfig.Key, cfg.AwsConfig.Secret, ""))
	_, err = c.Retrieve(ctx)
	if err != nil {
		return client, loafergo.ErrInvalidCreds.Context(err)
	}

	conf := []func(*config.LoadOptions) error{
		config.WithRegion(cfg.AwsConfig.Region),
		config.WithCredentialsProvider(c),
		config.WithRetryMaxAttempts(cfg.RetryCount),
	}

	if cfg.AwsConfig.Profile != "" {
		conf = append(conf, config.WithSharedConfigProfile(cfg.AwsConfig.Profile))
	}

	aCfg, cfgErr := config.LoadDefaultConfig(
		ctx,
		conf...,
	)
	if cfgErr != nil {
		return client, cfgErr
	}

	// if an optional hostname config is provided, then replace the default one
	//
	// This will set the default AWS URL to a hostname of your choice. Perfect for testing, or mocking functionality
	if cfg.AwsConfig.Hostname != "" {
		aCfg.BaseEndpoint = aws.String(cfg.AwsConfig.Hostname)
	}

	client = sqs.NewFromConfig(aCfg)
	return
}
