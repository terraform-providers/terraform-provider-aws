package finder

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/networkfirewall"
)

// LoggingConfiguration returns the LoggingConfigurationOutput from a call to DescribeLoggingConfigurationWithContext
// given the context and Firewall ARN.
// Returns nil if the LoggingConfiguration is not found.
func LoggingConfiguration(ctx context.Context, conn *networkfirewall.NetworkFirewall, arn string) (*networkfirewall.DescribeLoggingConfigurationOutput, error) {
	input := &networkfirewall.DescribeLoggingConfigurationInput{
		FirewallArn: aws.String(arn),
	}
	output, err := conn.DescribeLoggingConfigurationWithContext(ctx, input)
	if err != nil {
		return nil, err
	}
	return output, nil
}

// Firewall returns the FirewallOutput from a call to DescribeFirewallWithContext
// given the context and Firewall ARN.
// Returns nil if the Firewall is not found.
func Firewall(ctx context.Context, conn *networkfirewall.NetworkFirewall, arn string) (*networkfirewall.DescribeFirewallOutput, error) {
	input := &networkfirewall.DescribeFirewallInput{
		FirewallArn: aws.String(arn),
	}
	output, err := conn.DescribeFirewallWithContext(ctx, input)
	if err != nil {
		return nil, err
	}
	return output, nil
}

// FirewallPolicy returns the FirewallPolicyOutput from a call to DescribeFirewallPolicyWithContext
// given the context and FirewallPolicy ARN.
// Returns nil if the FirewallPolicy is not found.
func FirewallPolicy(ctx context.Context, conn *networkfirewall.NetworkFirewall, arn string) (*networkfirewall.DescribeFirewallPolicyOutput, error) {
	input := &networkfirewall.DescribeFirewallPolicyInput{
		FirewallPolicyArn: aws.String(arn),
	}
	output, err := conn.DescribeFirewallPolicyWithContext(ctx, input)
	if err != nil {
		return nil, err
	}
	return output, nil
}

// ResourcePolicy returns the Policy string from a call to DescribeResourcePolicyWithContext
// given the context and resource ARN.
// Returns nil if the ResourcePolicy is not found.
func ResourcePolicy(ctx context.Context, conn *networkfirewall.NetworkFirewall, arn string) (*string, error) {
	input := &networkfirewall.DescribeResourcePolicyInput{
		ResourceArn: aws.String(arn),
	}
	output, err := conn.DescribeResourcePolicyWithContext(ctx, input)
	if err != nil {
		return nil, err
	}
	if output == nil {
		return nil, nil
	}
	return output.Policy, nil
}

// RuleGroup returns the RuleGroupOutput from a call to DescribeRuleGroupWithContext
// given the context and RuleGroup ARN.
// Returns nil if the RuleGroup is not found.
func RuleGroup(ctx context.Context, conn *networkfirewall.NetworkFirewall, arn string) (*networkfirewall.DescribeRuleGroupOutput, error) {
	input := &networkfirewall.DescribeRuleGroupInput{
		RuleGroupArn: aws.String(arn),
	}
	output, err := conn.DescribeRuleGroupWithContext(ctx, input)
	if err != nil {
		return nil, err
	}
	return output, nil
}
