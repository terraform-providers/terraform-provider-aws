package waiter

import (
	"time"

	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	// Maximum amount of time to wait for an EventSubscription to return Deleted
	EventSubscriptionDeletedTimeout  = 10 * time.Minute
	RdsClusterInitiateUpgradeTimeout = 5 * time.Minute
)

// EventSubscriptionDeleted waits for a EventSubscription to return Deleted
func EventSubscriptionDeleted(conn *rds.RDS, subscriptionName string) (*rds.EventSubscription, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{"deleting"},
		Target:  []string{EventSubscriptionStatusNotFound},
		Refresh: EventSubscriptionStatus(conn, subscriptionName),
		Timeout: EventSubscriptionDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*rds.EventSubscription); ok {
		return v, err
	}

	return nil, err
}

// DBProxyEndpointAvailable waits for a DBProxyEndpoint to return Available
func DBProxyEndpointAvailable(conn *rds.RDS, id string, timeout time.Duration) (*rds.DBProxyEndpoint, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			rds.DBProxyEndpointStatusCreating,
			rds.DBProxyEndpointStatusModifying,
		},
		Target:  []string{rds.DBProxyEndpointStatusAvailable},
		Refresh: DBProxyEndpointStatus(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*rds.DBProxyEndpoint); ok {
		return v, err
	}

	return nil, err
}

// DBProxyEndpointDeleted waits for a DBProxyEndpoint to return Deleted
func DBProxyEndpointDeleted(conn *rds.RDS, id string, timeout time.Duration) (*rds.DBProxyEndpoint, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{rds.DBProxyEndpointStatusDeleting},
		Target:  []string{},
		Refresh: DBProxyEndpointStatus(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*rds.DBProxyEndpoint); ok {
		return v, err
	}

	return nil, err
}
