package aws_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	awsAWS "github.com/aws/aws-sdk-go-v2/aws"

	loafergo "github.com/justcodes/loafer-go/v2"
	"github.com/justcodes/loafer-go/v2/aws"
)

func TestDataType_String(t *testing.T) {
	d := aws.DataType("Number")
	assert.Equal(t, "Number", d.String())
}

func TestConfig_NewCustomAttribute(t *testing.T) {
	t.Run("With data type string", func(t *testing.T) {
		got := &aws.Config{}
		want := []aws.CustomAttribute{{
			Title:    "title",
			DataType: "String",
			Value:    "my title",
		}}
		err := got.NewCustomAttribute(aws.DataTypeString, "title", "my title")
		assert.NoError(t, err)
		assert.Equal(t, want, got.Attributes)
	})

	t.Run("With data type string error", func(t *testing.T) {
		got := aws.Config{}
		err := got.NewCustomAttribute(aws.DataTypeString, "title", 1.6)
		assert.NotNil(t, err)
		assert.ErrorIs(t, loafergo.ErrMarshal, err)
	})

	t.Run("With data type number", func(t *testing.T) {
		got := &aws.Config{}
		want := []aws.CustomAttribute{{
			Title:    "title",
			DataType: "Number",
			Value:    "1",
		}}
		err := got.NewCustomAttribute(aws.DataTypeNumber, "title", 1)
		assert.NoError(t, err)
		assert.Equal(t, want, got.Attributes)
	})

	t.Run("With data type number error", func(t *testing.T) {
		got := &aws.Config{}
		err := got.NewCustomAttribute(aws.DataTypeNumber, "title", 1.6)
		assert.NotNil(t, err)
		assert.ErrorIs(t, loafergo.ErrMarshal, err)
	})
}

func TestSQSClientValidateConfig(t *testing.T) {
	testsCases := []struct {
		err      error
		cfg      *aws.ClientConfig
		expected *aws.ClientConfig
		name     string
	}{
		{
			name: "valid with retry default",
			cfg: &aws.ClientConfig{
				Config: &aws.Config{
					Key:    "dummy",
					Secret: "dummy",
					Region: "us-east-1",
				},
			},
			expected: &aws.ClientConfig{
				Config: &aws.Config{
					Key:    "dummy",
					Secret: "dummy",
					Region: "us-east-1",
				},
				RetryCount: 10,
			},
		},
		{
			name: "valid with custom retry",
			cfg: &aws.ClientConfig{
				Config: &aws.Config{
					Key:    "dummy",
					Secret: "dummy",
					Region: "us-east-1",
				},
				RetryCount: 42,
			},
			expected: &aws.ClientConfig{
				Config: &aws.Config{
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
			cfg:      &aws.ClientConfig{},
			err:      loafergo.ErrEmptyParam,
			expected: nil,
		},
		{
			name: "empty AWS region",
			cfg: &aws.ClientConfig{
				Config: &aws.Config{
					Key:    "dummy",
					Secret: "dummy",
				},
			},
			err:      loafergo.ErrEmptyRequiredField,
			expected: nil,
		},
	}
	for _, tc := range testsCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := aws.ValidateConfig(tc.cfg)
			assert.Equal(t, tc.err, err)
			assert.Equal(t, tc.expected, got)
		})
	}
}

func TestLoadAWSConfig(t *testing.T) {
	ctx := context.Background()
	t.Run("Success", func(t *testing.T) {
		cfg := &aws.ClientConfig{
			Config: &aws.Config{
				Key:      "key",
				Secret:   "secret",
				Region:   "us-east-1",
				Profile:  "",
				Hostname: "",
			},
			RetryCount: 0,
		}

		got, err := aws.LoadAWSConfig(ctx, cfg, &awsAWS.CredentialsCache{})
		assert.Nil(t, err)
		assert.NotNil(t, got)
	})

	t.Run("Error", func(t *testing.T) {
		cfg := &aws.ClientConfig{
			Config: &aws.Config{
				Key:      "key",
				Secret:   "secret",
				Region:   "us-east-1",
				Profile:  "profile",
				Hostname: "",
			},
			RetryCount: 0,
		}

		_, err := aws.LoadAWSConfig(ctx, cfg, &awsAWS.CredentialsCache{})
		assert.NotNil(t, err)
	})
}
