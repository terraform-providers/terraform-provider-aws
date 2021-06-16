package waiter

import (
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/ecs/finder"
)

const (
	// EventSubscription NotFound
	CapacityProviderStatusNotFound = "NotFound"

	// EventSubscription Unknown
	CapacityProviderStatusUnknown = "Unknown"

	// AWS will likely add consts for these at some point
	ServiceStatusInactive = "INACTIVE"
	ServiceStatusActive   = "ACTIVE"
	ServiceStatusDraining = "DRAINING"

	ServiceStatusError = "ERROR"
	ServiceStatusNone  = "NONE"

	ClusterStatusError = "ERROR"
	ClusterStatusNone  = "NONE"
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

func ServiceStatus(conn *ecs.ECS, id, cluster string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &ecs.DescribeServicesInput{
			Services: aws.StringSlice([]string{id}),
			Cluster:  aws.String(cluster),
		}

		output, err := conn.DescribeServices(input)

		if tfawserr.ErrCodeEquals(err, ecs.ErrCodeServiceNotFoundException) {
			return nil, ServiceStatusNone, nil
		}

		if err != nil {
			return nil, ServiceStatusError, err
		}

		if len(output.Services) == 0 {
			return nil, ServiceStatusNone, nil
		}

		log.Printf("[DEBUG] ECS service (%s) is currently %q", id, *output.Services[0].Status)
		return output, aws.StringValue(output.Services[0].Status), err
	}
}

func ClusterStatus(conn *ecs.ECS, arn string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := finder.ClusterByARN(conn, arn)

		if tfawserr.ErrCodeEquals(err, ecs.ErrCodeClusterNotFoundException) {
			return nil, ClusterStatusNone, nil
		}

		if err != nil {
			return nil, ClusterStatusError, err
		}

		if len(output.Clusters) == 0 {
			return nil, ClusterStatusNone, nil
		}

		return output, aws.StringValue(output.Clusters[0].Status), err
	}
}
