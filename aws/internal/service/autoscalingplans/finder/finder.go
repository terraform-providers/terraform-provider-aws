package finder

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscalingplans"
)

// ScalingPlan returns the scaling plan corresponding to the specified name and version.
// Returns nil and potentially an API error if no scaling plan is found.
func ScalingPlan(conn *autoscalingplans.AutoScalingPlans, scalingPlanName string, scalingPlanVersion int) (*autoscalingplans.ScalingPlan, error) {
	input := &autoscalingplans.DescribeScalingPlansInput{
		ScalingPlanNames:   aws.StringSlice([]string{scalingPlanName}),
		ScalingPlanVersion: aws.Int64(int64(scalingPlanVersion)),
	}

	output, err := conn.DescribeScalingPlans(input)
	if err != nil {
		return nil, err
	}

	if output == nil || len(output.ScalingPlans) == 0 {
		return nil, nil
	}

	return output.ScalingPlans[0], nil
}
