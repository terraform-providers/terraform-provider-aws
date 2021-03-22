package waiter

import (
	"time"

	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	// Maximum amount of time to wait for an InstanceRefresh to be started
	// Must be at least as long as InstanceRefreshCancelledTimeout, since we try to cancel any
	// existing Instance Refreshes when starting.
	InstanceRefreshStartedTimeout = InstanceRefreshCancelledTimeout

	// Maximum amount of time to wait for an Instance Refresh to be Cancelled
	InstanceRefreshCancelledTimeout = 15 * time.Minute
)

func InstanceRefreshCancelled(conn *autoscaling.AutoScaling, asgName, instanceRefreshId string) (*autoscaling.InstanceRefresh, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			autoscaling.InstanceRefreshStatusPending,
			autoscaling.InstanceRefreshStatusInProgress,
			autoscaling.InstanceRefreshStatusCancelling,
		},
		Target: []string{
			autoscaling.InstanceRefreshStatusCancelled,
			// Failed and Successful are also acceptable end-states
			autoscaling.InstanceRefreshStatusFailed,
			autoscaling.InstanceRefreshStatusSuccessful,
		},
		Refresh: InstanceRefreshStatus(conn, asgName, instanceRefreshId),
		Timeout: InstanceRefreshCancelledTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*autoscaling.InstanceRefresh); ok {
		return v, err
	}

	return nil, err
}
