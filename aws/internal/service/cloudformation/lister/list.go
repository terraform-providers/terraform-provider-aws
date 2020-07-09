//go:generate go run ../../../generators/listpages/main.go -function=ListStackSets,ListStackInstances github.com/aws/aws-sdk-go/service/cloudformation

package lister

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
)

func ListAllStackSetsPages(conn *cloudformation.CloudFormation, fn func(*cloudformation.ListStackSetsOutput, bool) bool) error {
	input := &cloudformation.ListStackSetsInput{
		Status: aws.String(cloudformation.StackSetStatusActive),
	}

	return ListStackSetsPages(conn, input, fn)
}

func ListAllStackSetInstancesPages(conn *cloudformation.CloudFormation, stackSetName string, fn func(*cloudformation.ListStackInstancesOutput, bool) bool) error {
	input := &cloudformation.ListStackInstancesInput{
		StackSetName: aws.String(stackSetName),
	}

	return ListStackInstancesPages(conn, input, fn)
}
