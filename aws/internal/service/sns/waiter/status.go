package waiter

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func SubscriptionPendingConfirmation(conn *sns.SNS, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := conn.GetSubscriptionAttributes(&sns.GetSubscriptionAttributesInput{
			SubscriptionArn: aws.String(id),
		})

		if tfawserr.ErrCodeEquals(err, sns.ErrCodeResourceNotFoundException) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		if output == nil {
			return nil, "", nil
		}

		return output, aws.StringValue(output.Attributes["PendingConfirmation"]), nil
	}
}
