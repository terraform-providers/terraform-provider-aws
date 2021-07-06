package waiter

import (
	"time"

	"github.com/aws/aws-sdk-go/service/efs"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	// Maximum amount of time to wait for an Operation to return Success
	AccessPointCreatedTimeout       = 10 * time.Minute
	AccessPointDeletedTimeout       = 10 * time.Minute
	FileSystemAvailableTimeout      = 10 * time.Minute
	FileSystemAvailableDelayTimeout = 2 * time.Second
	FileSystemAvailableMinTimeout   = 3 * time.Second
	FileSystemDeletedTimeout        = 10 * time.Minute
	FileSystemDeletedDelayTimeout   = 2 * time.Second
	FileSystemDeletedMinTimeout     = 3 * time.Second

	BackupPolicyDisabledTimeout = 10 * time.Minute
	BackupPolicyEnabledTimeout  = 10 * time.Minute
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

// AccessPointDeleted waits for an Access Point to return Deleted
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

// FileSystemAvailable waits for an Operation to return Available
func FileSystemAvailable(conn *efs.EFS, fileSystemID string) (*efs.FileSystemDescription, error) {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{efs.LifeCycleStateCreating, efs.LifeCycleStateUpdating},
		Target:     []string{efs.LifeCycleStateAvailable},
		Refresh:    FileSystemLifeCycleState(conn, fileSystemID),
		Timeout:    FileSystemAvailableTimeout,
		Delay:      FileSystemAvailableDelayTimeout,
		MinTimeout: FileSystemAvailableMinTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*efs.FileSystemDescription); ok {
		return output, err
	}

	return nil, err
}

// FileSystemDeleted waits for an Operation to return Deleted
func FileSystemDeleted(conn *efs.EFS, fileSystemID string) (*efs.FileSystemDescription, error) {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{efs.LifeCycleStateAvailable, efs.LifeCycleStateDeleting},
		Target:     []string{},
		Refresh:    FileSystemLifeCycleState(conn, fileSystemID),
		Timeout:    FileSystemDeletedTimeout,
		Delay:      FileSystemDeletedDelayTimeout,
		MinTimeout: FileSystemDeletedMinTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*efs.FileSystemDescription); ok {
		return output, err
	}

	return nil, err
}

func BackupPolicyDisabled(conn *efs.EFS, id string) (*efs.BackupPolicy, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{efs.StatusDisabling},
		Target:  []string{efs.StatusDisabled},
		Refresh: BackupPolicyStatus(conn, id),
		Timeout: BackupPolicyDisabledTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*efs.BackupPolicy); ok {
		return output, err
	}

	return nil, err
}

func BackupPolicyEnabled(conn *efs.EFS, id string) (*efs.BackupPolicy, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{efs.StatusEnabling},
		Target:  []string{efs.StatusEnabled},
		Refresh: BackupPolicyStatus(conn, id),
		Timeout: BackupPolicyEnabledTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*efs.BackupPolicy); ok {
		return output, err
	}

	return nil, err
}
