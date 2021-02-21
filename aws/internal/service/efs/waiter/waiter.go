package waiter

import (
	"time"

	"github.com/aws/aws-sdk-go/service/efs"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	// Maximum amount of time to wait for an Operation to return Success
	AccessPointCreatedTimeout = 10 * time.Minute
	AccessPointDeletedTimeout = 10 * time.Minute
)

// AccessPointCreated waits for an Operation to return Success
func AccessPointCreated(conn *efs.EFS, accessPointId string) (*efs.AccessPointDescription, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{efs.LifeCycleStateCreating},
		Target:  []string{efs.LifeCycleStateAvailable},
		Refresh: AccessPointLifeCycleState(conn, accessPointId),
		Timeout: AccessPointCreatedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*efs.AccessPointDescription); ok {
		return output, err
	}

	return nil, err
}

// AccessPointDelete waits for an Access Point to return Deleted
func AccessPointDeleted(conn *efs.EFS, accessPointId string) (*efs.AccessPointDescription, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{efs.LifeCycleStateAvailable, efs.LifeCycleStateDeleting, efs.LifeCycleStateDeleted},
		Target:  []string{},
		Refresh: AccessPointLifeCycleState(conn, accessPointId),
		Timeout: AccessPointDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*efs.AccessPointDescription); ok {
		return output, err
	}

	return nil, err
}
