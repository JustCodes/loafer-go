package sqs

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/stretchr/testify/assert"

	loafergo "github.com/justcodes/loafer-go"
)

func TestSQSClientLoadConfig(t *testing.T) {
	acc := &aws.CredentialsCache{}
	cfg := &ClientConfig{
		AwsConfig: &AWSConfig{
			Key:      "dummy",
			Secret:   "dummy",
			Region:   "us-east-1",
			Hostname: "",
		},
	}
	ctx := context.Background()

	got, err := loadAWSConfig(ctx, cfg, acc)
	assert.NoError(t, err)
	assert.NotNil(t, got)
}

func TestSQSClientValidateConfig(t *testing.T) {
	testsCases := []struct {
		name     string
		cfg      *ClientConfig
		err      error
		expected *ClientConfig
	}{
		{
			name: "valid with retry default",
			cfg: &ClientConfig{
				AwsConfig: &AWSConfig{
					Key:    "dummy",
					Secret: "dummy",
					Region: "us-east-1",
				},
			},
			expected: &ClientConfig{
				AwsConfig: &AWSConfig{
					Key:    "dummy",
					Secret: "dummy",
					Region: "us-east-1",
				},
				RetryCount: defaultRetryCount,
			},
		},
		{
			name: "valid with custom retry",
			cfg: &ClientConfig{
				AwsConfig: &AWSConfig{
					Key:    "dummy",
					Secret: "dummy",
					Region: "us-east-1",
				},
				RetryCount: 42,
			},
			expected: &ClientConfig{
				AwsConfig: &AWSConfig{
					Key:    "dummy",
					Secret: "dummy",
					Region: "us-east-1",
				},
				RetryCount: 42,
			},
		},
		{
			name:     "invalid",
			cfg:      nil,
			err:      loafergo.ErrEmptyParam,
			expected: nil,
		},
		{
			name:     "invalid AWS config",
			cfg:      &ClientConfig{},
			err:      loafergo.ErrEmptyParam,
			expected: nil,
		},
		{
			name: "empty AWS region",
			cfg: &ClientConfig{
				AwsConfig: &AWSConfig{
					Key:    "dummy",
					Secret: "dummy",
				},
			},
			err:      loafergo.ErrEmptyRequiredField,
			expected: nil,
		},
		{
			name: "empty AWS Key",
			cfg: &ClientConfig{
				AwsConfig: &AWSConfig{
					Secret: "dummy",
					Region: "us-east-1",
				},
			},
			err:      loafergo.ErrEmptyRequiredField,
			expected: nil,
		},
		{
			name: "empty AWS Secret",
			cfg: &ClientConfig{
				AwsConfig: &AWSConfig{
					Key:    "dummy",
					Region: "us-east-1",
				},
			},
			err:      loafergo.ErrEmptyRequiredField,
			expected: nil,
		},
	}
	for _, tc := range testsCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := validateConfig(tc.cfg)
			assert.Equal(t, tc.err, err)
			assert.Equal(t, tc.expected, got)
		})
	}
}
