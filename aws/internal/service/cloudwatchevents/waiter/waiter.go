package waiter

import (
	"time"

	events "github.com/aws/aws-sdk-go/service/cloudwatchevents"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	ConnectionCreatedTimeout = 2 * time.Minute
	ConnectionDeletedTimeout = 2 * time.Minute
	ConnectionUpdatedTimeout = 2 * time.Minute
)

func ConnectionCreated(conn *events.CloudWatchEvents, id string) (*events.DescribeConnectionOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{events.ConnectionStateCreating, events.ConnectionStateAuthorizing},
		Target:  []string{events.ConnectionStateAuthorized, events.ConnectionStateDeauthorized},
		Refresh: ConnectionState(conn, id),
		Timeout: ConnectionCreatedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*events.DescribeConnectionOutput); ok {
		return v, err
	}

	return nil, err
}

func ConnectionDeleted(conn *events.CloudWatchEvents, id string) (*events.DescribeConnectionOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{events.ConnectionStateDeleting},
		Target:  []string{},
		Refresh: ConnectionState(conn, id),
		Timeout: ConnectionDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*events.DescribeConnectionOutput); ok {
		return v, err
	}

	return nil, err
}

func ConnectionUpdated(conn *events.CloudWatchEvents, id string) (*events.DescribeConnectionOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{events.ConnectionStateUpdating, events.ConnectionStateAuthorizing, events.ConnectionStateDeauthorizing},
		Target:  []string{events.ConnectionStateAuthorized, events.ConnectionStateDeauthorized},
		Refresh: ConnectionState(conn, id),
		Timeout: ConnectionUpdatedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*events.DescribeConnectionOutput); ok {
		return v, err
	}

	return nil, err
}
