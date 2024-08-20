package sqs_test

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/stretchr/testify/assert"

	loafergo "github.com/justcodes/loafer-go/v2"
	loaferAWS "github.com/justcodes/loafer-go/v2/aws"
	"github.com/justcodes/loafer-go/v2/aws/sqs"
)

func TestSQSClientLoadConfig(t *testing.T) {
	t.Run("LoadAWSConfig with credentials", func(t *testing.T) {
		acc := &aws.CredentialsCache{}
		cfg := &loaferAWS.ClientConfig{
			Config: &loaferAWS.Config{
				Key:      "dummy",
				Secret:   "dummy",
				Region:   "us-east-1",
				Hostname: "",
			},
		}
		ctx := context.Background()

		got, err := loaferAWS.LoadAWSConfig(ctx, cfg, acc)
		assert.NoError(t, err)
		assert.NotNil(t, got)
	})

	t.Run("LoadAWSConfig without credentials", func(t *testing.T) {
		cfg := &loaferAWS.ClientConfig{
			Config: &loaferAWS.Config{
				Region: "us-east-1",
			},
		}
		ctx := context.Background()

		got, err := loaferAWS.LoadAWSConfig(ctx, cfg, nil)
		assert.NoError(t, err)
		assert.NotNil(t, got)
	})
}

func TestNewProducer_ValidateConfig(t *testing.T) {
	ctx := context.Background()
	testsCases := []struct {
		name        string
		cfg         *loaferAWS.ClientConfig
		expectedErr error
	}{
		{
			name:        "Config nil",
			cfg:         nil,
			expectedErr: loafergo.ErrEmptyParam,
		},
		{
			name: "Aws config nil",
			cfg: &loaferAWS.ClientConfig{
				Config:     nil,
				RetryCount: 0,
			},
			expectedErr: loafergo.ErrEmptyParam,
		},
		{
			name: "Empty Region",
			cfg: &loaferAWS.ClientConfig{
				Config: &loaferAWS.Config{
					Key:        "key",
					Secret:     "secret",
					Region:     "",
					Profile:    "profile",
					Hostname:   "hostname",
					Attributes: nil,
				},
			},
			expectedErr: loafergo.ErrEmptyRequiredField,
		},
	}
	for _, tc := range testsCases {
		t.Run(tc.name, func(t *testing.T) {
			c, err := sqs.NewClient(ctx, tc.cfg)
			assert.Nil(t, c)
			assert.ErrorIs(t, err, tc.expectedErr)
		})
	}
}
