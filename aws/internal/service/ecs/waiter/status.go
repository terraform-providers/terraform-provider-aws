package waiter

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

const (
	// EventSubscription NotFound
	CapacityProviderStatusNotFound = "NotFound"

	// EventSubscription Unknown
	CapacityProviderStatusUnknown = "Unknown"
)

// CapacityProviderStatus fetches the Capacity Provider and its Status
func CapacityProviderStatus(conn *ecs.ECS, capacityProvider string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &ecs.DescribeCapacityProvidersInput{
			CapacityProviders: aws.StringSlice([]string{capacityProvider}),
		}

		output, err := conn.DescribeCapacityProviders(input)

		if err != nil {
			return nil, CapacityProviderStatusUnknown, err
		}

		if len(output.CapacityProviders) == 0 {
			return nil, CapacityProviderStatusNotFound, nil
		}

		return output.CapacityProviders[0], aws.StringValue(output.CapacityProviders[0].Status), nil
	}
}
