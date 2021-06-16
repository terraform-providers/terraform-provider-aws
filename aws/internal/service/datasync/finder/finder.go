package finder

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/datasync"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func AgentByARN(conn *datasync.DataSync, arn string) (*datasync.DescribeAgentOutput, error) {
	input := &datasync.DescribeAgentInput{
		AgentArn: aws.String(arn),
	}

	output, err := conn.DescribeAgent(input)

	if tfawserr.ErrMessageContains(err, datasync.ErrCodeInvalidRequestException, "does not exist") {
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

func TaskByARN(conn *datasync.DataSync, arn string) (*datasync.DescribeTaskOutput, error) {
	input := &datasync.DescribeTaskInput{
		TaskArn: aws.String(arn),
	}

	output, err := conn.DescribeTask(input)

	if tfawserr.ErrMessageContains(err, datasync.ErrCodeInvalidRequestException, "not found") {
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
