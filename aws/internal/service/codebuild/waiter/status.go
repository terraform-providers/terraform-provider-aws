package waiter

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codebuild"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/codebuild/finder"
)

const (
	ReportGroupStatusUnknown  = "Unknown"
	ReportGroupStatusNotFound = "NotFound"
)

// ReportGroupStatus fetches the Report Group and its Status
func ReportGroupStatus(conn *codebuild.CodeBuild, arn string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := finder.ReportGroupByArn(conn, arn)
		if err != nil {
			return nil, ReportGroupStatusUnknown, err
		}

		if output == nil {
			return nil, ReportGroupStatusNotFound, nil
		}

		return output, aws.StringValue(output.Status), nil
	}
}
