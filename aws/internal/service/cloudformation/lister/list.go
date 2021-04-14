package lister

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
)

func ListStackEventsForOperation(conn *cloudformation.CloudFormation, stackID, requestToken string, fn func(*cloudformation.StackEvent)) error {
	tokenSeen := false
	err := conn.DescribeStackEventsPages(&cloudformation.DescribeStackEventsInput{
		StackName: aws.String(stackID),
	}, func(page *cloudformation.DescribeStackEventsOutput, lastPage bool) bool {
		for _, e := range page.StackEvents {
			currentToken := aws.StringValue(e.ClientRequestToken)
			if !tokenSeen {
				if currentToken != requestToken {
					continue
				}
				tokenSeen = true
			} else {
				if currentToken != requestToken {
					return false
				}
			}

			fn(e)
		}
		return !lastPage
	})
	return err
}
