package waiter

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/workspaces"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
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
			return nil, workspaces.WorkspaceStateTerminated, nil
		}

		workspace := output.Workspaces[0]
		return workspace, aws.StringValue(workspace.State), nil
	}
}
