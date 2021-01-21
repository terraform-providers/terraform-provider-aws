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
	resource.AddTestSweepers("aws_networkfirewall_firewall_policy", &resource.Sweeper{
		Name: "aws_networkfirewall_firewall_policy",
		F:    testSweepNetworkFirewallFirewallPolicies,
		Dependencies: []string{
			"aws_networkfirewall_firewall",
		},
	})
}

func testSweepNetworkFirewallFirewallPolicies(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.(*AWSClient).networkfirewallconn
	ctx := context.Background()
	input := &networkfirewall.ListFirewallPoliciesInput{MaxResults: aws.Int64(100)}
	var sweeperErrs *multierror.Error

	for {
		resp, err := conn.ListFirewallPoliciesWithContext(ctx, input)
		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping NetworkFirewall Firewall Policy sweep for %s: %s", region, err)
			return nil
		}
		if err != nil {
			return fmt.Errorf("error retrieving NetworkFirewall Firewall Policies: %w", err)
		}

		for _, fp := range resp.FirewallPolicies {
			if fp == nil {
				continue
			}

			arn := aws.StringValue(fp.Arn)
			log.Printf("[INFO] Deleting NetworkFirewall Firewall Policy: %s", arn)

			r := resourceAwsNetworkFirewallFirewallPolicy()
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

func TestAccAwsNetworkFirewallFirewallPolicy_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_networkfirewall_firewall_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAwsNetworkFirewall(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsNetworkFirewallFirewallPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsNetworkFirewallFirewallPolicy_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsNetworkFirewallFirewallPolicyExists(resourceName),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "network-firewall", fmt.Sprintf("firewall-policy/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.0.stateless_fragment_default_actions.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "firewall_policy.0.stateless_fragment_default_actions.*", "aws:drop"),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.0.stateless_default_actions.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "firewall_policy.0.stateless_default_actions.*", "aws:pass"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
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

func TestAccAwsNetworkFirewallFirewallPolicy_statefulRuleGroupReference(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_networkfirewall_firewall_policy.test"
	ruleGroupResourceName := "aws_networkfirewall_rule_group.test.0"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAwsNetworkFirewall(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsNetworkFirewallFirewallPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsNetworkFirewallFirewallPolicy_statefulRuleGroupReference(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsNetworkFirewallFirewallPolicyExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.0.stateful_rule_group_reference.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "firewall_policy.0.stateful_rule_group_reference.*.resource_arn", ruleGroupResourceName, "arn"),
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

func TestAccAwsNetworkFirewallFirewallPolicy_updateStatefulRuleGroupReference(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_networkfirewall_firewall_policy.test"
	ruleGroupResourceName := "aws_networkfirewall_rule_group.test.0"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAwsNetworkFirewall(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsNetworkFirewallFirewallPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsNetworkFirewallFirewallPolicy_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsNetworkFirewallFirewallPolicyExists(resourceName),
				),
			},
			{
				Config: testAccAwsNetworkFirewallFirewallPolicy_statefulRuleGroupReference(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsNetworkFirewallFirewallPolicyExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.0.stateful_rule_group_reference.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "firewall_policy.0.stateful_rule_group_reference.*.resource_arn", ruleGroupResourceName, "arn"),
				),
			},
			{
				Config: testAccAwsNetworkFirewallFirewallPolicy_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsNetworkFirewallFirewallPolicyExists(resourceName),
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

func TestAccAwsNetworkFirewallFirewallPolicy_multipleStatefulRuleGroupReferences(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_networkfirewall_firewall_policy.test"
	ruleGroupResourceName1 := "aws_networkfirewall_rule_group.test.0"
	ruleGroupResourceName2 := "aws_networkfirewall_rule_group.test.1"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAwsNetworkFirewall(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsNetworkFirewallFirewallPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsNetworkFirewallFirewallPolicy_multipleStatefulRuleGroupReferences(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsNetworkFirewallFirewallPolicyExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.0.stateful_rule_group_reference.#", "2"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "firewall_policy.0.stateful_rule_group_reference.*.resource_arn", ruleGroupResourceName1, "arn"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "firewall_policy.0.stateful_rule_group_reference.*.resource_arn", ruleGroupResourceName2, "arn"),
				),
			},
			{
				Config: testAccAwsNetworkFirewallFirewallPolicy_singleStatefulRuleGroupReference(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsNetworkFirewallFirewallPolicyExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.0.stateful_rule_group_reference.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "firewall_policy.0.stateful_rule_group_reference.*.resource_arn", ruleGroupResourceName1, "arn"),
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

func TestAccAwsNetworkFirewallFirewallPolicy_statelessRuleGroupReference(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_networkfirewall_firewall_policy.test"
	ruleGroupResourceName := "aws_networkfirewall_rule_group.test.0"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAwsNetworkFirewall(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsNetworkFirewallFirewallPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsNetworkFirewallFirewallPolicy_statelessRuleGroupReference(rName, 20),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsNetworkFirewallFirewallPolicyExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.0.stateless_rule_group_reference.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "firewall_policy.0.stateless_rule_group_reference.*.resource_arn", ruleGroupResourceName, "arn"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "firewall_policy.0.stateless_rule_group_reference.*", map[string]string{
						"priority": "20",
					}),
				),
			},
			{
				Config: testAccAwsNetworkFirewallFirewallPolicy_statelessRuleGroupReference(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsNetworkFirewallFirewallPolicyExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.0.stateless_rule_group_reference.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "firewall_policy.0.stateless_rule_group_reference.*", map[string]string{
						"priority": "1",
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

func TestAccAwsNetworkFirewallFirewallPolicy_updateStatelessRuleGroupReference(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_networkfirewall_firewall_policy.test"
	ruleGroupResourceName := "aws_networkfirewall_rule_group.test.0"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAwsNetworkFirewall(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsNetworkFirewallFirewallPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsNetworkFirewallFirewallPolicy_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsNetworkFirewallFirewallPolicyExists(resourceName),
				),
			},
			{
				Config: testAccAwsNetworkFirewallFirewallPolicy_statelessRuleGroupReference(rName, 20),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsNetworkFirewallFirewallPolicyExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "firewall_policy.0.stateless_rule_group_reference.*.resource_arn", ruleGroupResourceName, "arn"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "firewall_policy.0.stateless_rule_group_reference.*", map[string]string{
						"priority": "20",
					}),
				),
			},
			{
				Config: testAccAwsNetworkFirewallFirewallPolicy_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsNetworkFirewallFirewallPolicyExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.0.stateless_rule_group_reference.#", "0"),
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

func TestAccAwsNetworkFirewallFirewallPolicy_multipleStatelessRuleGroupReferences(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_networkfirewall_firewall_policy.test"
	ruleGroupResourceName1 := "aws_networkfirewall_rule_group.test.0"
	ruleGroupResourceName2 := "aws_networkfirewall_rule_group.test.1"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAwsNetworkFirewall(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsNetworkFirewallFirewallPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsNetworkFirewallFirewallPolicy_multipleStatelessRuleGroupReferences(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsNetworkFirewallFirewallPolicyExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.0.stateless_rule_group_reference.#", "2"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "firewall_policy.0.stateless_rule_group_reference.*.resource_arn", ruleGroupResourceName1, "arn"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "firewall_policy.0.stateless_rule_group_reference.*", map[string]string{
						"priority": "1",
					}),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "firewall_policy.0.stateless_rule_group_reference.*.resource_arn", ruleGroupResourceName2, "arn"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "firewall_policy.0.stateless_rule_group_reference.*", map[string]string{
						"priority": "2",
					}),
				),
			},
			{
				Config: testAccAwsNetworkFirewallFirewallPolicy_singleStatelessRuleGroupReference(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsNetworkFirewallFirewallPolicyExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.0.stateless_rule_group_reference.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "firewall_policy.0.stateless_rule_group_reference.*", map[string]string{
						"priority": "1",
					}),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "firewall_policy.0.stateless_rule_group_reference.*.resource_arn", ruleGroupResourceName1, "arn"),
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

func TestAccAwsNetworkFirewallFirewallPolicy_statelessCustomAction(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_networkfirewall_firewall_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAwsNetworkFirewall(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsNetworkFirewallFirewallPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsNetworkFirewallFirewallPolicy_statelessCustomAction(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsNetworkFirewallFirewallPolicyExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.0.stateless_custom_action.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "firewall_policy.0.stateless_custom_action.*", map[string]string{
						"action_definition.#":                                     "1",
						"action_definition.0.publish_metric_action.#":             "1",
						"action_definition.0.publish_metric_action.0.dimension.#": "1",
						"action_name": "CustomAction",
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

func TestAccAwsNetworkFirewallFirewallPolicy_updateStatelessCustomAction(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_networkfirewall_firewall_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAwsNetworkFirewall(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsNetworkFirewallFirewallPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsNetworkFirewallFirewallPolicy_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsNetworkFirewallFirewallPolicyExists(resourceName),
				),
			},
			{
				Config: testAccAwsNetworkFirewallFirewallPolicy_statelessCustomAction(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsNetworkFirewallFirewallPolicyExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.0.stateless_custom_action.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "firewall_policy.0.stateless_custom_action.*", map[string]string{
						"action_name":         "CustomAction",
						"action_definition.#": "1",
						"action_definition.0.publish_metric_action.#":             "1",
						"action_definition.0.publish_metric_action.0.dimension.#": "1",
					}),
				),
			},
			{
				Config: testAccAwsNetworkFirewallFirewallPolicy_updateStatelessCustomAction(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsNetworkFirewallFirewallPolicyExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.0.stateless_custom_action.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "firewall_policy.0.stateless_custom_action.*", map[string]string{
						"action_name":         "updated",
						"action_definition.#": "1",
						"action_definition.0.publish_metric_action.#":             "1",
						"action_definition.0.publish_metric_action.0.dimension.#": "1",
					}),
				),
			},
			{
				Config: testAccAwsNetworkFirewallFirewallPolicy_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsNetworkFirewallFirewallPolicyExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.0.stateless_custom_action.#", "0"),
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

func TestAccAwsNetworkFirewallFirewallPolicy_multipleStatelessCustomActions(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_networkfirewall_firewall_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAwsNetworkFirewall(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsNetworkFirewallFirewallPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsNetworkFirewallFirewallPolicy_multipleStatelessCustomActions(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsNetworkFirewallFirewallPolicyExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.0.stateless_custom_action.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "firewall_policy.0.stateless_custom_action.*", map[string]string{
						"action_name":         "CustomAction",
						"action_definition.#": "1",
						"action_definition.0.publish_metric_action.#":             "1",
						"action_definition.0.publish_metric_action.0.dimension.#": "1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "firewall_policy.0.stateless_custom_action.*", map[string]string{
						"action_name":         "CustomAction2",
						"action_definition.#": "1",
						"action_definition.0.publish_metric_action.#":             "1",
						"action_definition.0.publish_metric_action.0.dimension.#": "1",
					}),
				),
			},
			{
				Config: testAccAwsNetworkFirewallFirewallPolicy_statelessCustomAction(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsNetworkFirewallFirewallPolicyExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.0.stateless_custom_action.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "firewall_policy.0.stateless_custom_action.*", map[string]string{
						"action_name":         "CustomAction",
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

func TestAccAwsNetworkFirewallFirewallPolicy_statefulRuleGroupReferenceAndCustomAction(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_networkfirewall_firewall_policy.test"
	ruleGroupResourceName := "aws_networkfirewall_rule_group.test.0"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAwsNetworkFirewall(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsNetworkFirewallFirewallPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsNetworkFirewallFirewallPolicy_statefulRuleGroupReferenceAndStatelessCustomAction(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsNetworkFirewallFirewallPolicyExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.0.stateful_rule_group_reference.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "firewall_policy.0.stateful_rule_group_reference.*.resource_arn", ruleGroupResourceName, "arn"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "firewall_policy.0.stateless_custom_action.*", map[string]string{
						"action_name":         "CustomAction",
						"action_definition.#": "1",
						"action_definition.0.publish_metric_action.#":             "1",
						"action_definition.0.publish_metric_action.0.dimension.#": "1",
					}),
				),
			},
			{
				Config: testAccAwsNetworkFirewallFirewallPolicy_statefulRuleGroupReference(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsNetworkFirewallFirewallPolicyExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.0.stateful_rule_group_reference.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "firewall_policy.0.stateful_rule_group_reference.*.resource_arn", ruleGroupResourceName, "arn"),
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

func TestAccAwsNetworkFirewallFirewallPolicy_tags(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_networkfirewall_firewall_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAwsNetworkFirewall(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsNetworkFirewallFirewallPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsNetworkFirewallFirewallPolicy_oneTag(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsNetworkFirewallFirewallPolicyExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
				),
			},
			{
				Config: testAccAwsNetworkFirewallFirewallPolicy_twoTags(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsNetworkFirewallFirewallPolicyExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.Description", "updated"),
				),
			},
			{
				Config: testAccAwsNetworkFirewallFirewallPolicy_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsNetworkFirewallFirewallPolicyExists(resourceName),
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

func TestAccAwsNetworkFirewallFirewallPolicy_disappears(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_networkfirewall_firewall_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAwsNetworkFirewall(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsNetworkFirewallFirewallPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsNetworkFirewallFirewallPolicy_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsNetworkFirewallFirewallPolicyExists(resourceName),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsNetworkFirewallFirewallPolicy(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAwsNetworkFirewallFirewallPolicyDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_networkfirewall_firewall_policy" {
			continue
		}

		conn := testAccProvider.Meta().(*AWSClient).networkfirewallconn
		output, err := finder.FirewallPolicy(context.Background(), conn, rs.Primary.ID)
		if tfawserr.ErrCodeEquals(err, networkfirewall.ErrCodeResourceNotFoundException) {
			continue
		}
		if err != nil {
			return err
		}
		if output != nil {
			return fmt.Errorf("NetworkFirewall Firewall Policy still exists: %s", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckAwsNetworkFirewallFirewallPolicyExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No NetworkFirewall Firewall Policy ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).networkfirewallconn
		output, err := finder.FirewallPolicy(context.Background(), conn, rs.Primary.ID)
		if err != nil {
			return err
		}

		if output == nil {
			return fmt.Errorf("NetworkFirewall Firewall Policy (%s) not found", rs.Primary.ID)
		}

		return nil
	}
}

func testAccAwsNetworkFirewallFirewallPolicyStatelessRuleGroupDependencies(rName string, count int) string {
	return fmt.Sprintf(`
resource "aws_networkfirewall_rule_group" "test" {
  count    = %d
  capacity = 100
  name     = "%s-${count.index}"
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
  lifecycle {
    create_before_destroy = true
  }
}
`, count, rName)
}

func testAccAwsNetworkFirewallFirewallPolicyStatefulRuleGroupDependencies(rName string, count int) string {
	return fmt.Sprintf(`
resource "aws_networkfirewall_rule_group" "test" {
  count    = %d
  capacity = 100
  name     = "%s-${count.index}"
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
  lifecycle {
    create_before_destroy = true
  }
}
`, count, rName)
}

func testAccAwsNetworkFirewallFirewallPolicy_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_networkfirewall_firewall_policy" "test" {
  name = %q
  firewall_policy {
    stateless_fragment_default_actions = ["aws:drop"]
    stateless_default_actions          = ["aws:pass"]
  }
}
`, rName)
}

func testAccAwsNetworkFirewallFirewallPolicy_oneTag(rName string) string {
	return fmt.Sprintf(`
resource "aws_networkfirewall_firewall_policy" "test" {
  name = %[1]q
  firewall_policy {
    stateless_fragment_default_actions = ["aws:drop"]
    stateless_default_actions          = ["aws:pass"]
  }
  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccAwsNetworkFirewallFirewallPolicy_twoTags(rName string) string {
	return fmt.Sprintf(`
resource "aws_networkfirewall_firewall_policy" "test" {
  name = %[1]q
  firewall_policy {
    stateless_fragment_default_actions = ["aws:drop"]
    stateless_default_actions          = ["aws:pass"]
  }
  tags = {
    Name        = %[1]q
    Description = "updated"
  }
}
`, rName)
}

func testAccAwsNetworkFirewallFirewallPolicy_statefulRuleGroupReference(rName string) string {
	return composeConfig(
		testAccAwsNetworkFirewallFirewallPolicyStatefulRuleGroupDependencies(rName, 1),
		fmt.Sprintf(`
resource "aws_networkfirewall_firewall_policy" "test" {
  name = %q
  firewall_policy {
    stateless_fragment_default_actions = ["aws:drop"]
    stateless_default_actions          = ["aws:pass"]
    stateful_rule_group_reference {
      resource_arn = aws_networkfirewall_rule_group.test[0].arn
    }
  }
}
`, rName))
}

func testAccAwsNetworkFirewallFirewallPolicy_multipleStatefulRuleGroupReferences(rName string) string {
	return composeConfig(
		testAccAwsNetworkFirewallFirewallPolicyStatefulRuleGroupDependencies(rName, 2),
		fmt.Sprintf(`
resource "aws_networkfirewall_firewall_policy" "test" {
  name = %q
  firewall_policy {
    stateless_fragment_default_actions = ["aws:drop"]
    stateless_default_actions          = ["aws:pass"]
    stateful_rule_group_reference {
      resource_arn = aws_networkfirewall_rule_group.test[0].arn
    }
    stateful_rule_group_reference {
      resource_arn = aws_networkfirewall_rule_group.test[1].arn
    }
  }
}
`, rName))
}

func testAccAwsNetworkFirewallFirewallPolicy_singleStatefulRuleGroupReference(rName string) string {
	return composeConfig(
		testAccAwsNetworkFirewallFirewallPolicyStatefulRuleGroupDependencies(rName, 2),
		fmt.Sprintf(`
resource "aws_networkfirewall_firewall_policy" "test" {
  name = %q
  firewall_policy {
    stateless_fragment_default_actions = ["aws:drop"]
    stateless_default_actions          = ["aws:pass"]
    stateful_rule_group_reference {
      resource_arn = aws_networkfirewall_rule_group.test[0].arn
    }
  }
}
`, rName))
}

func testAccAwsNetworkFirewallFirewallPolicy_statelessRuleGroupReference(rName string, priority int) string {
	return composeConfig(
		testAccAwsNetworkFirewallFirewallPolicyStatelessRuleGroupDependencies(rName, 1),
		fmt.Sprintf(`
resource "aws_networkfirewall_firewall_policy" "test" {
  name = %q
  firewall_policy {
    stateless_fragment_default_actions = ["aws:drop"]
    stateless_default_actions          = ["aws:pass"]
    stateless_rule_group_reference {
      priority     = %d
      resource_arn = aws_networkfirewall_rule_group.test[0].arn
    }
  }
}
`, rName, priority))
}

func testAccAwsNetworkFirewallFirewallPolicy_multipleStatelessRuleGroupReferences(rName string) string {
	return composeConfig(
		testAccAwsNetworkFirewallFirewallPolicyStatelessRuleGroupDependencies(rName, 2),
		fmt.Sprintf(`
resource "aws_networkfirewall_firewall_policy" "test" {
  name = %q
  firewall_policy {
    stateless_fragment_default_actions = ["aws:drop"]
    stateless_default_actions          = ["aws:pass"]
    stateless_rule_group_reference {
      priority     = 1
      resource_arn = aws_networkfirewall_rule_group.test[0].arn
    }
    stateless_rule_group_reference {
      priority     = 2
      resource_arn = aws_networkfirewall_rule_group.test[1].arn
    }
  }
}
`, rName))
}

func testAccAwsNetworkFirewallFirewallPolicy_singleStatelessRuleGroupReference(rName string) string {
	return composeConfig(
		testAccAwsNetworkFirewallFirewallPolicyStatelessRuleGroupDependencies(rName, 2),
		fmt.Sprintf(`
resource "aws_networkfirewall_firewall_policy" "test" {
  name = %q
  firewall_policy {
    stateless_fragment_default_actions = ["aws:drop"]
    stateless_default_actions          = ["aws:pass"]
    stateless_rule_group_reference {
      priority     = 1
      resource_arn = aws_networkfirewall_rule_group.test[0].arn
    }
  }
}
`, rName))
}

func testAccAwsNetworkFirewallFirewallPolicy_statelessCustomAction(rName string) string {
	return fmt.Sprintf(`
resource "aws_networkfirewall_firewall_policy" "test" {
  name = %[1]q
  firewall_policy {
    stateless_fragment_default_actions = ["aws:drop"]
    stateless_default_actions          = ["aws:pass"]
    stateless_custom_action {
      action_name = "CustomAction"
      action_definition {
        publish_metric_action {
          dimension {
            value = "example"
          }
        }
      }
    }
  }
}
`, rName)
}

func testAccAwsNetworkFirewallFirewallPolicy_updateStatelessCustomAction(rName string) string {
	return fmt.Sprintf(`
resource "aws_networkfirewall_firewall_policy" "test" {
  name = %[1]q
  firewall_policy {
    stateless_fragment_default_actions = ["aws:drop"]
    stateless_default_actions          = ["aws:pass"]
    stateless_custom_action {
      action_name = "updated"
      action_definition {
        publish_metric_action {
          dimension {
            value = "example-update"
          }
        }
      }
    }
  }
}
`, rName)
}

func testAccAwsNetworkFirewallFirewallPolicy_multipleStatelessCustomActions(rName string) string {
	return fmt.Sprintf(`
resource "aws_networkfirewall_firewall_policy" "test" {
  name = %[1]q
  firewall_policy {
    stateless_fragment_default_actions = ["aws:drop"]
    stateless_default_actions          = ["aws:pass"]
    stateless_custom_action {
      action_definition {
        publish_metric_action {
          dimension {
            value = "example"
          }
        }
      }
      action_name = "CustomAction"
    }
    stateless_custom_action {
      action_definition {
        publish_metric_action {
          dimension {
            value = "example-custom-action"
          }
        }
      }
      action_name = "CustomAction2"
    }
  }
}
`, rName)
}

func testAccAwsNetworkFirewallFirewallPolicy_statefulRuleGroupReferenceAndStatelessCustomAction(rName string) string {
	return composeConfig(
		testAccAwsNetworkFirewallFirewallPolicyStatefulRuleGroupDependencies(rName, 1),
		fmt.Sprintf(`
resource "aws_networkfirewall_firewall_policy" "test" {
  name = %[1]q
  firewall_policy {
    stateless_fragment_default_actions = ["aws:drop"]
    stateless_default_actions          = ["aws:pass"]
    stateful_rule_group_reference {
      resource_arn = aws_networkfirewall_rule_group.test[0].arn
    }
    stateless_custom_action {
      action_definition {
        publish_metric_action {
          dimension {
            value = "example"
          }
        }
      }
      action_name = "CustomAction"
    }
  }
}
`, rName))
}
