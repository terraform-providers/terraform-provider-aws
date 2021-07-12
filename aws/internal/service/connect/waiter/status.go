package waiter

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/connect"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func InstanceStatus(ctx context.Context, conn *connect.Connect, instanceId string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &connect.DescribeInstanceInput{
			InstanceId: aws.String(instanceId),
		}

		output, err := conn.DescribeInstanceWithContext(ctx, input)

		if err != nil {
			return nil, connect.ErrCodeResourceNotFoundException, err
		}

		state := aws.StringValue(output.Instance.InstanceStatus)

		return output, state, nil
	}
}
