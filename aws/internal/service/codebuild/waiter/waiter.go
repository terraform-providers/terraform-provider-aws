package waiter

import (
	"time"

	"github.com/aws/aws-sdk-go/service/codebuild"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	// Maximum amount of time to wait for an Operation to return Deleted
	ReportGroupDeleteTimeout = 2 * time.Minute
)

// ReportGroupDeleted waits for an ReportGroup to return Deleted
func ReportGroupDeleted(conn *codebuild.CodeBuild, arn string) (*codebuild.ReportGroup, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{codebuild.ReportGroupStatusTypeDeleting},
		Target:  []string{},
		Refresh: ReportGroupStatus(conn, arn),
		Timeout: ReportGroupDeleteTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*codebuild.ReportGroup); ok {
		return output, err
	}

	return nil, err
}
