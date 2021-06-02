package waiter

import (
	"time"

	"github.com/aws/aws-sdk-go/service/sfn"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	StateMachineCreatedTimeout = 5 * time.Minute
	StateMachineDeletedTimeout = 5 * time.Minute
	StateMachineUpdatedTimeout = 1 * time.Minute
)

func StateMachineDeleted(conn *sfn.SFN, stateMachineArn string) (*sfn.DescribeStateMachineOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{sfn.StateMachineStatusActive, sfn.StateMachineStatusDeleting},
		Target:  []string{},
		Refresh: StateMachineStatus(conn, stateMachineArn),
		Timeout: StateMachineDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*sfn.DescribeStateMachineOutput); ok {
		return output, err
	}

	return nil, err
}
