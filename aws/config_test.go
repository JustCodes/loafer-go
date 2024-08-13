package aws_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	loafergo "github.com/justcodes/loafer-go"
	"github.com/justcodes/loafer-go/aws"
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
		name     string
		cfg      *aws.ClientConfig
		err      error
		expected *aws.ClientConfig
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
		{
			name: "empty AWS Key",
			cfg: &aws.ClientConfig{
				Config: &aws.Config{
					Secret: "dummy",
					Region: "us-east-1",
				},
			},
			err:      loafergo.ErrEmptyRequiredField,
			expected: nil,
		},
		{
			name: "empty AWS Secret",
			cfg: &aws.ClientConfig{
				Config: &aws.Config{
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
			got, err := aws.ValidateConfig(tc.cfg)
			assert.Equal(t, tc.err, err)
			assert.Equal(t, tc.expected, got)
		})
	}
}
