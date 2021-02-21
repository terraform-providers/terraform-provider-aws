package waiter

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func ChangeSetStatus(conn *cloudformation.CloudFormation, stackID, changeSetName string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		resp, err := conn.DescribeChangeSet(&cloudformation.DescribeChangeSetInput{
			ChangeSetName: aws.String(changeSetName),
			StackName:     aws.String(stackID),
		})
		if err != nil {
			log.Printf("[ERROR] Failed to describe CloudFormation change set: %s", err)
			return nil, "", err
		}

		if resp == nil {
			log.Printf("[WARN] Describing CloudFormation change set returned no response")
			return nil, "", nil
		}

		status := aws.StringValue(resp.Status)

		return resp, status, err
	}
}

func StackSetOperationStatus(conn *cloudformation.CloudFormation, stackSetName, operationID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &cloudformation.DescribeStackSetOperationInput{
			OperationId:  aws.String(operationID),
			StackSetName: aws.String(stackSetName),
		}

		output, err := conn.DescribeStackSetOperation(input)

		if tfawserr.ErrCodeEquals(err, cloudformation.ErrCodeOperationNotFoundException) {
			return nil, cloudformation.StackSetOperationStatusRunning, nil
		}

		if err != nil {
			return nil, cloudformation.StackSetOperationStatusFailed, err
		}

		if output == nil || output.StackSetOperation == nil {
			return nil, cloudformation.StackSetOperationStatusRunning, nil
		}

		if aws.StringValue(output.StackSetOperation.Status) == cloudformation.StackSetOperationStatusFailed {
			allResults := make([]string, 0)
			listOperationResultsInput := &cloudformation.ListStackSetOperationResultsInput{
				OperationId:  aws.String(operationID),
				StackSetName: aws.String(stackSetName),
			}

			// TODO: PAGES
			for {
				listOperationResultsOutput, err := conn.ListStackSetOperationResults(listOperationResultsInput)

				if err != nil {
					return output.StackSetOperation, cloudformation.StackSetOperationStatusFailed, fmt.Errorf("error listing Operation (%s) errors: %w", operationID, err)
				}

				if listOperationResultsOutput == nil {
					continue
				}

				for _, summary := range listOperationResultsOutput.Summaries {
					allResults = append(allResults, fmt.Sprintf("Account (%s) Region (%s) Status (%s) Status Reason: %s", aws.StringValue(summary.Account), aws.StringValue(summary.Region), aws.StringValue(summary.Status), aws.StringValue(summary.StatusReason)))
				}

				if aws.StringValue(listOperationResultsOutput.NextToken) == "" {
					break
				}

				listOperationResultsInput.NextToken = listOperationResultsOutput.NextToken
			}

			return output.StackSetOperation, cloudformation.StackSetOperationStatusFailed, fmt.Errorf("Operation (%s) Results:\n%s", operationID, strings.Join(allResults, "\n"))
		}

		return output.StackSetOperation, aws.StringValue(output.StackSetOperation.Status), nil
	}
}

const (
	stackStatusError    = "Error"
	stackStatusNotFound = "NotFound"
)

func StackStatus(conn *cloudformation.CloudFormation, stackName string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		resp, err := conn.DescribeStacks(&cloudformation.DescribeStacksInput{
			StackName: aws.String(stackName),
		})
		if err != nil {
			return nil, stackStatusError, err
		}

		if resp.Stacks == nil || len(resp.Stacks) == 0 {
			return nil, stackStatusNotFound, nil
		}

		return resp.Stacks[0], aws.StringValue(resp.Stacks[0].StackStatus), err
	}
}
