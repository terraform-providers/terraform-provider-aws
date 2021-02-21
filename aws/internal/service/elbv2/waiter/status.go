package waiter

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func LoadBalancerState(conn *elbv2.ELBV2, arn string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &elbv2.DescribeLoadBalancersInput{
			LoadBalancerArns: []*string{aws.String(arn)},
		}

		output, err := conn.DescribeLoadBalancers(input)

		if tfawserr.ErrCodeEquals(err, elbv2.ErrCodeLoadBalancerNotFoundException) {
			return nil, "", nil
		}
		if err != nil {
			return nil, "", err
		}

		if len(output.LoadBalancers) != 1 {
			return nil, "", fmt.Errorf("No load balancers found for %s", arn)
		}
		lb := output.LoadBalancers[0]

		return output, aws.StringValue(lb.State.Code), nil
	}
}
