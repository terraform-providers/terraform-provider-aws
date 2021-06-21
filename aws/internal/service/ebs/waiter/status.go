package waiter

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	tfec2 "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/ec2"
)

func EbsSnapshotImportStatus(conn *ec2.EC2, importTaskId string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		params := &ec2.DescribeImportSnapshotTasksInput{
			ImportTaskIds: []*string{aws.String(importTaskId)},
		}

		resp, err := conn.DescribeImportSnapshotTasks(params)
		if err != nil {
			return nil, "", err
		}

		if task := resp.ImportSnapshotTasks[0]; task != nil {
			detail := task.SnapshotTaskDetail
			if aws.StringValue(detail.Status) != "" && *detail.Status == tfec2.EbsSnapshotImportDeleting {
				if aws.StringValue(detail.StatusMessage) != "" {
					err = fmt.Errorf("Snapshot import task is deleting: %s", aws.StringValue(detail.StatusMessage))
				} else {
					err = fmt.Errorf("Snapshot import task is deleting: (no status message provided)")
				}

			}

			return detail, aws.StringValue(detail.Status), err
		} else {
			return nil, "", fmt.Errorf("AWS doesn't know about our import task ID (%s)", importTaskId)
		}

	}
}
