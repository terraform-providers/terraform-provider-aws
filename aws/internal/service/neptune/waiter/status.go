package waiter

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/neptune"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	// EventSubscription NotFound
	EventSubscriptionStatusNotFound = "NotFound"

	// EventSubscription Unknown
	EventSubscriptionStatusUnknown = "Unknown"
)

// EventSubscriptionStatus fetches the EventSubscription and its Status
func EventSubscriptionStatus(conn *neptune.Neptune, subscriptionName string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &neptune.DescribeEventSubscriptionsInput{
			SubscriptionName: aws.String(subscriptionName),
		}

		output, err := conn.DescribeEventSubscriptions(input)

		if err != nil {
			return nil, EventSubscriptionStatusUnknown, err
		}

		if len(output.EventSubscriptionsList) == 0 {
			return nil, EventSubscriptionStatusNotFound, nil
		}

		return output.EventSubscriptionsList[0], aws.StringValue(output.EventSubscriptionsList[0].Status), nil
	}
}
