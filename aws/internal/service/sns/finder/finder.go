package finder

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
)

func SubscriptionByARN(conn *sns.SNS, id string) (*sns.GetSubscriptionAttributesOutput, error) {
	output, err := conn.GetSubscriptionAttributes(&sns.GetSubscriptionAttributesInput{
		SubscriptionArn: aws.String(id),
	})
	if tfawserr.ErrCodeEquals(err, sns.ErrCodeNotFoundException) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Attributes == nil || len(output.Attributes) == 0 {
		return nil, nil
	}

	return output, nil
}
