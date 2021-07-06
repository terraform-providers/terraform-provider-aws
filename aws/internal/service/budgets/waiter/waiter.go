package waiter

import (
	"time"

	"github.com/aws/aws-sdk-go/service/budgets"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	ActionAvailableTimeout = 2 * time.Minute
)

func ActionAvailable(conn *budgets.Budgets, id string) (*budgets.Action, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			budgets.ActionStatusExecutionInProgress,
			budgets.ActionStatusStandby,
		},
		Target: []string{
			budgets.ActionStatusExecutionSuccess,
			budgets.ActionStatusExecutionFailure,
			budgets.ActionStatusPending,
		},
		Refresh: ActionStatus(conn, id),
		Timeout: ActionAvailableTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*budgets.Action); ok {
		return v, err
	}

	return nil, err
}
