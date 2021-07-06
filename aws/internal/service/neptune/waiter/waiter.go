package waiter

import (
	"time"

	"github.com/aws/aws-sdk-go/service/neptune"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	// Maximum amount of time to wait for an EventSubscription to return Deleted
	EventSubscriptionDeletedTimeout = 10 * time.Minute

	// Maximum amount of time to wait for an DBClusterEndpoint to return Available
	DBClusterEndpointAvailableTimeout = 10 * time.Minute

	// Maximum amount of time to wait for an DBClusterEndpoint to return Deleted
	DBClusterEndpointDeletedTimeout = 10 * time.Minute
)

// EventSubscriptionDeleted waits for a EventSubscription to return Deleted
func EventSubscriptionDeleted(conn *neptune.Neptune, subscriptionName string) (*neptune.EventSubscription, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{"deleting"},
		Target:  []string{EventSubscriptionStatusNotFound},
		Refresh: EventSubscriptionStatus(conn, subscriptionName),
		Timeout: EventSubscriptionDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*neptune.EventSubscription); ok {
		return v, err
	}

	return nil, err
}

// DBClusterDeleted waits for a Cluster to return Deleted
func DBClusterDeleted(conn *neptune.Neptune, id string, timeout time.Duration) (*neptune.DBCluster, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			"available",
			"deleting",
			"backing-up",
			"modifying",
		},
		Target:     []string{ClusterStatusNotFound},
		Refresh:    ClusterStatus(conn, id),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*neptune.DBCluster); ok {
		return v, err
	}

	return nil, err
}

// DBClusterAvailable waits for a Cluster to return Available
func DBClusterAvailable(conn *neptune.Neptune, id string, timeout time.Duration) (*neptune.DBCluster, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			"creating",
			"backing-up",
			"modifying",
			"preparing-data-migration",
			"migrating",
			"configuring-iam-database-auth",
		},
		Target:     []string{"available"},
		Refresh:    ClusterStatus(conn, id),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*neptune.DBCluster); ok {
		return v, err
	}

	return nil, err
}

// DBClusterEndpointAvailable waits for a DBClusterEndpoint to return Available
func DBClusterEndpointAvailable(conn *neptune.Neptune, id string) (*neptune.DBClusterEndpoint, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{"creating", "modifying"},
		Target:  []string{"available"},
		Refresh: DBClusterEndpointStatus(conn, id),
		Timeout: DBClusterEndpointAvailableTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*neptune.DBClusterEndpoint); ok {
		return v, err
	}

	return nil, err
}

// DBClusterEndpointDeleted waits for a DBClusterEndpoint to return Deleted
func DBClusterEndpointDeleted(conn *neptune.Neptune, id string) (*neptune.DBClusterEndpoint, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{"deleting"},
		Target:  []string{},
		Refresh: DBClusterEndpointStatus(conn, id),
		Timeout: DBClusterEndpointDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*neptune.DBClusterEndpoint); ok {
		return v, err
	}

	return nil, err
}
