package finder

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sfn"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func StateMachineByARN(conn *sfn.SFN, arn string) (*sfn.DescribeStateMachineOutput, error) {
	input := &sfn.DescribeStateMachineInput{
		StateMachineArn: aws.String(arn),
	}

	output, err := conn.DescribeStateMachine(input)

	if tfawserr.ErrCodeEquals(err, sfn.ErrCodeStateMachineDoesNotExist) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, &resource.NotFoundError{
			Message:     "Empty result",
			LastRequest: input,
		}
	}

	return output, nil
}
