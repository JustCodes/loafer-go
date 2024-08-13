package sns

import (
	"errors"
	"fmt"
)

const topicARNPattern = "arn:aws:sns:%s:%s:%s"

// BuildTopicARN is used to build a topic arn
//
// pattern: "arn:aws:sns:<region>:<aws_account_id>:<topic_name>"
// It returns an error if one of the parameters is empty
func BuildTopicARN(region, awsAccountID, topicName string) (string, error) {
	if region == "" || awsAccountID == "" || topicName == "" {
		return "", errors.New("region, awsAccountID, topicName is empty")
	}
	return fmt.Sprintf(topicARNPattern, region, awsAccountID, topicName), nil
}
