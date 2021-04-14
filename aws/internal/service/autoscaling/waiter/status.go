package waiter

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func InstanceRefreshStatus(conn *autoscaling.AutoScaling, asgName, instanceRefreshId string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := autoscaling.DescribeInstanceRefreshesInput{
			AutoScalingGroupName: aws.String(asgName),
			InstanceRefreshIds:   []*string{aws.String(instanceRefreshId)},
		}
		output, err := conn.DescribeInstanceRefreshes(&input)
		if err != nil {
			return nil, "", err
		}

		if output == nil || len(output.InstanceRefreshes) == 0 || output.InstanceRefreshes[0] == nil {
			return nil, "", nil
		}

		instanceRefresh := output.InstanceRefreshes[0]

		return instanceRefresh, aws.StringValue(instanceRefresh.Status), nil
	}
}
