package sqs_test

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/stretchr/testify/assert"

	loafergo "github.com/justcodes/loafer-go"
	loaferAWS "github.com/justcodes/loafer-go/aws"
	"github.com/justcodes/loafer-go/aws/sqs"
)

func TestSQSClientLoadConfig(t *testing.T) {
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
			name: "Empty key",
			cfg: &loaferAWS.ClientConfig{
				Config: &loaferAWS.Config{
					Key:        "",
					Secret:     "secret",
					Region:     "us-east-1",
					Profile:    "profile",
					Hostname:   "hostname",
					Attributes: nil,
				},
			},
			expectedErr: loafergo.ErrEmptyRequiredField,
		},
		{
			name: "Empty Secret",
			cfg: &loaferAWS.ClientConfig{
				Config: &loaferAWS.Config{
					Key:        "key",
					Secret:     "",
					Region:     "us-east-1",
					Profile:    "profile",
					Hostname:   "hostname",
					Attributes: nil,
				},
			},
			expectedErr: loafergo.ErrEmptyRequiredField,
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
