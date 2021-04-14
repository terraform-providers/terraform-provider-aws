package waiter

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/networkfirewall"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	ResourceStatusFailed  = "Failed"
	ResourceStatusUnknown = "Unknown"
	ResourceStatusDeleted = "Deleted"
)

// FirewallCreatedStatus fetches the Firewall and its Status.
// A Firewall is READY only when the ConfigurationSyncStateSummary value
// is IN_SYNC and the Attachment Status values for ALL of the configured
// subnets are READY.
func FirewallCreatedStatus(ctx context.Context, conn *networkfirewall.NetworkFirewall, arn string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &networkfirewall.DescribeFirewallInput{
			FirewallArn: aws.String(arn),
		}

		output, err := conn.DescribeFirewallWithContext(ctx, input)

		if err != nil {
			return output, ResourceStatusFailed, err
		}

		if output == nil || output.FirewallStatus == nil {
			return output, ResourceStatusUnknown, nil
		}

		return output.Firewall, aws.StringValue(output.FirewallStatus.Status), nil
	}
}

// FirewallUpdatedStatus fetches the Firewall and its Status and UpdateToken.
func FirewallUpdatedStatus(ctx context.Context, conn *networkfirewall.NetworkFirewall, arn string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &networkfirewall.DescribeFirewallInput{
			FirewallArn: aws.String(arn),
		}

		output, err := conn.DescribeFirewallWithContext(ctx, input)

		if err != nil {
			return output, ResourceStatusFailed, err
		}

		if output == nil || output.FirewallStatus == nil {
			return output, ResourceStatusUnknown, nil
		}

		return output.UpdateToken, aws.StringValue(output.FirewallStatus.Status), nil
	}
}

// FirewallDeletedStatus fetches the Firewall and its Status
func FirewallDeletedStatus(ctx context.Context, conn *networkfirewall.NetworkFirewall, arn string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &networkfirewall.DescribeFirewallInput{
			FirewallArn: aws.String(arn),
		}

		output, err := conn.DescribeFirewallWithContext(ctx, input)

		if tfawserr.ErrCodeEquals(err, networkfirewall.ErrCodeResourceNotFoundException) {
			return output, ResourceStatusDeleted, nil
		}

		if err != nil {
			return output, ResourceStatusUnknown, err
		}

		if output == nil || output.FirewallStatus == nil {
			return output, ResourceStatusUnknown, nil
		}

		return output.Firewall, aws.StringValue(output.FirewallStatus.Status), nil
	}
}

// FirewallPolicyStatus fetches the Firewall Policy and its Status
func FirewallPolicyStatus(ctx context.Context, conn *networkfirewall.NetworkFirewall, arn string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &networkfirewall.DescribeFirewallPolicyInput{
			FirewallPolicyArn: aws.String(arn),
		}

		output, err := conn.DescribeFirewallPolicyWithContext(ctx, input)

		if tfawserr.ErrCodeEquals(err, networkfirewall.ErrCodeResourceNotFoundException) {
			return output, ResourceStatusDeleted, nil
		}

		if err != nil {
			return nil, ResourceStatusUnknown, err
		}

		if output == nil || output.FirewallPolicyResponse == nil {
			return nil, ResourceStatusUnknown, nil
		}

		return output.FirewallPolicy, aws.StringValue(output.FirewallPolicyResponse.FirewallPolicyStatus), nil
	}
}

// RuleGroupStatus fetches the Rule Group and its Status
func RuleGroupStatus(ctx context.Context, conn *networkfirewall.NetworkFirewall, arn string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &networkfirewall.DescribeRuleGroupInput{
			RuleGroupArn: aws.String(arn),
		}

		output, err := conn.DescribeRuleGroupWithContext(ctx, input)

		if tfawserr.ErrCodeEquals(err, networkfirewall.ErrCodeResourceNotFoundException) {
			return output, ResourceStatusDeleted, nil
		}

		if err != nil {
			return nil, ResourceStatusUnknown, err
		}

		if output == nil || output.RuleGroupResponse == nil {
			return nil, ResourceStatusUnknown, nil
		}

		return output.RuleGroup, aws.StringValue(output.RuleGroupResponse.RuleGroupStatus), nil
	}
}
