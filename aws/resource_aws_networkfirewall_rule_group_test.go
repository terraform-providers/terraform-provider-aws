package aws

import (
	"context"
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/networkfirewall"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/networkfirewall/finder"
)

func init() {
	resource.AddTestSweepers("aws_networkfirewall_rule_group", &resource.Sweeper{
		Name: "aws_networkfirewall_rule_group",
		F:    testSweepNetworkFirewallRuleGroups,
		Dependencies: []string{
			"aws_networkfirewall_firewall_policy",
		},
	})
}

func testSweepNetworkFirewallRuleGroups(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.(*AWSClient).networkfirewallconn
	ctx := context.Background()
	input := &networkfirewall.ListRuleGroupsInput{MaxResults: aws.Int64(100)}
	var sweeperErrs *multierror.Error

	for {
		resp, err := conn.ListRuleGroupsWithContext(ctx, input)
		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping NetworkFirewall Rule Group sweep for %s: %s", region, err)
			return nil
		}
		if err != nil {
			return fmt.Errorf("error retrieving NetworkFirewall Rule Groups: %w", err)
		}

		for _, r := range resp.RuleGroups {
			if r == nil {
				continue
			}

			arn := aws.StringValue(r.Arn)
			log.Printf("[INFO] Deleting NetworkFirewall Rule Group: %s", arn)

			r := resourceAwsNetworkFirewallRuleGroup()
			d := r.Data(nil)
			d.SetId(arn)
			diags := r.DeleteContext(ctx, d, client)
			for i := range diags {
				if diags[i].Severity == diag.Error {
					log.Printf("[ERROR] %s", diags[i].Summary)
					sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf(diags[i].Summary))
					continue
				}
			}
		}

		if aws.StringValue(resp.NextToken) == "" {
			break
		}
		input.NextToken = resp.NextToken
	}

	return sweeperErrs.ErrorOrNil()
}

func TestAccAwsNetworkFirewallRuleGroup_basic_rulesSourceList(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_networkfirewall_rule_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAwsNetworkFirewall(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsNetworkFirewallRuleGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkFirewallRuleGroup_basic_rulesSourceList(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsNetworkFirewallRuleGroupExists(resourceName),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "network-firewall", fmt.Sprintf("stateful-rulegroup/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "capacity", "100"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "type", networkfirewall.RuleGroupTypeStateful),
					resource.TestCheckResourceAttr(resourceName, "rule_group.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rules_source.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rules_source.0.rules_source_list.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rules_source.0.rules_source_list.0.generated_rules_type", networkfirewall.GeneratedRulesTypeAllowlist),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rules_source.0.rules_source_list.0.target_types.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "rule_group.0.rules_source.0.rules_source_list.0.target_types.*", networkfirewall.TargetTypeHttpHost),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rules_source.0.rules_source_list.0.targets.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "rule_group.0.rules_source.0.rules_source_list.0.targets.*", "test.example.com"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAwsNetworkFirewallRuleGroup_basic_statefulRule(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_networkfirewall_rule_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAwsNetworkFirewall(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsNetworkFirewallRuleGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkFirewallRuleGroup_basic_statefulRule(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsNetworkFirewallRuleGroupExists(resourceName),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "network-firewall", fmt.Sprintf("stateful-rulegroup/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "capacity", "100"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "type", networkfirewall.RuleGroupTypeStateful),
					resource.TestCheckResourceAttr(resourceName, "rule_group.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rules_source.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rules_source.0.stateful_rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule_group.0.rules_source.0.stateful_rule.*", map[string]string{
						"action":                    networkfirewall.StatefulActionPass,
						"header.#":                  "1",
						"header.0.destination":      "124.1.1.24/32",
						"header.0.destination_port": "53",
						"header.0.direction":        networkfirewall.StatefulRuleDirectionAny,
						"header.0.protocol":         networkfirewall.StatefulRuleProtocolTcp,
						"header.0.source":           "1.2.3.4/32",
						"header.0.source_port":      "53",
						"rule_option.#":             "1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule_group.0.rules_source.0.stateful_rule.*.rule_option.*", map[string]string{
						"keyword": "sid:1",
					}),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAwsNetworkFirewallRuleGroup_basic_statelessRule(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_networkfirewall_rule_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAwsNetworkFirewall(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsNetworkFirewallRuleGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkFirewallRuleGroup_basic_statelessRule(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsNetworkFirewallRuleGroupExists(resourceName),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "network-firewall", fmt.Sprintf("stateless-rulegroup/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "capacity", "100"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "type", networkfirewall.RuleGroupTypeStateless),
					resource.TestCheckResourceAttr(resourceName, "rule_group.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rules_source.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rules_source.0.stateless_rules_and_custom_actions.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule_group.0.rules_source.0.stateless_rules_and_custom_actions.0.stateless_rule.*", map[string]string{
						"priority":                                           "1",
						"rule_definition.#":                                  "1",
						"rule_definition.0.actions.#":                        "1",
						"rule_definition.0.match_attributes.#":               "1",
						"rule_definition.0.match_attributes.0.destination.#": "1",
						"rule_definition.0.match_attributes.0.source.#":      "1",
					}),
					resource.TestCheckTypeSetElemAttr(resourceName, "rule_group.0.rules_source.0.stateless_rules_and_custom_actions.0.stateless_rule.*.rule_definition.0.actions.*", "aws:drop"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAwsNetworkFirewallRuleGroup_basic_rules(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_networkfirewall_rule_group.test"
	rules := `alert http any any -> any any (http_response_line; content:"403 Forbidden"; sid:1;)`

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAwsNetworkFirewall(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsNetworkFirewallRuleGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkFirewallRuleGroup_basic_rules(rName, rules),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsNetworkFirewallRuleGroupExists(resourceName),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "network-firewall", fmt.Sprintf("stateful-rulegroup/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "capacity", "100"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "type", networkfirewall.RuleGroupTypeStateful),
					resource.TestCheckResourceAttr(resourceName, "rules", rules),
					resource.TestCheckResourceAttr(resourceName, "rule_group.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rules_source.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rules_source.0.rules_string", rules),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rules_source.0.stateful_rule.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"rules"}, // argument not returned in RuleGroup API response
			},
		},
	})
}

func TestAccAwsNetworkFirewallRuleGroup_statelessRuleWithCustomAction(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_networkfirewall_rule_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAwsNetworkFirewall(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsNetworkFirewallRuleGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkFirewallRuleGroup_statelessRuleWithCustomAction(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsNetworkFirewallRuleGroupExists(resourceName),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "network-firewall", fmt.Sprintf("stateless-rulegroup/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "capacity", "100"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "type", networkfirewall.RuleGroupTypeStateless),
					resource.TestCheckResourceAttr(resourceName, "rule_group.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rules_source.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rules_source.0.stateless_rules_and_custom_actions.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule_group.0.rules_source.0.stateless_rules_and_custom_actions.0.stateless_rule.*", map[string]string{
						"priority":                                           "1",
						"rule_definition.#":                                  "1",
						"rule_definition.0.actions.#":                        "2",
						"rule_definition.0.match_attributes.#":               "1",
						"rule_definition.0.match_attributes.0.destination.#": "1",
						"rule_definition.0.match_attributes.0.source.#":      "1",
					}),
					resource.TestCheckTypeSetElemAttr(resourceName, "rule_group.0.rules_source.0.stateless_rules_and_custom_actions.0.stateless_rule.*.rule_definition.0.actions.*", "aws:pass"),
					resource.TestCheckTypeSetElemAttr(resourceName, "rule_group.0.rules_source.0.stateless_rules_and_custom_actions.0.stateless_rule.*.rule_definition.0.actions.*", "example"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule_group.0.rules_source.0.stateless_rules_and_custom_actions.0.custom_action.*", map[string]string{
						"action_name":         "example",
						"action_definition.#": "1",
						"action_definition.0.publish_metric_action.#":             "1",
						"action_definition.0.publish_metric_action.0.dimension.#": "1",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAwsNetworkFirewallRuleGroup_updateRulesSourceList(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_networkfirewall_rule_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAwsNetworkFirewall(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsNetworkFirewallRuleGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkFirewallRuleGroup_basic_rulesSourceList(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsNetworkFirewallRuleGroupExists(resourceName),
				),
			},
			{
				Config: testAccNetworkFirewallRuleGroup_updateRulesSourceList(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsNetworkFirewallRuleGroupExists(resourceName),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "network-firewall", fmt.Sprintf("stateful-rulegroup/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "capacity", "100"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "type", networkfirewall.RuleGroupTypeStateful),
					resource.TestCheckResourceAttr(resourceName, "rule_group.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rules_source.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rules_source.0.rules_source_list.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rules_source.0.rules_source_list.0.generated_rules_type", networkfirewall.GeneratedRulesTypeDenylist),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rules_source.0.rules_source_list.0.target_types.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "rule_group.0.rules_source.0.rules_source_list.0.target_types.*", networkfirewall.TargetTypeHttpHost),
					resource.TestCheckTypeSetElemAttr(resourceName, "rule_group.0.rules_source.0.rules_source_list.0.target_types.*", networkfirewall.TargetTypeTlsSni),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rules_source.0.rules_source_list.0.targets.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "rule_group.0.rules_source.0.rules_source_list.0.targets.*", "test.example.com"),
					resource.TestCheckTypeSetElemAttr(resourceName, "rule_group.0.rules_source.0.rules_source_list.0.targets.*", "test2.example.com"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAwsNetworkFirewallRuleGroup_rulesSourceAndRuleVariables(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_networkfirewall_rule_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAwsNetworkFirewall(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsNetworkFirewallRuleGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkFirewallRuleGroup_rulesSourceList_ruleVariables(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsNetworkFirewallRuleGroupExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "rule_group.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rule_variables.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rules_source.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rule_variables.0.ip_sets.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule_group.0.rule_variables.0.ip_sets.*", map[string]string{
						"key":                   "example",
						"ip_set.#":              "1",
						"ip_set.0.definition.#": "2",
					}),
					resource.TestCheckTypeSetElemAttr(resourceName, "rule_group.0.rule_variables.0.ip_sets.*.ip_set.0.definition.*", "10.0.0.0/16"),
					resource.TestCheckTypeSetElemAttr(resourceName, "rule_group.0.rule_variables.0.ip_sets.*.ip_set.0.definition.*", "10.0.1.0/24"),
				),
			},
			{
				Config: testAccNetworkFirewallRuleGroup_rulesSourceList_updateRuleVariables(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsNetworkFirewallRuleGroupExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "rule_group.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rule_variables.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rules_source.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rule_variables.0.ip_sets.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rule_variables.0.port_sets.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule_group.0.rule_variables.0.ip_sets.*", map[string]string{
						"key":                   "example",
						"ip_set.#":              "1",
						"ip_set.0.definition.#": "3",
					}),
					resource.TestCheckTypeSetElemAttr(resourceName, "rule_group.0.rule_variables.0.ip_sets.*.ip_set.0.definition.*", "10.0.0.0/16"),
					resource.TestCheckTypeSetElemAttr(resourceName, "rule_group.0.rule_variables.0.ip_sets.*.ip_set.0.definition.*", "10.0.1.0/24"),
					resource.TestCheckTypeSetElemAttr(resourceName, "rule_group.0.rule_variables.0.ip_sets.*.ip_set.0.definition.*", "192.168.0.0/16"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule_group.0.rule_variables.0.ip_sets.*", map[string]string{
						"key":                   "example2",
						"ip_set.#":              "1",
						"ip_set.0.definition.#": "1",
					}),
					resource.TestCheckTypeSetElemAttr(resourceName, "rule_group.0.rule_variables.0.ip_sets.*.ip_set.0.definition.*", "1.2.3.4/32"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule_group.0.rule_variables.0.port_sets.*", map[string]string{
						"key":                     "example",
						"port_set.#":              "1",
						"port_set.0.definition.#": "2",
					}),
					resource.TestCheckTypeSetElemAttr(resourceName, "rule_group.0.rule_variables.0.port_sets.*.port_set.0.definition.*", "443"),
					resource.TestCheckTypeSetElemAttr(resourceName, "rule_group.0.rule_variables.0.port_sets.*.port_set.0.definition.*", "80"),
				),
			},
			{
				Config: testAccNetworkFirewallRuleGroup_basic_rulesSourceList(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsNetworkFirewallRuleGroupExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "rule_group.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rule_variables.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rules_source.#", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// TestAccAwsNetworkFirewallRuleGroup_updateStatefulRule validates
// in-place updates to a single stateful_rule configuration block
func TestAccAwsNetworkFirewallRuleGroup_updateStatefulRule(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_networkfirewall_rule_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAwsNetworkFirewall(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsNetworkFirewallRuleGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkFirewallRuleGroup_basic_statefulRule(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsNetworkFirewallRuleGroupExists(resourceName),
				),
			},
			{
				Config: testAccNetworkFirewallRuleGroup_updateStatefulRule(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsNetworkFirewallRuleGroupExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rules_source.0.stateful_rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule_group.0.rules_source.0.stateful_rule.*", map[string]string{
						"action":                    networkfirewall.StatefulActionDrop,
						"header.#":                  "1",
						"header.0.destination":      "1.2.3.4/32",
						"header.0.destination_port": "1001",
						"header.0.direction":        networkfirewall.StatefulRuleDirectionForward,
						"header.0.protocol":         networkfirewall.StatefulRuleProtocolIp,
						"header.0.source":           "124.1.1.24/32",
						"header.0.source_port":      "1001",
						"rule_option.#":             "1",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// TestAccAwsNetworkFirewallRuleGroup_updateMultipleStatefulRules validates
// in-place updates to stateful_rule configuration blocks
// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/16868
func TestAccAwsNetworkFirewallRuleGroup_updateMultipleStatefulRules(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_networkfirewall_rule_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAwsNetworkFirewall(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsNetworkFirewallRuleGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkFirewallRuleGroup_basic_statefulRule(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsNetworkFirewallRuleGroupExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rules_source.0.stateful_rule.#", "1"),
				),
			},
			{
				Config: testAccNetworkFirewallRuleGroup_multipleStatefulRules(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsNetworkFirewallRuleGroupExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rules_source.0.stateful_rule.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule_group.0.rules_source.0.stateful_rule.*", map[string]string{
						"action":                    networkfirewall.StatefulActionPass,
						"header.#":                  "1",
						"header.0.destination":      "124.1.1.24/32",
						"header.0.destination_port": "53",
						"header.0.direction":        networkfirewall.StatefulRuleDirectionAny,
						"header.0.protocol":         networkfirewall.StatefulRuleProtocolTcp,
						"header.0.source":           "1.2.3.4/32",
						"header.0.source_port":      "53",
						"rule_option.#":             "1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule_group.0.rules_source.0.stateful_rule.*", map[string]string{
						"action":                    networkfirewall.StatefulActionAlert,
						"header.#":                  "1",
						"header.0.destination":      networkfirewall.StatefulRuleDirectionAny,
						"header.0.destination_port": networkfirewall.StatefulRuleDirectionAny,
						"header.0.direction":        networkfirewall.StatefulRuleDirectionAny,
						"header.0.protocol":         networkfirewall.StatefulRuleProtocolIp,
						"header.0.source":           networkfirewall.StatefulRuleDirectionAny,
						"header.0.source_port":      networkfirewall.StatefulRuleDirectionAny,
						"rule_option.#":             "1",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccNetworkFirewallRuleGroup_updateStatefulRule(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsNetworkFirewallRuleGroupExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rules_source.0.stateful_rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule_group.0.rules_source.0.stateful_rule.*", map[string]string{
						"action":                    networkfirewall.StatefulActionDrop,
						"header.#":                  "1",
						"header.0.destination":      "1.2.3.4/32",
						"header.0.destination_port": "1001",
						"header.0.direction":        networkfirewall.StatefulRuleDirectionForward,
						"header.0.protocol":         networkfirewall.StatefulRuleProtocolIp,
						"header.0.source":           "124.1.1.24/32",
						"header.0.source_port":      "1001",
						"rule_option.#":             "1",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// TestAccAwsNetworkFirewallRuleGroup_statefulRule_action validates in-place
// updates to the "action" argument within 1 stateful_rule configuration block
// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/16868
func TestAccAwsNetworkFirewallRuleGroup_statefulRule_action(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_networkfirewall_rule_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAwsNetworkFirewall(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsNetworkFirewallRuleGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkFirewallRuleGroup_statefulRule_action(rName, networkfirewall.StatefulActionAlert),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsNetworkFirewallRuleGroupExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rules_source.0.stateful_rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule_group.0.rules_source.0.stateful_rule.*", map[string]string{
						"action": networkfirewall.StatefulActionAlert,
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccNetworkFirewallRuleGroup_statefulRule_action(rName, networkfirewall.StatefulActionPass),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsNetworkFirewallRuleGroupExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rules_source.0.stateful_rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule_group.0.rules_source.0.stateful_rule.*", map[string]string{
						"action": networkfirewall.StatefulActionPass,
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccNetworkFirewallRuleGroup_statefulRule_action(rName, networkfirewall.StatefulActionDrop),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsNetworkFirewallRuleGroupExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rules_source.0.stateful_rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule_group.0.rules_source.0.stateful_rule.*", map[string]string{
						"action": networkfirewall.StatefulActionDrop,
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/16470
func TestAccAwsNetworkFirewallRuleGroup_statefulRule_header(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_networkfirewall_rule_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsNetworkFirewallRuleGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkFirewallRuleGroup_statefulRule_header(rName, "1990", "1994"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsNetworkFirewallRuleGroupExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rules_source.0.stateful_rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule_group.0.rules_source.0.stateful_rule.*", map[string]string{
						"action":                    networkfirewall.StatefulActionPass,
						"header.#":                  "1",
						"header.0.destination":      "ANY",
						"header.0.destination_port": "1990",
						"header.0.direction":        networkfirewall.StatefulRuleDirectionAny,
						"header.0.protocol":         networkfirewall.StatefulRuleProtocolTcp,
						"header.0.source":           "ANY",
						"header.0.source_port":      "1994",
						"rule_option.#":             "1",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccNetworkFirewallRuleGroup_statefulRule_header(rName, "ANY", "ANY"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsNetworkFirewallRuleGroupExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rules_source.0.stateful_rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule_group.0.rules_source.0.stateful_rule.*", map[string]string{
						"action":                    networkfirewall.StatefulActionPass,
						"header.#":                  "1",
						"header.0.destination":      "ANY",
						"header.0.destination_port": "ANY",
						"header.0.direction":        networkfirewall.StatefulRuleDirectionAny,
						"header.0.protocol":         networkfirewall.StatefulRuleProtocolTcp,
						"header.0.source":           "ANY",
						"header.0.source_port":      "ANY",
						"rule_option.#":             "1",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAwsNetworkFirewallRuleGroup_updateStatelessRule(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_networkfirewall_rule_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAwsNetworkFirewall(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsNetworkFirewallRuleGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkFirewallRuleGroup_basic_statelessRule(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsNetworkFirewallRuleGroupExists(resourceName),
				),
			},
			{
				Config: testAccNetworkFirewallRuleGroup_updateStatelessRule(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsNetworkFirewallRuleGroupExists(resourceName),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule_group.0.rules_source.0.stateless_rules_and_custom_actions.0.stateless_rule.*", map[string]string{
						"priority":                                                "10",
						"rule_definition.#":                                       "1",
						"rule_definition.0.actions.#":                             "1",
						"rule_definition.0.match_attributes.#":                    "1",
						"rule_definition.0.match_attributes.0.destination.#":      "1",
						"rule_definition.0.match_attributes.0.destination_port.#": "1",
						"rule_definition.0.match_attributes.0.protocols.#":        "1",
						"rule_definition.0.match_attributes.0.source.#":           "1",
						"rule_definition.0.match_attributes.0.source_port.#":      "1",
						"rule_definition.0.match_attributes.0.tcp_flag.#":         "1",
					}),
					resource.TestCheckTypeSetElemAttr(resourceName, "rule_group.0.rules_source.0.stateless_rules_and_custom_actions.0.stateless_rule.*.rule_definition.0.actions.*", "aws:pass"),
					resource.TestCheckTypeSetElemAttr(resourceName, "rule_group.0.rules_source.0.stateless_rules_and_custom_actions.0.stateless_rule.*.rule_definition.0.match_attributes.0.protocols.*", "6"),
					resource.TestCheckTypeSetElemAttr(resourceName, "rule_group.0.rules_source.0.stateless_rules_and_custom_actions.0.stateless_rule.*.rule_definition.0.match_attributes.0.tcp_flag.*.flags.*", "SYN"),
					resource.TestCheckTypeSetElemAttr(resourceName, "rule_group.0.rules_source.0.stateless_rules_and_custom_actions.0.stateless_rule.*.rule_definition.0.match_attributes.0.tcp_flag.*.masks.*", "SYN"),
					resource.TestCheckTypeSetElemAttr(resourceName, "rule_group.0.rules_source.0.stateless_rules_and_custom_actions.0.stateless_rule.*.rule_definition.0.match_attributes.0.tcp_flag.*.masks.*", "ACK"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAwsNetworkFirewallRuleGroup_tags(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_networkfirewall_rule_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAwsNetworkFirewall(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsNetworkFirewallRuleGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkFirewallRuleGroup_oneTag(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsNetworkFirewallRuleGroupExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
				),
			},
			{
				Config: testAccNetworkFirewallRuleGroup_twoTags(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsNetworkFirewallRuleGroupExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.Description", "updated"),
				),
			},
			{
				Config: testAccNetworkFirewallRuleGroup_basic_rulesSourceList(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsNetworkFirewallRuleGroupExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAwsNetworkFirewallRuleGroup_disappears(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_networkfirewall_rule_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAwsNetworkFirewall(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsNetworkFirewallRuleGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkFirewallRuleGroup_basic_rulesSourceList(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsNetworkFirewallRuleGroupExists(resourceName),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsNetworkFirewallRuleGroup(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAwsNetworkFirewallRuleGroupDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_networkfirewall_rule_group" {
			continue
		}

		conn := testAccProvider.Meta().(*AWSClient).networkfirewallconn
		output, err := finder.RuleGroup(context.Background(), conn, rs.Primary.ID)
		if tfawserr.ErrCodeEquals(err, networkfirewall.ErrCodeResourceNotFoundException) {
			continue
		}
		if err != nil {
			return err
		}
		if output != nil {
			return fmt.Errorf("NetworkFirewall Rule Group still exists: %s", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckAwsNetworkFirewallRuleGroupExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No NetworkFirewall Rule Group ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).networkfirewallconn
		output, err := finder.RuleGroup(context.Background(), conn, rs.Primary.ID)
		if err != nil {
			return err
		}

		if output == nil {
			return fmt.Errorf("NetworkFirewall Rule Group (%s) not found", rs.Primary.ID)
		}

		return nil
	}
}

func testAccNetworkFirewallRuleGroup_basic_rulesSourceList(rName string) string {
	return fmt.Sprintf(`
resource "aws_networkfirewall_rule_group" "test" {
  capacity = 100
  name     = %q
  type     = "STATEFUL"
  rule_group {
    rules_source {
      rules_source_list {
        generated_rules_type = "ALLOWLIST"
        target_types         = ["HTTP_HOST"]
        targets              = ["test.example.com"]
      }
    }
  }
}
`, rName)
}

func testAccNetworkFirewallRuleGroup_rulesSourceList_ruleVariables(rName string) string {
	return fmt.Sprintf(`
resource "aws_networkfirewall_rule_group" "test" {
  capacity = 100
  name     = %[1]q
  type     = "STATEFUL"
  rule_group {
    rule_variables {
      ip_sets {
        key = "example"
        ip_set {
          definition = ["10.0.0.0/16", "10.0.1.0/24"]
        }
      }
    }
    rules_source {
      rules_source_list {
        generated_rules_type = "ALLOWLIST"
        target_types         = ["HTTP_HOST"]
        targets              = ["test.example.com"]
      }
    }
  }
}
`, rName)
}

func testAccNetworkFirewallRuleGroup_rulesSourceList_updateRuleVariables(rName string) string {
	return fmt.Sprintf(`
resource "aws_networkfirewall_rule_group" "test" {
  capacity = 100
  name     = %[1]q
  type     = "STATEFUL"
  rule_group {
    rule_variables {
      ip_sets {
        key = "example"
        ip_set {
          definition = ["10.0.0.0/16", "10.0.1.0/24", "192.168.0.0/16"]
        }
      }
      ip_sets {
        key = "example2"
        ip_set {
          definition = ["1.2.3.4/32"]
        }
      }
      port_sets {
        key = "example"
        port_set {
          definition = ["443", "80"]
        }
      }
    }
    rules_source {
      rules_source_list {
        generated_rules_type = "ALLOWLIST"
        target_types         = ["HTTP_HOST"]
        targets              = ["test.example.com"]
      }
    }
  }
}
`, rName)
}

func testAccNetworkFirewallRuleGroup_updateRulesSourceList(rName string) string {
	return fmt.Sprintf(`
resource "aws_networkfirewall_rule_group" "test" {
  capacity = 100
  name     = %q
  type     = "STATEFUL"
  rule_group {
    rules_source {
      rules_source_list {
        generated_rules_type = "DENYLIST"
        target_types         = ["HTTP_HOST", "TLS_SNI"]
        targets              = ["test.example.com", "test2.example.com"]
      }
    }
  }
}
`, rName)
}

func testAccNetworkFirewallRuleGroup_basic_statefulRule(rName string) string {
	return fmt.Sprintf(`
resource "aws_networkfirewall_rule_group" "test" {
  capacity    = 100
  name        = %[1]q
  description = %[1]q
  type        = "STATEFUL"
  rule_group {
    rules_source {
      stateful_rule {
        action = "PASS"
        header {
          destination      = "124.1.1.24/32"
          destination_port = 53
          direction        = "ANY"
          protocol         = "TCP"
          source           = "1.2.3.4/32"
          source_port      = 53
        }
        rule_option {
          keyword = "sid:1"
        }
      }
    }
  }
}
`, rName)
}

func testAccNetworkFirewallRuleGroup_statefulRule_action(rName, action string) string {
	return fmt.Sprintf(`
resource "aws_networkfirewall_rule_group" "test" {
  capacity    = 100
  name        = %[1]q
  description = %[1]q
  type        = "STATEFUL"
  rule_group {
    rules_source {
      stateful_rule {
        action = %q
        header {
          destination      = "124.1.1.24/32"
          destination_port = 53
          direction        = "ANY"
          protocol         = "TCP"
          source           = "1.2.3.4/32"
          source_port      = 53
        }
        rule_option {
          keyword = "sid:1"
        }
      }
    }
  }
}
`, rName, action)
}

func testAccNetworkFirewallRuleGroup_statefulRule_header(rName, dstPort, srcPort string) string {
	return fmt.Sprintf(`
resource "aws_networkfirewall_rule_group" "test" {
  capacity    = 100
  name        = %[1]q
  description = %[1]q
  type        = "STATEFUL"
  rule_group {
    rules_source {
      stateful_rule {
        action = "PASS"
        header {
          destination      = "ANY"
          destination_port = %q
          direction        = "ANY"
          protocol         = "TCP"
          source           = "ANY"
          source_port      = %q
        }
        rule_option {
          keyword = "sid:1"
        }
      }
    }
  }
}
`, rName, dstPort, srcPort)
}

func testAccNetworkFirewallRuleGroup_updateStatefulRule(rName string) string {
	return fmt.Sprintf(`
resource "aws_networkfirewall_rule_group" "test" {
  capacity = 100
  name     = %[1]q
  type     = "STATEFUL"
  rule_group {
    rules_source {
      stateful_rule {
        action = "DROP"
        header {
          destination      = "1.2.3.4/32"
          destination_port = 1001
          direction        = "FORWARD"
          protocol         = "IP"
          source           = "124.1.1.24/32"
          source_port      = 1001
        }
        rule_option {
          keyword = "sid:1;rev:2"
        }
      }
    }
  }
}
`, rName)
}

func testAccNetworkFirewallRuleGroup_multipleStatefulRules(rName string) string {
	return fmt.Sprintf(`
resource "aws_networkfirewall_rule_group" "test" {
  capacity = 100
  name     = %[1]q
  type     = "STATEFUL"
  rule_group {
    rules_source {
      stateful_rule {
        action = "PASS"
        header {
          destination      = "124.1.1.24/32"
          destination_port = 53
          direction        = "ANY"
          protocol         = "TCP"
          source           = "1.2.3.4/32"
          source_port      = 53
        }
        rule_option {
          keyword = "sid:1"
        }
      }
      stateful_rule {
        action = "ALERT"
        header {
          destination      = "ANY"
          destination_port = "ANY"
          direction        = "ANY"
          protocol         = "IP"
          source           = "ANY"
          source_port      = "ANY"
        }
        rule_option {
          keyword = "sid:2"
        }
      }
    }
  }
}
`, rName)
}

func testAccNetworkFirewallRuleGroup_basic_statelessRule(rName string) string {
	return fmt.Sprintf(`
resource "aws_networkfirewall_rule_group" "test" {
  capacity = 100
  name     = %[1]q
  type     = "STATELESS"
  rule_group {
    rules_source {
      stateless_rules_and_custom_actions {
        stateless_rule {
          priority = 1
          rule_definition {
            actions = ["aws:drop"]
            match_attributes {
              destination {
                address_definition = "1.2.3.4/32"
              }
              source {
                address_definition = "124.1.1.5/32"
              }
            }
          }
        }
      }
    }
  }
}
`, rName)
}

func testAccNetworkFirewallRuleGroup_updateStatelessRule(rName string) string {
	return fmt.Sprintf(`
resource "aws_networkfirewall_rule_group" "test" {
  capacity = 100
  name     = %[1]q
  type     = "STATELESS"
  rule_group {
    rules_source {
      stateless_rules_and_custom_actions {
        stateless_rule {
          priority = 10
          rule_definition {
            actions = ["aws:pass"]
            match_attributes {
              destination {
                address_definition = "1.2.3.4/32"
              }
              destination_port {
                from_port = 53
                to_port   = 53
              }
              protocols = [6]
              source {
                address_definition = "124.1.1.5/32"
              }
              source_port {
                from_port = 53
                to_port   = 53
              }
              tcp_flag {
                flags = ["SYN"]
                masks = ["SYN", "ACK"]
              }
            }
          }
        }
      }
    }
  }
}
`, rName)
}

func testAccNetworkFirewallRuleGroup_basic_rules(rName, rules string) string {
	return fmt.Sprintf(`
resource "aws_networkfirewall_rule_group" "test" {
  capacity = 100
  name     = %q
  type     = "STATEFUL"
  rules    = %q
}
`, rName, rules)
}

func testAccNetworkFirewallRuleGroup_statelessRuleWithCustomAction(rName string) string {
	return fmt.Sprintf(`
resource "aws_networkfirewall_rule_group" "test" {
  capacity = 100
  name     = %[1]q
  type     = "STATELESS"
  rule_group {
    rules_source {
      stateless_rules_and_custom_actions {
        custom_action {
          action_name = "example"
          action_definition {
            publish_metric_action {
              dimension {
                value = "2"
              }
            }
          }
        }
        stateless_rule {
          priority = 1
          rule_definition {
            actions = ["aws:pass", "example"]
            match_attributes {
              destination {
                address_definition = "1.2.3.4/32"
              }
              source {
                address_definition = "124.1.1.5/32"
              }
            }
          }
        }
      }
    }
  }
}
`, rName)
}

func testAccNetworkFirewallRuleGroup_oneTag(rName string) string {
	return fmt.Sprintf(`
resource "aws_networkfirewall_rule_group" "test" {
  capacity = 100
  name     = %q
  type     = "STATEFUL"
  rule_group {
    rules_source {
      rules_source_list {
        generated_rules_type = "ALLOWLIST"
        target_types         = ["HTTP_HOST"]
        targets              = ["test.example.com"]
      }
    }
  }
  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccNetworkFirewallRuleGroup_twoTags(rName string) string {
	return fmt.Sprintf(`
resource "aws_networkfirewall_rule_group" "test" {
  capacity = 100
  name     = %q
  type     = "STATEFUL"
  rule_group {
    rules_source {
      rules_source_list {
        generated_rules_type = "ALLOWLIST"
        target_types         = ["HTTP_HOST"]
        targets              = ["test.example.com"]
      }
    }
  }
  tags = {
    Name        = %[1]q
    Description = "updated"
  }
}
`, rName)
}
