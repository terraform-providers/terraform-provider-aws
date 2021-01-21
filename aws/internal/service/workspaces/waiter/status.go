package waiter

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/workspaces"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func DirectoryState(conn *workspaces.WorkSpaces, directoryID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := conn.DescribeWorkspaceDirectories(&workspaces.DescribeWorkspaceDirectoriesInput{
			DirectoryIds: aws.StringSlice([]string{directoryID}),
		})
		if err != nil {
			return nil, workspaces.WorkspaceDirectoryStateError, err
		}

		if len(output.Directories) == 0 {
			return output, workspaces.WorkspaceDirectoryStateDeregistered, nil
		}

		directory := output.Directories[0]
		return directory, aws.StringValue(directory.State), nil
	}
}

func WorkspaceState(conn *workspaces.WorkSpaces, workspaceID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := conn.DescribeWorkspaces(&workspaces.DescribeWorkspacesInput{
			WorkspaceIds: aws.StringSlice([]string{workspaceID}),
		})
		if err != nil {
			return nil, workspaces.WorkspaceStateError, err
		}

		if len(output.Workspaces) == 0 {
			return output, workspaces.WorkspaceStateTerminated, nil
		}

		workspace := output.Workspaces[0]

		// https://docs.aws.amazon.com/workspaces/latest/api/API_TerminateWorkspaces.html
		// State TERMINATED is overridden with TERMINATING to catch up directory metadata clean up.
		if aws.StringValue(workspace.State) == workspaces.WorkspaceStateTerminated {
			return workspace, workspaces.WorkspaceStateTerminating, nil
		}

		return workspace, aws.StringValue(workspace.State), nil
	}
}
