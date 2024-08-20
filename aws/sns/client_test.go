package sns_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	loafergo "github.com/justcodes/loafer-go/v2"
	loaferAWS "github.com/justcodes/loafer-go/v2/aws"
	"github.com/justcodes/loafer-go/v2/aws/sns"
)

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
			c, err := sns.NewClient(ctx, tc.cfg)
			assert.Nil(t, c)
			assert.ErrorIs(t, err, tc.expectedErr)
		})
	}
}
