package waiter

import (
	"time"

	"github.com/aws/aws-sdk-go/service/cloud9"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	// Maximum amount of time to wait for an Operation to return Success
	EnvironmentReadyTimeout   = 10 * time.Minute
	EnvironmentDeletedTimeout = 20 * time.Minute
)

// EnvironmentReady waits for an Operation to return Success
func EnvironmentReady(conn *cloud9.Cloud9, id string) (*cloud9.DescribeEnvironmentStatusOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			cloud9.EnvironmentStatusConnecting,
			cloud9.EnvironmentStatusCreating,
		},
		Target:  []string{cloud9.EnvironmentStatusReady},
		Refresh: EnvironmentStatus(conn, id),
		Timeout: EnvironmentReadyTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*cloud9.DescribeEnvironmentStatusOutput); ok {
		return output, err
	}

	return nil, err
}

// EnvironmentDeleted waits for an Operation to return Success
func EnvironmentDeleted(conn *cloud9.Cloud9, id string) (*cloud9.DescribeEnvironmentStatusOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			cloud9.EnvironmentStatusStopping,
			cloud9.EnvironmentStatusStopped,
			cloud9.EnvironmentStatusDeleting,
		},
		Target:  []string{},
		Refresh: EnvironmentStatus(conn, id),
		Timeout: EnvironmentDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*cloud9.DescribeEnvironmentStatusOutput); ok {
		return output, err
	}

	return nil, err
}
