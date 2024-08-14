package sns_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/justcodes/loafer-go/v2/aws/sns"
)

func TestBuildTopicARN(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		arn, err := sns.BuildTopicARN("us-east-1", "00000000", "my_topic_name")
		assert.NoError(t, err)
		assert.Equal(t, "arn:aws:sns:us-east-1:00000000:my_topic_name", arn)
	})

	t.Run("Failure", func(t *testing.T) {
		testsCases := []struct {
			name         string
			region       string
			awsAccountID string
			topicName    string
		}{
			{
				name:         "No Region",
				region:       "",
				awsAccountID: "00000000",
				topicName:    "my_topic_name",
			},
			{
				name:         "No AccountID",
				region:       "us-east-1",
				awsAccountID: "",
				topicName:    "my_topic_name",
			},
			{
				name:         "No TopicName",
				region:       "us-east-1",
				awsAccountID: "00000000",
				topicName:    "",
			},
		}
		for _, tc := range testsCases {
			t.Run(tc.name, func(t *testing.T) {
				got, err := sns.BuildTopicARN(tc.region, tc.awsAccountID, tc.topicName)
				assert.Error(t, err)
				assert.Empty(t, got)
			})
		}
	})
}
