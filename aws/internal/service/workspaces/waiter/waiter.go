package waiter

import (
	"time"

	"github.com/aws/aws-sdk-go/service/workspaces"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

const (
	// Maximum amount of time to wait for a Directory to return Registered
	DirectoryRegisteredTimeout = 10 * time.Minute

	// Maximum amount of time to wait for a Directory to return Deregistered
	DirectoryDeregisteredTimeout = 10 * time.Minute

	// Maximum amount of time to wait for a WorkSpace to return Available
	WorkspaceAvailableTimeout = 30 * time.Minute

	// Maximum amount of time to wait for a WorkSpace while returning Updating
	WorkspaceUpdatingTimeout = 10 * time.Minute

	// Amount of time to delay before checking WorkSpace when updating
	WorkspaceUpdatingDelay = 1 * time.Minute

	// Maximum amount of time to wait for a WorkSpace to return Terminated
	WorkspaceTerminatedTimeout = 10 * time.Minute
)

func DirectoryRegistered(conn *workspaces.WorkSpaces, directoryID string) (*workspaces.WorkspaceDirectory, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			workspaces.WorkspaceDirectoryStateRegistering,
		},
		Target:  []string{workspaces.WorkspaceDirectoryStateRegistered},
		Refresh: DirectoryState(conn, directoryID),
		Timeout: DirectoryRegisteredTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*workspaces.WorkspaceDirectory); ok {
		return v, err
	}

	return nil, err
}

func DirectoryDeregistered(conn *workspaces.WorkSpaces, directoryID string) (*workspaces.WorkspaceDirectory, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			workspaces.WorkspaceDirectoryStateRegistering,
			workspaces.WorkspaceDirectoryStateRegistered,
			workspaces.WorkspaceDirectoryStateDeregistering,
		},
		Target: []string{
			workspaces.WorkspaceDirectoryStateDeregistered,
		},
		Refresh: DirectoryState(conn, directoryID),
		Timeout: DirectoryDeregisteredTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*workspaces.WorkspaceDirectory); ok {
		return v, err
	}

	return nil, err
}

func WorkspaceAvailable(conn *workspaces.WorkSpaces, workspaceID string) (*workspaces.Workspace, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			workspaces.WorkspaceStatePending,
			workspaces.WorkspaceStateStarting,
		},
		Target:  []string{workspaces.WorkspaceStateAvailable},
		Refresh: WorkspaceState(conn, workspaceID),
		Timeout: WorkspaceAvailableTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*workspaces.Workspace); ok {
		return v, err
	}

	return nil, err
}

func WorkspaceTerminated(conn *workspaces.WorkSpaces, workspaceID string) (*workspaces.Workspace, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			workspaces.WorkspaceStatePending,
			workspaces.WorkspaceStateAvailable,
			workspaces.WorkspaceStateImpaired,
			workspaces.WorkspaceStateUnhealthy,
			workspaces.WorkspaceStateRebooting,
			workspaces.WorkspaceStateStarting,
			workspaces.WorkspaceStateRebuilding,
			workspaces.WorkspaceStateRestoring,
			workspaces.WorkspaceStateMaintenance,
			workspaces.WorkspaceStateAdminMaintenance,
			workspaces.WorkspaceStateSuspended,
			workspaces.WorkspaceStateUpdating,
			workspaces.WorkspaceStateStopping,
			workspaces.WorkspaceStateStopped,
			workspaces.WorkspaceStateTerminating,
			workspaces.WorkspaceStateError,
		},
		Target: []string{
			workspaces.WorkspaceStateTerminated,
		},
		Refresh: WorkspaceState(conn, workspaceID),
		Timeout: WorkspaceTerminatedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*workspaces.Workspace); ok {
		return v, err
	}

	return nil, err
}

func WorkspaceUpdated(conn *workspaces.WorkSpaces, workspaceID string) (*workspaces.Workspace, error) {
	// OperationInProgressException: The properties of this WorkSpace are currently under modification. Please try again in a moment.
	// AWS Workspaces service doesn't change instance status to "Updating" during property modification. Respective AWS Support feature request has been created. Meanwhile, artificial delay is placed here as a workaround.
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			workspaces.WorkspaceStateUpdating,
		},
		Target: []string{
			workspaces.WorkspaceStateAvailable,
			workspaces.WorkspaceStateStopped,
		},
		Refresh: WorkspaceState(conn, workspaceID),
		Delay:   WorkspaceUpdatingDelay,
		Timeout: WorkspaceUpdatingTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*workspaces.Workspace); ok {
		return v, err
	}

	return nil, err
}
