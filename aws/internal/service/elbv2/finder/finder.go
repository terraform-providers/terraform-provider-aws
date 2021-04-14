package finder

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elbv2"
)

func TargetGroupByARN(conn *elbv2.ELBV2, arn string) (*elbv2.TargetGroup, error) {
	input := &elbv2.DescribeTargetGroupsInput{
		TargetGroupArns: aws.StringSlice([]string{arn}),
	}

	output, err := conn.DescribeTargetGroups(input)

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, nil
	}

	for _, targetGroup := range output.TargetGroups {
		if targetGroup == nil {
			continue
		}

		if aws.StringValue(targetGroup.TargetGroupArn) != arn {
			continue
		}

		return targetGroup, nil
	}

	return nil, nil
}
