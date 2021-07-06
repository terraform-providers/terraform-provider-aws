package waiter

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/neptune"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/neptune/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

const (
	// EventSubscription NotFound
	EventSubscriptionStatusNotFound = "NotFound"

	// EventSubscription Unknown
	EventSubscriptionStatusUnknown = "Unknown"

	// Cluster NotFound
	ClusterStatusNotFound = "NotFound"

	// Cluster Unknown
	ClusterStatusUnknown = "Unknown"

	// DBClusterEndpoint Unknown
	DBClusterEndpointStatusUnknown = "Unknown"
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

// ClusterStatus fetches the Cluster and its Status
func ClusterStatus(conn *neptune.Neptune, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &neptune.DescribeDBClustersInput{
			DBClusterIdentifier: aws.String(id),
		}

		output, err := conn.DescribeDBClusters(input)

		if err != nil {
			return nil, ClusterStatusUnknown, err
		}

		if len(output.DBClusters) == 0 {
			return nil, ClusterStatusNotFound, nil
		}

		cluster := output.DBClusters[0]

		return cluster, aws.StringValue(cluster.Status), nil
	}
}

// DBClusterEndpointStatus fetches the DBClusterEndpoint and its Status
func DBClusterEndpointStatus(conn *neptune.Neptune, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := finder.EndpointById(conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, DBClusterEndpointStatusUnknown, err
		}

		return output, aws.StringValue(output.Status), nil
	}
}
