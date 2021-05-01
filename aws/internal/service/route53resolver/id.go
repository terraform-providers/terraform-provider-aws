package route53resolver

import (
	"fmt"
	"strings"
)

const ruleIdSeparator = ":"

func FirewallRuleCreateID(firewallRuleGroupId, firewallDomainListId string) string {
	parts := []string{firewallRuleGroupId, firewallDomainListId}
	id := strings.Join(parts, ruleIdSeparator)

	return id
}

func FirewallRuleParseID(id string) (string, string, error) {
	parts := strings.SplitN(id, ruleIdSeparator, 2)

	if len(parts) < 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format of ID (%s), expected firewall_rule_group_id%sfirewall_domain_list_id", id, ruleIdSeparator)
	}

	return parts[0], parts[1], nil
}
