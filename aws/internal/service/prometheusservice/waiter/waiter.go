package waiter

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/service/prometheusservice"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	// Maximum amount of time to wait for a Workspace to be created, updated, or deleted
	WorkspaceTimeout = 5 * time.Minute
)

// WorkspaceCreated waits for a Workspace to return "Active"
func WorkspaceCreated(ctx context.Context, conn *prometheusservice.PrometheusService, id string) (*prometheusservice.WorkspaceSummary, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{prometheusservice.WorkspaceStatusCodeCreating},
		Target:  []string{prometheusservice.WorkspaceStatusCodeActive},
		Refresh: WorkspaceCreatedStatus(ctx, conn, id),
		Timeout: WorkspaceTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*prometheusservice.WorkspaceSummary); ok {
		return v, err
	}

	return nil, err
}

// WorkspaceDeleted waits for a Workspace to return "Deleted"
func WorkspaceDeleted(ctx context.Context, conn *prometheusservice.PrometheusService, arn string) (*prometheusservice.WorkspaceSummary, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{prometheusservice.WorkspaceStatusCodeDeleting},
		Target:  []string{ResourceStatusDeleted},
		Refresh: WorkspaceDeletedStatus(ctx, conn, arn),
		Timeout: WorkspaceTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*prometheusservice.WorkspaceSummary); ok {
		return v, err
	}

	return nil, err
}
