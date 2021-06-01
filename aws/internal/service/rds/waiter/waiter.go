package waiter

import (
	"time"

	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	// Maximum amount of time to wait for an EventSubscription to return Deleted
	EventSubscriptionDeletedTimeout = 10 * time.Minute
)

// DeploymentDeployed waits for a EventSubscription to return Deleted
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
