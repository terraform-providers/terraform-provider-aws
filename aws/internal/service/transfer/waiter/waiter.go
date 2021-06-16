package waiter

import (
	"time"

	"github.com/aws/aws-sdk-go/service/transfer"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	ServerDeletedTimeout = 10 * time.Minute
	UserDeletedTimeout   = 10 * time.Minute
)

func ServerCreated(conn *transfer.Transfer, id string, timeout time.Duration) (*transfer.DescribedServer, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{transfer.StateStarting},
		Target:  []string{transfer.StateOnline},
		Refresh: ServerState(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*transfer.DescribedServer); ok {
		return output, err
	}

	return nil, err
}

func ServerDeleted(conn *transfer.Transfer, id string) (*transfer.DescribedServer, error) {
	stateConf := &resource.StateChangeConf{
		Pending: transfer.State_Values(),
		Target:  []string{},
		Refresh: ServerState(conn, id),
		Timeout: ServerDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*transfer.DescribedServer); ok {
		return output, err
	}

	return nil, err
}

func ServerStarted(conn *transfer.Transfer, id string, timeout time.Duration) (*transfer.DescribedServer, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{transfer.StateStarting, transfer.StateOffline, transfer.StateStopping},
		Target:  []string{transfer.StateOnline},
		Refresh: ServerState(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*transfer.DescribedServer); ok {
		return output, err
	}

	return nil, err
}

func ServerStopped(conn *transfer.Transfer, id string, timeout time.Duration) (*transfer.DescribedServer, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{transfer.StateStarting, transfer.StateOnline, transfer.StateStopping},
		Target:  []string{transfer.StateOffline},
		Refresh: ServerState(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*transfer.DescribedServer); ok {
		return output, err
	}

	return nil, err
}

func UserDeleted(conn *transfer.Transfer, serverId, userName string) (*transfer.DescribedUser, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{"Available"},
		Target:  []string{},
		Refresh: UserState(conn, serverId, userName),
		Timeout: UserDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*transfer.DescribedUser); ok {
		return output, err
	}

	return nil, err
}
