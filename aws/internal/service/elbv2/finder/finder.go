package finder

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elbv2"
)

func ListenerByARN(conn *elbv2.ELBV2, arn string) (*elbv2.Listener, error) {
	input := &elbv2.DescribeListenersInput{
		ListenerArns: aws.StringSlice([]string{arn}),
	}

	var result *elbv2.Listener

	err := conn.DescribeListenersPages(input, func(page *elbv2.DescribeListenersOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, l := range page.Listeners {
			if l == nil {
				continue
			}

			if aws.StringValue(l.ListenerArn) == arn {
				result = l
				return false
			}
		}

		return !lastPage
	})

	return result, err
}

func LoadBalancerByARN(conn *elbv2.ELBV2, arn string) (*elbv2.LoadBalancer, error) {
	input := &elbv2.DescribeLoadBalancersInput{
		LoadBalancerArns: aws.StringSlice([]string{arn}),
	}

	var result *elbv2.LoadBalancer

	err := conn.DescribeLoadBalancersPages(input, func(page *elbv2.DescribeLoadBalancersOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, lb := range page.LoadBalancers {
			if lb == nil {
				continue
			}

			if aws.StringValue(lb.LoadBalancerArn) == arn {
				result = lb
				return false
			}
		}

		return !lastPage
	})

	return result, err
}

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
