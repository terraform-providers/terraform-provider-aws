package waiter

import (
	"time"

	"github.com/aws/aws-sdk-go/service/autoscalingplans"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	// Maximum amount of time to wait for a ScalingPlan to return Created
	ScalingPlanCreatedTimeout = 5 * time.Minute

	// Maximum amount of time to wait for a ScalingPlan to return Deleted
	ScalingPlanDeletedTimeout = 5 * time.Minute

	// Maximum amount of time to wait for a ScalingPlan to return Updated
	ScalingPlanUpdatedTimeout = 5 * time.Minute
)

// ScalingPlanCreated waits for a ScalingPlan to return Created
func ScalingPlanCreated(conn *autoscalingplans.AutoScalingPlans, scalingPlanName string, scalingPlanVersion int) (*autoscalingplans.ScalingPlan, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{autoscalingplans.ScalingPlanStatusCodeCreationInProgress},
		Target:  []string{autoscalingplans.ScalingPlanStatusCodeActive, autoscalingplans.ScalingPlanStatusCodeActiveWithProblems},
		Refresh: ScalingPlanStatus(conn, scalingPlanName, scalingPlanVersion),
		Timeout: ScalingPlanCreatedTimeout,
		Delay:   10 * time.Second,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*autoscalingplans.ScalingPlan); ok {
		return v, err
	}

	return nil, err
}

// ScalingPlanDeleted waits for a ScalingPlan to return Deleted
func ScalingPlanDeleted(conn *autoscalingplans.AutoScalingPlans, scalingPlanName string, scalingPlanVersion int) (*autoscalingplans.ScalingPlan, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{autoscalingplans.ScalingPlanStatusCodeDeletionInProgress},
		Target:  []string{},
		Refresh: ScalingPlanStatus(conn, scalingPlanName, scalingPlanVersion),
		Timeout: ScalingPlanDeletedTimeout,
		Delay:   10 * time.Second,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*autoscalingplans.ScalingPlan); ok {
		return v, err
	}

	return nil, err
}

// ScalingPlanUpdated waits for a ScalingPlan to return Updated
func ScalingPlanUpdated(conn *autoscalingplans.AutoScalingPlans, scalingPlanName string, scalingPlanVersion int) (*autoscalingplans.ScalingPlan, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{autoscalingplans.ScalingPlanStatusCodeUpdateInProgress},
		Target:  []string{autoscalingplans.ScalingPlanStatusCodeActive, autoscalingplans.ScalingPlanStatusCodeActiveWithProblems},
		Refresh: ScalingPlanStatus(conn, scalingPlanName, scalingPlanVersion),
		Timeout: ScalingPlanUpdatedTimeout,
		Delay:   10 * time.Second,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*autoscalingplans.ScalingPlan); ok {
		return v, err
	}

	return nil, err
}
