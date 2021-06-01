package finder

import (
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func Stack(conn *cloudformation.CloudFormation, stackID string) (*cloudformation.Stack, error) {
	input := &cloudformation.DescribeStacksInput{
		StackName: aws.String(stackID),
	}
	log.Printf("[DEBUG] Querying CloudFormation Stack: %s", input)
	resp, err := conn.DescribeStacks(input)
	if tfawserr.ErrCodeEquals(err, "ValidationError") {
		return nil, &resource.NotFoundError{
			LastError:    err,
			LastRequest:  input,
			LastResponse: resp,
		}
	}
	if err != nil {
		return nil, err
	}

	if resp == nil {
		return nil, &resource.NotFoundError{
			LastRequest:  input,
			LastResponse: resp,
			Message:      "returned empty response",
		}

	}
	stacks := resp.Stacks
	if len(stacks) < 1 {
		return nil, &resource.NotFoundError{
			LastRequest:  input,
			LastResponse: resp,
			Message:      "returned no results",
		}
	}

	stack := stacks[0]
	if aws.StringValue(stack.StackStatus) == cloudformation.StackStatusDeleteComplete {
		return nil, &resource.NotFoundError{
			LastRequest:  input,
			LastResponse: resp,
			Message:      "CloudFormation Stack deleted",
		}
	}

	return stack, nil
}
