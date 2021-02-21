package waiter

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/prometheusservice"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	ResourceStatusFailed  = "Failed"
	ResourceStatusUnknown = "Unknown"
	ResourceStatusDeleted = "Deleted"
)

// WorkspaceCreatedStatus fetches the Workspace and its Status.
func WorkspaceCreatedStatus(ctx context.Context, conn *prometheusservice.PrometheusService, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &prometheusservice.DescribeWorkspaceInput{
			WorkspaceId: aws.String(id),
		}

		output, err := conn.DescribeWorkspaceWithContext(ctx, input)

		if err != nil {
			return output, ResourceStatusFailed, err
		}

		if output == nil || output.Workspace == nil {
			return output, ResourceStatusUnknown, nil
		}

		return output.Workspace, aws.StringValue(output.Workspace.Status.StatusCode), nil
	}
}

// WorkspaceDeletedStatus fetches the Workspace and its Status
func WorkspaceDeletedStatus(ctx context.Context, conn *prometheusservice.PrometheusService, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &prometheusservice.DescribeWorkspaceInput{
			WorkspaceId: aws.String(id),
		}

		output, err := conn.DescribeWorkspaceWithContext(ctx, input)

		if tfawserr.ErrCodeEquals(err, prometheusservice.ErrCodeResourceNotFoundException) {
			return output, ResourceStatusDeleted, nil
		}

		if err != nil {
			return output, ResourceStatusUnknown, err
		}

		if output == nil || output.Workspace == nil {
			return output, ResourceStatusUnknown, nil
		}

		return output.Workspace, aws.StringValue(output.Workspace.Status.StatusCode), nil
	}
}
