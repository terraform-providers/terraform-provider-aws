package waiter

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/datasync"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	TaskStatusUnknown = "Unknown"
)

// TaskStatus fetches the Operation and its Status
func TaskStatus(conn *datasync.DataSync, arn string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &datasync.DescribeTaskInput{
			TaskArn: aws.String(arn),
		}

		output, err := conn.DescribeTask(input)

		if tfawserr.ErrMessageContains(err, datasync.ErrCodeInvalidRequestException, "not found") {
			return nil, "", nil
		}

		if err != nil {
			return output, TaskStatusUnknown, err
		}

		if output == nil {
			return output, TaskStatusUnknown, nil
		}

		return output, aws.StringValue(output.Status), nil
	}
}
