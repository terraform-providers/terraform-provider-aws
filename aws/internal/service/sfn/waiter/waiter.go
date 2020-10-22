package waiter

import (
	"time"

	"github.com/aws/aws-sdk-go/service/sfn"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	// Maximum amount of time to wait for an Operation to return Success
	StateMachineDeleteTimeout = 5 * time.Minute
)

// StateMachineDeleted waits for an Operation to return Success
func StateMachineDeleted(conn *sfn.SFN, stateMachineArn string) (*sfn.DescribeStateMachineOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{sfn.StateMachineStatusActive, sfn.StateMachineStatusDeleting},
		Target:  []string{},
		Refresh: StateMachineStatus(conn, stateMachineArn),
		Timeout: StateMachineDeleteTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*sfn.DescribeStateMachineOutput); ok {
		return output, err
	}

	return nil, err
}
