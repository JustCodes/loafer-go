package sqs

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/sqs"

	loafergo "github.com/justcodes/loafer-go/v2"
	loaferAWS "github.com/justcodes/loafer-go/v2/aws"
)

// NewClient instantiates a new sqs client to be used on the sqs route
func NewClient(ctx context.Context, cfg *loaferAWS.ClientConfig) (client loafergo.SQSClient, err error) {
	cfg, err = loaferAWS.ValidateConfig(cfg)
	if err != nil {
		return nil, err
	}

	aCfg, err := loaferAWS.LoadAWSConfig(ctx, cfg, nil)
	if err != nil {
		return nil, err
	}
	client = sqs.NewFromConfig(aCfg)
	return
}
