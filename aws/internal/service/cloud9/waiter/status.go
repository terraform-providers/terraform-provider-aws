package waiter

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloud9"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// EnvironmentStatus fetches the Environment and its LifecycleStatus
func EnvironmentStatus(conn *cloud9.Cloud9, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := conn.DescribeEnvironmentStatus(&cloud9.DescribeEnvironmentStatusInput{
			EnvironmentId: aws.String(id),
		})
		if err != nil {
			return nil, "", err
		}

		status := aws.StringValue(out.Status)

		if status == cloud9.EnvironmentStatusError && out.Message != nil {
			return out, status, fmt.Errorf("Reason: %s", aws.StringValue(out.Message))
		}

		return out, status, nil
	}
}
