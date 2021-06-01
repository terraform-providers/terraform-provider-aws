package waiter

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/datasync"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// TaskStatusAvailable waits for a Task to return Available
func TaskStatusAvailable(conn *datasync.DataSync, arn string, timeout time.Duration) (*datasync.DescribeTaskOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			datasync.TaskStatusCreating,
			datasync.TaskStatusUnavailable,
		},
		Target: []string{
			datasync.TaskStatusAvailable,
			datasync.TaskStatusRunning,
		},
		Refresh: TaskStatus(conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*datasync.DescribeTaskOutput); ok {
		if err != nil && output != nil && output.ErrorCode != nil && output.ErrorDetail != nil {
			newErr := fmt.Errorf("%s: %s", aws.StringValue(output.ErrorCode), aws.StringValue(output.ErrorDetail))

			switch e := err.(type) {
			case *resource.TimeoutError:
				if e.LastError == nil {
					e.LastError = newErr
				}
			case *resource.UnexpectedStateError:
				if e.LastError == nil {
					e.LastError = newErr
				}
			}
		}

		return output, err
	}

	return nil, err
}
