package waiter

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sfn"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// StateMachineStatus fetches the Operation and its Status
func StateMachineStatus(conn *sfn.SFN, stateMachineArn string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &sfn.DescribeStateMachineInput{
			StateMachineArn: aws.String(stateMachineArn),
		}

		output, err := conn.DescribeStateMachine(input)

		if err != nil {
			return nil, "", err
		}

		if output == nil {
			return nil, "", nil
		}

		return output, aws.StringValue(output.Status), nil
	}
}
