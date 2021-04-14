package waiter

import (
	"time"

	"github.com/aws/aws-sdk-go/service/mwaa"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	// Maximum amount of time to wait for an environment creation
	EnvironmentCreatedTimeout = 90 * time.Minute

	// Maximum amount of time to wait for an environment update
	EnvironmentUpdatedTimeout = 90 * time.Minute

	// Maximum amount of time to wait for an environment deletion
	EnvironmentDeletedTimeout = 90 * time.Minute
)

// EnvironmentCreated waits for a Environment to return AVAILABLE
func EnvironmentCreated(conn *mwaa.MWAA, name string) (*mwaa.Environment, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{mwaa.EnvironmentStatusCreating},
		Target:  []string{mwaa.EnvironmentStatusAvailable},
		Refresh: EnvironmentStatus(conn, name),
		Timeout: EnvironmentCreatedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*mwaa.Environment); ok {
		return v, err
	}

	return nil, err
}

// EnvironmentUpdated waits for a Environment to return AVAILABLE
func EnvironmentUpdated(conn *mwaa.MWAA, name string) (*mwaa.Environment, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{mwaa.EnvironmentStatusUpdating},
		Target:  []string{mwaa.EnvironmentStatusAvailable},
		Refresh: EnvironmentStatus(conn, name),
		Timeout: EnvironmentUpdatedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*mwaa.Environment); ok {
		return v, err
	}

	return nil, err
}

// EnvironmentDeleted waits for a Environment to be deleted
func EnvironmentDeleted(conn *mwaa.MWAA, name string) (*mwaa.Environment, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{mwaa.EnvironmentStatusDeleting},
		Target:  []string{},
		Refresh: EnvironmentStatus(conn, name),
		Timeout: EnvironmentDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*mwaa.Environment); ok {
		return v, err
	}

	return nil, err
}
