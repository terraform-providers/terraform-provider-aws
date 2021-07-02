package waiter

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/rds/finder"
)

const (
	// EventSubscription NotFound
	EventSubscriptionStatusNotFound = "NotFound"

	// EventSubscription Unknown
	EventSubscriptionStatusUnknown = "Unknown"

	// ProxyEndpoint NotFound
	ProxyEndpointStatusNotFound = "NotFound"

	// ProxyEndpoint Unknown
	ProxyEndpointStatusUnknown = "Unknown"
)

// EventSubscriptionStatus fetches the EventSubscription and its Status
func EventSubscriptionStatus(conn *rds.RDS, subscriptionName string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &rds.DescribeEventSubscriptionsInput{
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

// DBProxyEndpointStatus fetches the ProxyEndpoint and its Status
func DBProxyEndpointStatus(conn *rds.RDS, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := finder.DBProxyEndpoint(conn, id)

		if err != nil {
			return nil, ProxyEndpointStatusUnknown, err
		}

		if output == nil {
			return nil, ProxyEndpointStatusNotFound, nil
		}

		return output, aws.StringValue(output.Status), nil
	}
}

// ActivityStreamStatus fetches the RDS Cluster Activity Stream Status
func ActivityStreamStatus(conn *rds.RDS, dbClusterIdentifier string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		emptyResp := &rds.DescribeDBClustersInput{}

		resp, err := conn.DescribeDBClusters(&rds.DescribeDBClustersInput{
			DBClusterIdentifier: aws.String(dbClusterIdentifier),
		})

		if err != nil {
			log.Printf("[DEBUG] Refreshing RDS Cluster Activity Stream State. Occur error: %s", err)
			if tfawserr.ErrCodeContains(err, rds.ErrCodeDBClusterNotFoundFault) {
				return emptyResp, rds.ActivityStreamStatusStopped, nil
			} else if resp != nil && len(resp.DBClusters) == 0 {
				return emptyResp, rds.ActivityStreamStatusStopped, nil
			} else {
				return emptyResp, "", fmt.Errorf("error on refresh: %+v", err)
			}
		}

		if resp == nil || resp.DBClusters == nil || len(resp.DBClusters) == 0 {
			log.Printf("[DEBUG] Refreshing RDS Cluster Activity Stream State. Invalid resp: %s", resp)
			return emptyResp, rds.ActivityStreamStatusStopped, nil
		}

		cluster := resp.DBClusters[0]
		status := aws.StringValue(cluster.ActivityStreamStatus)
		log.Printf("[DEBUG] Refreshing RDS Cluster Activity Stream State... %s", status)
		return cluster, status, nil
	}
}
