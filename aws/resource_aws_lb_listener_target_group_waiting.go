package aws

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/hashicorp/errwrap"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
)

// waitForTargetGroupCapacity gathers the current numbers of healthy instances
// in the target group and yields these numbers to a
// capacitySatifiedFunction. Loops for up to wait_for_capacity_timeout until
// the capacitySatisfiedFunc returns true.
//
// See "Waiting for Capacity" in docs for more discussion of the feature.
func waitForListenerTargetGroupCapacity(
	d *schema.ResourceData, // this is the resource data corresponding to the listener
	meta interface{},
	satisfiedFunc capacitySatisfiedFunc) error {

	wait, err := time.ParseDuration(d.Get("wait_for_capacity_timeout").(string))
	if err != nil {
		return err
	}

	defaultActionMap := d.Get("default_action").([]interface{})[0].(map[string]interface{})
	targetGroupArn := defaultActionMap["target_group_arn"].(string)

	targetGroupCapacity := -1

	if wait == 0 {
		log.Printf("[DEBUG] Capacity timeout set to 0, skipping capacity waiting.")
		return nil
	}

	log.Printf("[DEBUG] Waiting on %s for target group capacity...", d.Id())

	err = resource.Retry(wait, func() *resource.RetryError {

		states, err := getInstanceStatesForTargetGroup(&targetGroupArn, meta)
		if err != nil {
			return resource.NonRetryableError(err)
		}

		if len(states) == 0 {
			return resource.RetryableError(
				fmt.Errorf("%q: Waiting up to %s: states not yet available for target group %s",
					d.Id(), wait, targetGroupArn))
		}

		healthyCount := 0
		for _, state := range states {
			if strings.EqualFold(state, "healthy") {
				healthyCount++
			}
		}

		if targetGroupCapacity == -1 {
			instanceIDs := make([]string, 0, len(states))
			for instanceID := range states {
				instanceIDs = append(instanceIDs, instanceID)
			}
			targetGroupCapacity, err = getTargetCapacity(d, instanceIDs, meta)
			if err != nil {
				return resource.NonRetryableError(err)
			}
		}

		satisfied, reason := satisfiedFunc(d, healthyCount, targetGroupCapacity)

		log.Printf("[DEBUG] %q Capacity: %d in target group, %d healthy, satisfied: %t, reason: %q",
			d.Id(), targetGroupCapacity, healthyCount, satisfied, reason)

		if satisfied {
			return nil
		}

		return resource.RetryableError(
			fmt.Errorf("%q: Waiting up to %s: %s", d.Id(), wait, reason))
	})

	if err == nil {
		return nil
	}

	recentStatus := ""

	conn := meta.(*AWSClient).autoscalingconn
	resp, aErr := conn.DescribeScalingActivities(&autoscaling.DescribeScalingActivitiesInput{
		AutoScalingGroupName: aws.String(d.Id()),
		MaxRecords:           aws.Int64(1),
	})
	if aErr == nil {
		if len(resp.Activities) > 0 {
			recentStatus = fmt.Sprintf("%s", resp.Activities[0])
		} else {
			recentStatus = "(0 activities found)"
		}
	} else {
		recentStatus = fmt.Sprintf("(Failed to describe scaling activities: %s)", aErr)
	}

	msg := fmt.Sprintf("{{err}}. Most recent activity: %s", recentStatus)
	return errwrap.Wrapf(msg, err)
}

// Uses the 'wait_for_target_group_capacity' value configured on the listener
// resource, defaulting to the sum of the DesiredCapacity values configured
// for the associated autoscaling groups
func getTargetCapacity(d *schema.ResourceData, instanceIDs []string, meta interface{}) (int, error) {
	target := d.Get("min_target_group_capacity").(int)
	if target == -1 {

		ec2conn := meta.(*AWSClient).ec2conn
		autoscalingconn := meta.(*AWSClient).autoscalingconn
		autoscalingGroupNames := make(map[string]bool)
		instancesResp, err := ec2conn.DescribeInstances(&ec2.DescribeInstancesInput{
			InstanceIds: aws.StringSlice(instanceIDs),
		})

		if err != nil {
			return -1, fmt.Errorf("Error describing Instances %v; %v", instanceIDs, err)
		}
		for _, resv := range instancesResp.Reservations {
			for _, instance := range resv.Instances {
				for _, tag := range instance.Tags {
					if aws.StringValue(tag.Key) == "aws:autoscaling:groupName" {
						asgName := aws.StringValue(tag.Value)
						log.Printf("[DEBUG] getTargetCapacity: Found autoscaling group name %s on instance tag: %v",
							asgName, aws.StringValue(instance.InstanceId))
						autoscalingGroupNames[asgName] = true
					}
				}
			}
		}

		asgNames := make([]string, 0, len(autoscalingGroupNames))
		for name := range autoscalingGroupNames {
			asgNames = append(asgNames, name)
		}
		asgResp, err := autoscalingconn.DescribeAutoScalingGroups(&autoscaling.DescribeAutoScalingGroupsInput{
			AutoScalingGroupNames: aws.StringSlice(asgNames),
		})
		if err != nil {
			return -1, fmt.Errorf("Error describing AutoScaling Group: %s", err)
		}
		target = 0
		for _, asg := range asgResp.AutoScalingGroups {
			desired := int(aws.Int64Value(asg.DesiredCapacity))
			log.Printf("[DEBUG] getTargetCapacity: Adding %d capacity from %s", desired, aws.StringValue(asg.AutoScalingGroupName))
			target += desired
		}
		log.Printf("[DEBUG] getTargetCapacity: Resolved target capacity of %d from ASGs: %v", target, asgNames)
	}
	return target, nil
}

// getInstanceStatesForTargetGroup returns a mapping of the instance states of
// all the ALB target groups attached to the provided ASG.
//
// Note that this is the instance state function for Application Load
// Balancing (aka ELBv2).
//
// Nested like: targetGroupArn -> instanceId -> instanceState
func getInstanceStatesForTargetGroup(targetGroupArn *string, meta interface{}) (map[string]string, error) {
	targetInstanceStates := make(map[string]string)

	elbv2conn := meta.(*AWSClient).elbv2conn
	opts := &elbv2.DescribeTargetHealthInput{TargetGroupArn: targetGroupArn}
	r, err := elbv2conn.DescribeTargetHealth(opts)
	if err != nil {
		return nil, err
	}
	for _, desc := range r.TargetHealthDescriptions {
		if desc.Target == nil || desc.Target.Id == nil || desc.TargetHealth == nil || desc.TargetHealth.State == nil {
			continue
		}
		targetInstanceStates[*desc.Target.Id] = *desc.TargetHealth.State
	}

	return targetInstanceStates, nil
}
