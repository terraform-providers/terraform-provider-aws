package waiter

import (
	"time"

	"github.com/aws/aws-sdk-go/service/servicediscovery"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

const (
	// Maximum amount of time to wait for an Operation to return Success
	OperationSuccessTimeout = 5 * time.Minute
)

// OperationSuccess waits for an Operation to return Success
func OperationSuccess(conn *servicediscovery.ServiceDiscovery, operationID string) (*servicediscovery.GetOperationOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{servicediscovery.OperationStatusSubmitted, servicediscovery.OperationStatusPending},
		Target:  []string{servicediscovery.OperationStatusSuccess},
		Refresh: OperationStatus(conn, operationID),
		Timeout: OperationSuccessTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*servicediscovery.GetOperationOutput); ok {
		return output, err
	}

	return nil, err
}
