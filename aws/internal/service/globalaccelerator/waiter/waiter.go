package waiter

import (
	"time"

	"github.com/aws/aws-sdk-go/service/globalaccelerator"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// AcceleratorDeployed waits for an Accelerator to return Deployed
func AcceleratorDeployed(conn *globalaccelerator.GlobalAccelerator, arn string, timeout time.Duration) (*globalaccelerator.Accelerator, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{globalaccelerator.AcceleratorStatusInProgress},
		Target:  []string{globalaccelerator.AcceleratorStatusDeployed},
		Refresh: AcceleratorStatus(conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*globalaccelerator.Accelerator); ok {
		return v, err
	}

	return nil, err
}
