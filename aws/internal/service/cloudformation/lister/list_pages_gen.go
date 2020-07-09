// Code generated by "aws/internal/generators/listpages/main.go -function=ListStackSets,ListStackInstances github.com/aws/aws-sdk-go/service/cloudformation"; DO NOT EDIT.

package lister

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
)

func ListStackInstancesPages(conn *cloudformation.CloudFormation, input *cloudformation.ListStackInstancesInput, fn func(*cloudformation.ListStackInstancesOutput, bool) bool) error {
	for {
		output, err := conn.ListStackInstances(input)
		if err != nil {
			return err
		}

		lastPage := aws.StringValue(output.NextToken) == ""
		if !fn(output, lastPage) || lastPage {
			break
		}

		input.NextToken = output.NextToken
	}
	return nil
}

func ListStackSetsPages(conn *cloudformation.CloudFormation, input *cloudformation.ListStackSetsInput, fn func(*cloudformation.ListStackSetsOutput, bool) bool) error {
	for {
		output, err := conn.ListStackSets(input)
		if err != nil {
			return err
		}

		lastPage := aws.StringValue(output.NextToken) == ""
		if !fn(output, lastPage) || lastPage {
			break
		}

		input.NextToken = output.NextToken
	}
	return nil
}
