package waiter

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/service/networkfirewall"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	// Maximum amount of time to wait for a Firewall to be created, updated, or deleted
	FirewallTimeout = 20 * time.Minute
	// Maximum amount of time to wait for a Firewall Policy to be deleted
	FirewallPolicyTimeout = 10 * time.Minute
	// Maximum amount of time to wait for a Rule Group to be deleted
	RuleGroupDeleteTimeout = 10 * time.Minute
)

func FirewallCreated(ctx context.Context, conn *networkfirewall.NetworkFirewall, arn string) (*networkfirewall.Firewall, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{networkfirewall.FirewallStatusValueProvisioning},
		Target:  []string{networkfirewall.FirewallStatusValueReady},
		Refresh: FirewallCreatedStatus(ctx, conn, arn),
		Timeout: FirewallTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*networkfirewall.Firewall); ok {
		return v, err
	}

	return nil, err
}

func FirewallUpdated(ctx context.Context, conn *networkfirewall.NetworkFirewall, arn string) (*string, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{networkfirewall.FirewallStatusValueProvisioning},
		Target:  []string{networkfirewall.FirewallStatusValueReady},
		Refresh: FirewallUpdatedStatus(ctx, conn, arn),
		Timeout: FirewallTimeout,
		// Delay added to account for Associate/DisassociateSubnet calls that return
		// a READY status immediately after the method is called instead of immediately
		// returning PROVISIONING
		Delay: 30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*string); ok {
		return v, err
	}

	return nil, err
}

// FirewallDeleted waits for a Firewall to return "Deleted"
func FirewallDeleted(ctx context.Context, conn *networkfirewall.NetworkFirewall, arn string) (*networkfirewall.Firewall, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{networkfirewall.FirewallStatusValueDeleting},
		Target:  []string{ResourceStatusDeleted},
		Refresh: FirewallDeletedStatus(ctx, conn, arn),
		Timeout: FirewallTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*networkfirewall.Firewall); ok {
		return v, err
	}

	return nil, err
}

// FirewallPolicyDeleted waits for a Firewall Policy to return "Deleted"
func FirewallPolicyDeleted(ctx context.Context, conn *networkfirewall.NetworkFirewall, arn string) (*networkfirewall.FirewallPolicy, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{networkfirewall.ResourceStatusDeleting},
		Target:  []string{ResourceStatusDeleted},
		Refresh: FirewallPolicyStatus(ctx, conn, arn),
		Timeout: FirewallPolicyTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*networkfirewall.FirewallPolicy); ok {
		return v, err
	}

	return nil, err
}

// RuleGroupDeleted waits for a Rule Group to return "Deleted"
func RuleGroupDeleted(ctx context.Context, conn *networkfirewall.NetworkFirewall, arn string) (*networkfirewall.RuleGroup, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{networkfirewall.ResourceStatusDeleting},
		Target:  []string{ResourceStatusDeleted},
		Refresh: RuleGroupStatus(ctx, conn, arn),
		Timeout: RuleGroupDeleteTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*networkfirewall.RuleGroup); ok {
		return v, err
	}

	return nil, err
}
