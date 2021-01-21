package aws

import (
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/wafv2"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/wafv2/lister"
)

func init() {
	resource.AddTestSweepers("aws_wafv2_rule_group", &resource.Sweeper{
		Name: "aws_wafv2_rule_group",
		F:    testSweepWafv2RuleGroups,
		Dependencies: []string{
			"aws_wafv2_web_acl",
		},
	})
}

func testSweepWafv2RuleGroups(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).wafv2conn

	var sweeperErrs *multierror.Error

	input := &wafv2.ListRuleGroupsInput{
		Scope: aws.String(wafv2.ScopeRegional),
	}

	err = lister.ListRuleGroupsPages(conn, input, func(page *wafv2.ListRuleGroupsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, ruleGroup := range page.RuleGroups {
			id := aws.StringValue(ruleGroup.Id)

			r := resourceAwsWafv2RuleGroup()
			d := r.Data(nil)
			d.SetId(id)
			d.Set("lock_token", ruleGroup.LockToken)
			d.Set("name", ruleGroup.Name)
			d.Set("scope", input.Scope)
			err := r.Delete(d, client)

			if err != nil {
				sweeperErr := fmt.Errorf("error deleting WAFv2 Rule Group (%s): %w", id, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}
		}

		return !lastPage
	})

	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping WAFv2 Rule Group sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error describing WAFv2 Rule Groups: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func TestAccAwsWafv2RuleGroup_basic(t *testing.T) {
	var v wafv2.RuleGroup
	ruleGroupName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_wafv2_rule_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSWafv2ScopeRegional(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsWafv2RuleGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsWafv2RuleGroupConfig_Basic(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2RuleGroupExists(resourceName, &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "capacity", "2"),
					resource.TestCheckResourceAttr(resourceName, "name", ruleGroupName),
					resource.TestCheckResourceAttr(resourceName, "description", ruleGroupName),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "scope", wafv2.ScopeRegional),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.cloudwatch_metrics_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.metric_name", "friendly-metric-name"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.sampled_requests_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccAwsWafv2RuleGroupImportStateIdFunc(resourceName),
			},
		},
	})
}

func TestAccAwsWafv2RuleGroup_updateRule(t *testing.T) {
	var v wafv2.RuleGroup
	ruleGroupName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_wafv2_rule_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSWafv2ScopeRegional(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsWafv2RuleGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsWafv2RuleGroupConfig_Basic(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2RuleGroupExists(resourceName, &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "capacity", "2"),
					resource.TestCheckResourceAttr(resourceName, "name", ruleGroupName),
					resource.TestCheckResourceAttr(resourceName, "description", ruleGroupName),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "scope", wafv2.ScopeRegional),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.cloudwatch_metrics_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.metric_name", "friendly-metric-name"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.sampled_requests_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				Config: testAccAwsWafv2RuleGroupConfig_BasicUpdate(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2RuleGroupExists(resourceName, &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "capacity", "50"),
					resource.TestCheckResourceAttr(resourceName, "name", ruleGroupName),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated"),
					resource.TestCheckResourceAttr(resourceName, "scope", wafv2.ScopeRegional),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.cloudwatch_metrics_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.metric_name", "friendly-metric-name"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.sampled_requests_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"name":                              "rule-1",
						"priority":                          "1",
						"action.#":                          "1",
						"action.0.allow.#":                  "0",
						"action.0.block.#":                  "0",
						"action.0.count.#":                  "1",
						"statement.#":                       "1",
						"statement.0.geo_match_statement.#": "1",
						"statement.0.geo_match_statement.0.country_codes.#": "2",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccAwsWafv2RuleGroupImportStateIdFunc(resourceName),
			},
		},
	})
}

func TestAccAwsWafv2RuleGroup_updateRuleProperties(t *testing.T) {
	var v wafv2.RuleGroup
	ruleGroupName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_wafv2_rule_group.test"
	ruleName2 := fmt.Sprintf("%s-2", ruleGroupName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSWafv2ScopeRegional(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsWafv2RuleGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsWafv2RuleGroupConfig_BasicUpdate(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2RuleGroupExists(resourceName, &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "capacity", "50"),
					resource.TestCheckResourceAttr(resourceName, "name", ruleGroupName),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated"),
					resource.TestCheckResourceAttr(resourceName, "scope", wafv2.ScopeRegional),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.cloudwatch_metrics_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.metric_name", "friendly-metric-name"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.sampled_requests_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"name":             "rule-1",
						"priority":         "1",
						"action.#":         "1",
						"action.0.allow.#": "0",
						"action.0.block.#": "0",
						"action.0.count.#": "1",
						"visibility_config.0.cloudwatch_metrics_enabled": "false",
						"visibility_config.0.metric_name":                "friendly-rule-metric-name",
						"visibility_config.0.sampled_requests_enabled":   "false",
						"statement.#":                                       "1",
						"statement.0.geo_match_statement.#":                 "1",
						"statement.0.geo_match_statement.0.country_codes.#": "2",
					}),
				),
			},
			{
				// Test step verifies addition of a rule block with the first block unchanged
				Config: testAccAwsWafv2RuleGroupConfig_UpdateMultipleRules(ruleGroupName, "rule-1", ruleName2, 1, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2RuleGroupExists(resourceName, &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "capacity", "50"),
					resource.TestCheckResourceAttr(resourceName, "name", ruleGroupName),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated"),
					resource.TestCheckResourceAttr(resourceName, "scope", wafv2.ScopeRegional),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.cloudwatch_metrics_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.metric_name", "friendly-metric-name"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.sampled_requests_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"name":                "rule-1",
						"priority":            "1",
						"action.#":            "1",
						"action.0.allow.#":    "0",
						"action.0.block.#":    "0",
						"action.0.count.#":    "1",
						"visibility_config.#": "1",
						"visibility_config.0.cloudwatch_metrics_enabled": "false",
						"visibility_config.0.metric_name":                "rule-1",
						"visibility_config.0.sampled_requests_enabled":   "false",
						"statement.#":                                       "1",
						"statement.0.geo_match_statement.#":                 "1",
						"statement.0.geo_match_statement.0.country_codes.#": "2",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"name":                ruleName2,
						"priority":            "2",
						"action.#":            "1",
						"action.0.allow.#":    "0",
						"action.0.block.#":    "1",
						"action.0.count.#":    "0",
						"visibility_config.#": "1",
						"visibility_config.0.cloudwatch_metrics_enabled": "false",
						"visibility_config.0.metric_name":                ruleName2,
						"visibility_config.0.sampled_requests_enabled":   "false",
						"statement.#": "1",
						"statement.0.size_constraint_statement.#":                                 "1",
						"statement.0.size_constraint_statement.0.comparison_operator":             "LT",
						"statement.0.size_constraint_statement.0.field_to_match.#":                "1",
						"statement.0.size_constraint_statement.0.field_to_match.0.query_string.#": "1",
						"statement.0.size_constraint_statement.0.size":                            "50",
						"statement.0.size_constraint_statement.0.text_transformation.#":           "2",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*.statement.0.size_constraint_statement.0.text_transformation.*", map[string]string{
						"priority": "2",
						"type":     "CMD_LINE",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*.statement.0.size_constraint_statement.0.text_transformation.*", map[string]string{
						"priority": "5",
						"type":     "NONE",
					}),
				),
			},
			{
				// Test step to verify a change in priority for rule #1 and a change in name and priority for rule #2
				Config: testAccAwsWafv2RuleGroupConfig_UpdateMultipleRules(ruleGroupName, "rule-1", "updated", 5, 10),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2RuleGroupExists(resourceName, &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "capacity", "50"),
					resource.TestCheckResourceAttr(resourceName, "name", ruleGroupName),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated"),
					resource.TestCheckResourceAttr(resourceName, "scope", wafv2.ScopeRegional),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.cloudwatch_metrics_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.metric_name", "friendly-metric-name"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.sampled_requests_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"name":                "rule-1",
						"priority":            "5",
						"action.#":            "1",
						"action.0.allow.#":    "0",
						"action.0.block.#":    "0",
						"action.0.count.#":    "1",
						"visibility_config.#": "1",
						"visibility_config.0.cloudwatch_metrics_enabled": "false",
						"visibility_config.0.metric_name":                "rule-1",
						"visibility_config.0.sampled_requests_enabled":   "false",
						"statement.#":                                       "1",
						"statement.0.geo_match_statement.#":                 "1",
						"statement.0.geo_match_statement.0.country_codes.#": "2",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"name":                "updated",
						"priority":            "10",
						"action.#":            "1",
						"action.0.allow.#":    "0",
						"action.0.block.#":    "1",
						"action.0.count.#":    "0",
						"visibility_config.#": "1",
						"visibility_config.0.cloudwatch_metrics_enabled": "false",
						"visibility_config.0.metric_name":                "updated",
						"visibility_config.0.sampled_requests_enabled":   "false",
						"statement.#": "1",
						"statement.0.size_constraint_statement.#":                                 "1",
						"statement.0.size_constraint_statement.0.comparison_operator":             "LT",
						"statement.0.size_constraint_statement.0.field_to_match.#":                "1",
						"statement.0.size_constraint_statement.0.field_to_match.0.query_string.#": "1",
						"statement.0.size_constraint_statement.0.size":                            "50",
						"statement.0.size_constraint_statement.0.text_transformation.#":           "2",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*.statement.0.size_constraint_statement.0.text_transformation.*", map[string]string{
						"priority": "2",
						"type":     "CMD_LINE",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*.statement.0.size_constraint_statement.0.text_transformation.*", map[string]string{
						"priority": "5",
						"type":     "NONE",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccAwsWafv2RuleGroupImportStateIdFunc(resourceName),
			},
		},
	})
}

func TestAccAwsWafv2RuleGroup_ByteMatchStatement(t *testing.T) {
	var v wafv2.RuleGroup
	ruleGroupName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_wafv2_rule_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSWafv2ScopeRegional(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsWafv2RuleGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsWafv2RuleGroupConfig_ByteMatchStatement(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2RuleGroupExists(resourceName, &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.#":                        "1",
						"statement.0.byte_match_statement.#": "1",
						"statement.0.byte_match_statement.0.positional_constraint": "CONTAINS",
						"statement.0.byte_match_statement.0.search_string":         "word",
						"statement.0.byte_match_statement.0.text_transformation.#": "2",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*.statement.0.byte_match_statement.0.text_transformation.*", map[string]string{
						"priority": "5",
						"type":     "NONE",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*.statement.0.byte_match_statement.0.text_transformation.*", map[string]string{
						"priority": "2",
						"type":     "LOWERCASE",
					}),
				),
			},
			{
				Config: testAccAwsWafv2RuleGroupConfig_ByteMatchStatement_Update(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2RuleGroupExists(resourceName, &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.#":                        "1",
						"statement.0.byte_match_statement.#": "1",
						"statement.0.byte_match_statement.0.positional_constraint": "EXACTLY",
						"statement.0.byte_match_statement.0.search_string":         "sentence",
						"statement.0.byte_match_statement.0.text_transformation.#": "1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*.statement.0.byte_match_statement.0.text_transformation.*", map[string]string{
						"priority": "3",
						"type":     "CMD_LINE",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccAwsWafv2RuleGroupImportStateIdFunc(resourceName),
			},
		},
	})
}

func TestAccAwsWafv2RuleGroup_ByteMatchStatement_FieldToMatch(t *testing.T) {
	var v wafv2.RuleGroup
	ruleGroupName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_wafv2_rule_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSWafv2ScopeRegional(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsWafv2RuleGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsWafv2RuleGroupConfig_ByteMatchStatement_FieldToMatchAllQueryArguments(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2RuleGroupExists(resourceName, &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.#":                                         "1",
						"statement.0.byte_match_statement.#":                  "1",
						"statement.0.byte_match_statement.0.field_to_match.#": "1",
						"statement.0.byte_match_statement.0.field_to_match.0.all_query_arguments.#":   "1",
						"statement.0.byte_match_statement.0.field_to_match.0.body.#":                  "0",
						"statement.0.byte_match_statement.0.field_to_match.0.method.#":                "0",
						"statement.0.byte_match_statement.0.field_to_match.0.query_string.#":          "0",
						"statement.0.byte_match_statement.0.field_to_match.0.single_header.#":         "0",
						"statement.0.byte_match_statement.0.field_to_match.0.single_query_argument.#": "0",
						"statement.0.byte_match_statement.0.field_to_match.0.uri_path.#":              "0",
					}),
				),
			},
			{
				Config: testAccAwsWafv2RuleGroupConfig_ByteMatchStatement_FieldToMatchBody(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2RuleGroupExists(resourceName, &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.#":                                         "1",
						"statement.0.byte_match_statement.#":                  "1",
						"statement.0.byte_match_statement.0.field_to_match.#": "1",
						"statement.0.byte_match_statement.0.field_to_match.0.all_query_arguments.#":   "0",
						"statement.0.byte_match_statement.0.field_to_match.0.body.#":                  "1",
						"statement.0.byte_match_statement.0.field_to_match.0.method.#":                "0",
						"statement.0.byte_match_statement.0.field_to_match.0.query_string.#":          "0",
						"statement.0.byte_match_statement.0.field_to_match.0.single_header.#":         "0",
						"statement.0.byte_match_statement.0.field_to_match.0.single_query_argument.#": "0",
						"statement.0.byte_match_statement.0.field_to_match.0.uri_path.#":              "0",
					}),
				),
			},
			{
				Config: testAccAwsWafv2RuleGroupConfig_ByteMatchStatement_FieldToMatchMethod(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2RuleGroupExists(resourceName, &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.#":                                         "1",
						"statement.0.byte_match_statement.#":                  "1",
						"statement.0.byte_match_statement.0.field_to_match.#": "1",
						"statement.0.byte_match_statement.0.field_to_match.0.all_query_arguments.#":   "0",
						"statement.0.byte_match_statement.0.field_to_match.0.body.#":                  "0",
						"statement.0.byte_match_statement.0.field_to_match.0.method.#":                "1",
						"statement.0.byte_match_statement.0.field_to_match.0.query_string.#":          "0",
						"statement.0.byte_match_statement.0.field_to_match.0.single_header.#":         "0",
						"statement.0.byte_match_statement.0.field_to_match.0.single_query_argument.#": "0",
						"statement.0.byte_match_statement.0.field_to_match.0.uri_path.#":              "0",
					}),
				),
			},
			{
				Config: testAccAwsWafv2RuleGroupConfig_ByteMatchStatement_FieldToMatchQueryString(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2RuleGroupExists(resourceName, &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.#":                                         "1",
						"statement.0.byte_match_statement.#":                  "1",
						"statement.0.byte_match_statement.0.field_to_match.#": "1",
						"statement.0.byte_match_statement.0.field_to_match.0.all_query_arguments.#":   "0",
						"statement.0.byte_match_statement.0.field_to_match.0.body.#":                  "0",
						"statement.0.byte_match_statement.0.field_to_match.0.method.#":                "0",
						"statement.0.byte_match_statement.0.field_to_match.0.query_string.#":          "1",
						"statement.0.byte_match_statement.0.field_to_match.0.single_header.#":         "0",
						"statement.0.byte_match_statement.0.field_to_match.0.single_query_argument.#": "0",
						"statement.0.byte_match_statement.0.field_to_match.0.uri_path.#":              "0",
					}),
				),
			},
			{
				Config: testAccAwsWafv2RuleGroupConfig_ByteMatchStatement_FieldToMatchSingleHeader(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2RuleGroupExists(resourceName, &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.#":                                         "1",
						"statement.0.byte_match_statement.#":                  "1",
						"statement.0.byte_match_statement.0.field_to_match.#": "1",
						"statement.0.byte_match_statement.0.field_to_match.0.all_query_arguments.#":   "0",
						"statement.0.byte_match_statement.0.field_to_match.0.body.#":                  "0",
						"statement.0.byte_match_statement.0.field_to_match.0.method.#":                "0",
						"statement.0.byte_match_statement.0.field_to_match.0.query_string.#":          "0",
						"statement.0.byte_match_statement.0.field_to_match.0.single_header.#":         "1",
						"statement.0.byte_match_statement.0.field_to_match.0.single_header.0.name":    "a-forty-character-long-header-name-40-40",
						"statement.0.byte_match_statement.0.field_to_match.0.single_query_argument.#": "0",
						"statement.0.byte_match_statement.0.field_to_match.0.uri_path.#":              "0",
					}),
				),
			},
			{
				Config: testAccAwsWafv2RuleGroupConfig_ByteMatchStatement_FieldToMatchSingleQueryArgument(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2RuleGroupExists(resourceName, &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.#":                                         "1",
						"statement.0.byte_match_statement.#":                  "1",
						"statement.0.byte_match_statement.0.field_to_match.#": "1",
						"statement.0.byte_match_statement.0.field_to_match.0.all_query_arguments.#":        "0",
						"statement.0.byte_match_statement.0.field_to_match.0.body.#":                       "0",
						"statement.0.byte_match_statement.0.field_to_match.0.method.#":                     "0",
						"statement.0.byte_match_statement.0.field_to_match.0.query_string.#":               "0",
						"statement.0.byte_match_statement.0.field_to_match.0.single_header.#":              "0",
						"statement.0.byte_match_statement.0.field_to_match.0.single_query_argument.#":      "1",
						"statement.0.byte_match_statement.0.field_to_match.0.single_query_argument.0.name": "a-max-30-characters-long-name-",
						"statement.0.byte_match_statement.0.field_to_match.0.uri_path.#":                   "0",
					}),
				),
			},
			{
				Config: testAccAwsWafv2RuleGroupConfig_ByteMatchStatement_FieldToMatchUriPath(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2RuleGroupExists(resourceName, &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.#":                                         "1",
						"statement.0.byte_match_statement.#":                  "1",
						"statement.0.byte_match_statement.0.field_to_match.#": "1",
						"statement.0.byte_match_statement.0.field_to_match.0.all_query_arguments.#":   "0",
						"statement.0.byte_match_statement.0.field_to_match.0.body.#":                  "0",
						"statement.0.byte_match_statement.0.field_to_match.0.method.#":                "0",
						"statement.0.byte_match_statement.0.field_to_match.0.query_string.#":          "0",
						"statement.0.byte_match_statement.0.field_to_match.0.single_header.#":         "0",
						"statement.0.byte_match_statement.0.field_to_match.0.single_query_argument.#": "0",
						"statement.0.byte_match_statement.0.field_to_match.0.uri_path.#":              "1",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccAwsWafv2RuleGroupImportStateIdFunc(resourceName),
			},
		},
	})
}

func TestAccAwsWafv2RuleGroup_ChangeNameForceNew(t *testing.T) {
	var before, after wafv2.RuleGroup
	ruleGroupName := acctest.RandomWithPrefix("tf-acc-test")
	ruleGroupNewName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_wafv2_rule_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSWafv2ScopeRegional(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsWafv2RuleGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsWafv2RuleGroupConfig_Basic(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2RuleGroupExists(resourceName, &before),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "capacity", "2"),
					resource.TestCheckResourceAttr(resourceName, "name", ruleGroupName),
					resource.TestCheckResourceAttr(resourceName, "description", ruleGroupName),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "scope", wafv2.ScopeRegional),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.cloudwatch_metrics_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.metric_name", "friendly-metric-name"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.sampled_requests_enabled", "false"),
				),
			},
			{
				Config: testAccAwsWafv2RuleGroupConfig_Basic(ruleGroupNewName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2RuleGroupExists(resourceName, &after),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "capacity", "2"),
					resource.TestCheckResourceAttr(resourceName, "name", ruleGroupNewName),
					resource.TestCheckResourceAttr(resourceName, "description", ruleGroupNewName),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "scope", wafv2.ScopeRegional),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.cloudwatch_metrics_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.metric_name", "friendly-metric-name"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.sampled_requests_enabled", "false"),
				),
			},
		},
	})
}

func TestAccAwsWafv2RuleGroup_ChangeCapacityForceNew(t *testing.T) {
	var before, after wafv2.RuleGroup
	ruleGroupName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_wafv2_rule_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSWafv2ScopeRegional(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsWafv2RuleGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsWafv2RuleGroupConfig_Basic(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2RuleGroupExists(resourceName, &before),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "capacity", "2"),
					resource.TestCheckResourceAttr(resourceName, "name", ruleGroupName),
					resource.TestCheckResourceAttr(resourceName, "description", ruleGroupName),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "scope", wafv2.ScopeRegional),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.cloudwatch_metrics_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.metric_name", "friendly-metric-name"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.sampled_requests_enabled", "false"),
				),
			},
			{
				Config: testAccAwsWafv2RuleGroupConfig_UpdateCapacity(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2RuleGroupExists(resourceName, &after),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "capacity", "3"),
					resource.TestCheckResourceAttr(resourceName, "name", ruleGroupName),
					resource.TestCheckResourceAttr(resourceName, "description", ruleGroupName),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "scope", wafv2.ScopeRegional),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.cloudwatch_metrics_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.metric_name", "friendly-metric-name"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.sampled_requests_enabled", "false"),
				),
			},
		},
	})
}

func TestAccAwsWafv2RuleGroup_ChangeMetricNameForceNew(t *testing.T) {
	var before, after wafv2.RuleGroup
	ruleGroupName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_wafv2_rule_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSWafv2ScopeRegional(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsWafv2RuleGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsWafv2RuleGroupConfig_Basic(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2RuleGroupExists(resourceName, &before),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "capacity", "2"),
					resource.TestCheckResourceAttr(resourceName, "name", ruleGroupName),
					resource.TestCheckResourceAttr(resourceName, "description", ruleGroupName),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "scope", wafv2.ScopeRegional),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.cloudwatch_metrics_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.metric_name", "friendly-metric-name"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.sampled_requests_enabled", "false"),
				),
			},
			{
				Config: testAccAwsWafv2RuleGroupConfig_UpdateMetricName(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2RuleGroupExists(resourceName, &after),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "capacity", "2"),
					resource.TestCheckResourceAttr(resourceName, "name", ruleGroupName),
					resource.TestCheckResourceAttr(resourceName, "description", ruleGroupName),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "scope", wafv2.ScopeRegional),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.cloudwatch_metrics_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.metric_name", "updated-friendly-metric-name"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.sampled_requests_enabled", "false"),
				),
			},
		},
	})
}

func TestAccAwsWafv2RuleGroup_Disappears(t *testing.T) {
	var v wafv2.RuleGroup
	ruleGroupName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_wafv2_rule_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSWafv2ScopeRegional(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsWafv2RuleGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsWafv2RuleGroupConfig_Minimal(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2RuleGroupExists(resourceName, &v),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsWafv2RuleGroup(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAwsWafv2RuleGroup_GeoMatchStatement(t *testing.T) {
	var v wafv2.RuleGroup
	ruleGroupName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_wafv2_rule_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSWafv2ScopeRegional(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsWafv2RuleGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsWafv2RuleGroupConfig_GeoMatchStatement(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2RuleGroupExists(resourceName, &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.#":                                             "1",
						"statement.0.geo_match_statement.#":                       "1",
						"statement.0.geo_match_statement.0.country_codes.#":       "2",
						"statement.0.geo_match_statement.0.country_codes.0":       "US",
						"statement.0.geo_match_statement.0.country_codes.1":       "NL",
						"statement.0.geo_match_statement.0.forwarded_ip_config.#": "0",
					}),
				),
			},
			{
				Config: testAccAwsWafv2RuleGroupConfig_GeoMatchStatement_Update(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2RuleGroupExists(resourceName, &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.#":                                             "1",
						"statement.0.geo_match_statement.#":                       "1",
						"statement.0.geo_match_statement.0.country_codes.#":       "3",
						"statement.0.geo_match_statement.0.country_codes.0":       "ZM",
						"statement.0.geo_match_statement.0.country_codes.1":       "EE",
						"statement.0.geo_match_statement.0.country_codes.2":       "MM",
						"statement.0.geo_match_statement.0.forwarded_ip_config.#": "0",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccAwsWafv2RuleGroupImportStateIdFunc(resourceName),
			},
		},
	})
}

func TestAccAwsWafv2RuleGroup_GeoMatchStatement_ForwardedIPConfig(t *testing.T) {
	var v wafv2.RuleGroup
	ruleGroupName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_wafv2_rule_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSWafv2ScopeRegional(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsWafv2RuleGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsWafv2RuleGroupConfig_GeoMatchStatement_ForwardedIPConfig(ruleGroupName, "MATCH", "X-Forwarded-For"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2RuleGroupExists(resourceName, &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.#":                                                               "1",
						"statement.0.geo_match_statement.#":                                         "1",
						"statement.0.geo_match_statement.0.country_codes.#":                         "2",
						"statement.0.geo_match_statement.0.country_codes.0":                         "US",
						"statement.0.geo_match_statement.0.country_codes.1":                         "NL",
						"statement.0.geo_match_statement.0.forwarded_ip_config.#":                   "1",
						"statement.0.geo_match_statement.0.forwarded_ip_config.0.fallback_behavior": "MATCH",
						"statement.0.geo_match_statement.0.forwarded_ip_config.0.header_name":       "X-Forwarded-For",
					}),
				),
			},
			{
				Config: testAccAwsWafv2RuleGroupConfig_GeoMatchStatement_ForwardedIPConfig(ruleGroupName, "NO_MATCH", "Updated"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2RuleGroupExists(resourceName, &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.#":                                                               "1",
						"statement.0.geo_match_statement.#":                                         "1",
						"statement.0.geo_match_statement.0.country_codes.#":                         "2",
						"statement.0.geo_match_statement.0.country_codes.0":                         "US",
						"statement.0.geo_match_statement.0.country_codes.1":                         "NL",
						"statement.0.geo_match_statement.0.forwarded_ip_config.#":                   "1",
						"statement.0.geo_match_statement.0.forwarded_ip_config.0.fallback_behavior": "NO_MATCH",
						"statement.0.geo_match_statement.0.forwarded_ip_config.0.header_name":       "Updated",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccAwsWafv2RuleGroupImportStateIdFunc(resourceName),
			},
		},
	})
}

func TestAccAwsWafv2RuleGroup_IpSetReferenceStatement(t *testing.T) {
	var v wafv2.RuleGroup
	ruleGroupName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_wafv2_rule_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSWafv2ScopeRegional(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsWafv2RuleGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsWafv2RuleGroupConfig_IpSetReferenceStatement(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2RuleGroupExists(resourceName, &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.#": "1",
						"statement.0.ip_set_reference_statement.#":                              "1",
						"statement.0.ip_set_reference_statement.0.ip_set_forwarded_ip_config.#": "0",
					}),
					resource.TestMatchTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]*regexp.Regexp{
						"statement.0.ip_set_reference_statement.0.arn": regexp.MustCompile(`regional/ipset/.+$`),
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccAwsWafv2RuleGroupImportStateIdFunc(resourceName),
			},
		},
	})
}

func TestAccAwsWafv2RuleGroup_IpSetReferenceStatement_IPSetForwardedIPConfig(t *testing.T) {
	var v wafv2.RuleGroup
	ruleGroupName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_wafv2_rule_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSWafv2ScopeRegional(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsWafv2RuleGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsWafv2RuleGroupConfig_IpSetReferenceStatement_IPSetForwardedIPConfig(ruleGroupName, "MATCH", "X-Forwarded-For", "FIRST"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2RuleGroupExists(resourceName, &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.#": "1",
						"statement.0.ip_set_reference_statement.#": "1",
					}),
					resource.TestMatchTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]*regexp.Regexp{
						"statement.0.ip_set_reference_statement.0.arn": regexp.MustCompile(`regional/ipset/.+$`),
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.0.ip_set_reference_statement.0.ip_set_forwarded_ip_config.#":                   "1",
						"statement.0.ip_set_reference_statement.0.ip_set_forwarded_ip_config.0.fallback_behavior": "MATCH",
						"statement.0.ip_set_reference_statement.0.ip_set_forwarded_ip_config.0.header_name":       "X-Forwarded-For",
						"statement.0.ip_set_reference_statement.0.ip_set_forwarded_ip_config.0.position":          "FIRST",
					}),
				),
			},
			{
				Config: testAccAwsWafv2RuleGroupConfig_IpSetReferenceStatement_IPSetForwardedIPConfig(ruleGroupName, "NO_MATCH", "X-Forwarded-For", "LAST"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2RuleGroupExists(resourceName, &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.#": "1",
						"statement.0.ip_set_reference_statement.#": "1",
					}),
					resource.TestMatchTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]*regexp.Regexp{
						"statement.0.ip_set_reference_statement.0.arn": regexp.MustCompile(`regional/ipset/.+$`),
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.0.ip_set_reference_statement.0.ip_set_forwarded_ip_config.#":                   "1",
						"statement.0.ip_set_reference_statement.0.ip_set_forwarded_ip_config.0.fallback_behavior": "NO_MATCH",
						"statement.0.ip_set_reference_statement.0.ip_set_forwarded_ip_config.0.header_name":       "X-Forwarded-For",
						"statement.0.ip_set_reference_statement.0.ip_set_forwarded_ip_config.0.position":          "LAST",
					}),
				),
			},
			{
				Config: testAccAwsWafv2RuleGroupConfig_IpSetReferenceStatement_IPSetForwardedIPConfig(ruleGroupName, "MATCH", "Updated", "ANY"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2RuleGroupExists(resourceName, &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.#": "1",
						"statement.0.ip_set_reference_statement.#": "1",
					}),
					resource.TestMatchTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]*regexp.Regexp{
						"statement.0.ip_set_reference_statement.0.arn": regexp.MustCompile(`regional/ipset/.+$`),
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.0.ip_set_reference_statement.0.ip_set_forwarded_ip_config.#":                   "1",
						"statement.0.ip_set_reference_statement.0.ip_set_forwarded_ip_config.0.fallback_behavior": "MATCH",
						"statement.0.ip_set_reference_statement.0.ip_set_forwarded_ip_config.0.header_name":       "Updated",
						"statement.0.ip_set_reference_statement.0.ip_set_forwarded_ip_config.0.position":          "ANY",
					}),
				),
			},
			{
				Config: testAccAwsWafv2RuleGroupConfig_IpSetReferenceStatement(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2RuleGroupExists(resourceName, &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.#": "1",
						"statement.0.ip_set_reference_statement.#":                              "1",
						"statement.0.ip_set_reference_statement.0.ip_set_forwarded_ip_config.#": "0",
					}),
					resource.TestMatchTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]*regexp.Regexp{
						"statement.0.ip_set_reference_statement.0.arn": regexp.MustCompile(`regional/ipset/.+$`),
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccAwsWafv2RuleGroupImportStateIdFunc(resourceName),
			},
		},
	})
}

func TestAccAwsWafv2RuleGroup_LogicalRuleStatements(t *testing.T) {
	var v wafv2.RuleGroup
	ruleGroupName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_wafv2_rule_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSWafv2ScopeRegional(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsWafv2RuleGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsWafv2RuleGroupConfig_LogicalRuleStatement_And(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2RuleGroupExists(resourceName, &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.#":                             "1",
						"statement.0.and_statement.#":             "1",
						"statement.0.and_statement.0.statement.#": "2",
						"statement.0.and_statement.0.statement.0.geo_match_statement.#": "1",
						"statement.0.and_statement.0.statement.1.geo_match_statement.#": "1",
					}),
				),
			},
			{
				Config: testAccAwsWafv2RuleGroupConfig_LogicalRuleStatement_NotAnd(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2RuleGroupExists(resourceName, &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.#":                                                         "1",
						"statement.0.not_statement.#":                                         "1",
						"statement.0.not_statement.0.statement.#":                             "1",
						"statement.0.not_statement.0.statement.0.and_statement.#":             "1",
						"statement.0.not_statement.0.statement.0.and_statement.0.statement.#": "2",
						"statement.0.not_statement.0.statement.0.and_statement.0.statement.0.geo_match_statement.#": "1",
						"statement.0.not_statement.0.statement.0.and_statement.0.statement.1.geo_match_statement.#": "1",
					}),
				),
			},
			{
				Config: testAccAwsWafv2RuleGroupConfig_LogicalRuleStatement_OrNotAnd(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2RuleGroupExists(resourceName, &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.#":                                                        "1",
						"statement.0.or_statement.#":                                         "1",
						"statement.0.or_statement.0.statement.#":                             "2",
						"statement.0.or_statement.0.statement.0.not_statement.#":             "1",
						"statement.0.or_statement.0.statement.0.not_statement.0.statement.#": "1",
						"statement.0.or_statement.0.statement.0.not_statement.0.statement.0.geo_match_statement.#": "1",
						"statement.0.or_statement.0.statement.1.and_statement.#":                                   "1",
						"statement.0.or_statement.0.statement.1.and_statement.0.statement.#":                       "2",
						"statement.0.or_statement.0.statement.1.and_statement.0.statement.0.geo_match_statement.#": "1",
						"statement.0.or_statement.0.statement.1.and_statement.0.statement.1.geo_match_statement.#": "1",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccAwsWafv2RuleGroupImportStateIdFunc(resourceName),
			},
		},
	})
}

func TestAccAwsWafv2RuleGroup_Minimal(t *testing.T) {
	var v wafv2.RuleGroup
	ruleGroupName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_wafv2_rule_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSWafv2ScopeRegional(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsWafv2RuleGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsWafv2RuleGroupConfig_Minimal(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2RuleGroupExists(resourceName, &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "capacity", "2"),
					resource.TestCheckResourceAttr(resourceName, "name", ruleGroupName),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "scope", wafv2.ScopeRegional),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.cloudwatch_metrics_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.metric_name", "friendly-metric-name"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.sampled_requests_enabled", "false"),
				),
			},
		},
	})
}

func TestAccAwsWafv2RuleGroup_RegexPatternSetReferenceStatement(t *testing.T) {
	var v wafv2.RuleGroup
	ruleGroupName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_wafv2_rule_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSWafv2ScopeRegional(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsWafv2RuleGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsWafv2RuleGroupConfig_RegexPatternSetReferenceStatement(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2RuleGroupExists(resourceName, &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.#": "1",
						"statement.0.regex_pattern_set_reference_statement.#":                       "1",
						"statement.0.regex_pattern_set_reference_statement.0.field_to_match.#":      "1",
						"statement.0.regex_pattern_set_reference_statement.0.text_transformation.#": "1",
					}),
					resource.TestMatchTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]*regexp.Regexp{
						"statement.0.regex_pattern_set_reference_statement.0.arn": regexp.MustCompile(`regional/regexpatternset/.+$`),
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccAwsWafv2RuleGroupImportStateIdFunc(resourceName),
			},
		},
	})
}

func TestAccAwsWafv2RuleGroup_RuleAction(t *testing.T) {
	var v wafv2.RuleGroup
	ruleGroupName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_wafv2_rule_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSWafv2ScopeRegional(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsWafv2RuleGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsWafv2RuleGroupConfig_RuleActionAllow(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2RuleGroupExists(resourceName, &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "capacity", "2"),
					resource.TestCheckResourceAttr(resourceName, "name", ruleGroupName),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "scope", wafv2.ScopeRegional),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"action.#":         "1",
						"action.0.allow.#": "1",
						"action.0.block.#": "0",
						"action.0.count.#": "0",
					}),
				),
			},
			{
				Config: testAccAwsWafv2RuleGroupConfig_RuleActionBlock(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2RuleGroupExists(resourceName, &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "capacity", "2"),
					resource.TestCheckResourceAttr(resourceName, "name", ruleGroupName),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "scope", wafv2.ScopeRegional),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"action.#":         "1",
						"action.0.allow.#": "0",
						"action.0.block.#": "1",
						"action.0.count.#": "0",
					}),
				),
			},
			{
				Config: testAccAwsWafv2RuleGroupConfig_RuleActionCount(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2RuleGroupExists(resourceName, &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "capacity", "2"),
					resource.TestCheckResourceAttr(resourceName, "name", ruleGroupName),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "scope", wafv2.ScopeRegional),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"action.#":         "1",
						"action.0.allow.#": "0",
						"action.0.block.#": "0",
						"action.0.count.#": "1",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccAwsWafv2RuleGroupImportStateIdFunc(resourceName),
			},
		},
	})
}
func TestAccAwsWafv2RuleGroup_SizeConstraintStatement(t *testing.T) {
	var v wafv2.RuleGroup
	ruleGroupName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_wafv2_rule_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSWafv2ScopeRegional(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsWafv2RuleGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsWafv2RuleGroupConfig_SizeConstraintStatement(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2RuleGroupExists(resourceName, &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.#": "1",
						"statement.0.size_constraint_statement.#":                           "1",
						"statement.0.size_constraint_statement.0.comparison_operator":       "GT",
						"statement.0.size_constraint_statement.0.size":                      "100",
						"statement.0.size_constraint_statement.0.field_to_match.#":          "1",
						"statement.0.size_constraint_statement.0.field_to_match.0.method.#": "1",
						"statement.0.size_constraint_statement.0.text_transformation.#":     "1",
					}),
				),
			},
			{
				Config: testAccAwsWafv2RuleGroupConfig_SizeConstraintStatement_Update(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2RuleGroupExists(resourceName, &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.#": "1",
						"statement.0.size_constraint_statement.#":                                 "1",
						"statement.0.size_constraint_statement.0.comparison_operator":             "LT",
						"statement.0.size_constraint_statement.0.size":                            "50",
						"statement.0.size_constraint_statement.0.field_to_match.#":                "1",
						"statement.0.size_constraint_statement.0.field_to_match.0.query_string.#": "1",
						"statement.0.size_constraint_statement.0.text_transformation.#":           "2",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccAwsWafv2RuleGroupImportStateIdFunc(resourceName),
			},
		},
	})
}
func TestAccAwsWafv2RuleGroup_SqliMatchStatement(t *testing.T) {
	var v wafv2.RuleGroup
	ruleGroupName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_wafv2_rule_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSWafv2ScopeRegional(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsWafv2RuleGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsWafv2RuleGroupConfig_SqliMatchStatement(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2RuleGroupExists(resourceName, &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.#":                                         "1",
						"statement.0.sqli_match_statement.#":                  "1",
						"statement.0.sqli_match_statement.0.field_to_match.#": "1",
						"statement.0.sqli_match_statement.0.field_to_match.0.all_query_arguments.#": "1",
						"statement.0.sqli_match_statement.0.text_transformation.#":                  "2",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*.statement.0.sqli_match_statement.0.text_transformation.*", map[string]string{
						"priority": "5",
						"type":     "URL_DECODE",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*.statement.0.sqli_match_statement.0.text_transformation.*", map[string]string{
						"priority": "2",
						"type":     "LOWERCASE",
					}),
				),
			},
			{
				Config: testAccAwsWafv2RuleGroupConfig_SqliMatchStatement_Update(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2RuleGroupExists(resourceName, &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.#":                                                "1",
						"statement.0.sqli_match_statement.#":                         "1",
						"statement.0.sqli_match_statement.0.field_to_match.#":        "1",
						"statement.0.sqli_match_statement.0.field_to_match.0.body.#": "1",
						"statement.0.sqli_match_statement.0.text_transformation.#":   "3",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*.statement.0.sqli_match_statement.0.text_transformation.*", map[string]string{
						"priority": "5",
						"type":     "URL_DECODE",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*.statement.0.sqli_match_statement.0.text_transformation.*", map[string]string{
						"priority": "4",
						"type":     "HTML_ENTITY_DECODE",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*.statement.0.sqli_match_statement.0.text_transformation.*", map[string]string{
						"priority": "3",
						"type":     "COMPRESS_WHITE_SPACE",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccAwsWafv2RuleGroupImportStateIdFunc(resourceName),
			},
		},
	})
}

func TestAccAwsWafv2RuleGroup_Tags(t *testing.T) {
	var v wafv2.RuleGroup
	ruleGroupName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_wafv2_rule_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSWafv2ScopeRegional(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsWafv2RuleGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsWafv2RuleGroupConfig_OneTag(ruleGroupName, "Tag1", "Value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2RuleGroupExists(resourceName, &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Tag1", "Value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccAwsWafv2RuleGroupImportStateIdFunc(resourceName),
			},
			{
				Config: testAccAwsWafv2RuleGroupConfig_TwoTags(ruleGroupName, "Tag1", "Value1Updated", "Tag2", "Value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2RuleGroupExists(resourceName, &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.Tag1", "Value1Updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.Tag2", "Value2"),
				),
			},
			{
				Config: testAccAwsWafv2RuleGroupConfig_OneTag(ruleGroupName, "Tag2", "Value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2RuleGroupExists(resourceName, &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Tag2", "Value2"),
				),
			},
		},
	})
}

func TestAccAwsWafv2RuleGroup_XssMatchStatement(t *testing.T) {
	var v wafv2.RuleGroup
	ruleGroupName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_wafv2_rule_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSWafv2ScopeRegional(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsWafv2RuleGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsWafv2RuleGroupConfig_XssMatchStatement(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2RuleGroupExists(resourceName, &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.#":                                               "1",
						"statement.0.xss_match_statement.#":                         "1",
						"statement.0.xss_match_statement.0.field_to_match.#":        "1",
						"statement.0.xss_match_statement.0.field_to_match.0.body.#": "1",
						"statement.0.xss_match_statement.0.text_transformation.#":   "1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*.statement.0.xss_match_statement.0.text_transformation.*", map[string]string{
						"priority": "2",
						"type":     "NONE",
					}),
				),
			},
			{
				Config: testAccAwsWafv2RuleGroupConfig_XssMatchStatement_Update(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2RuleGroupExists(resourceName, &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.#":                                               "1",
						"statement.0.xss_match_statement.#":                         "1",
						"statement.0.xss_match_statement.0.field_to_match.#":        "1",
						"statement.0.xss_match_statement.0.field_to_match.0.body.#": "1",
						"statement.0.xss_match_statement.0.text_transformation.#":   "1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*.statement.0.xss_match_statement.0.text_transformation.*", map[string]string{
						"priority": "2",
						"type":     "URL_DECODE",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccAwsWafv2RuleGroupImportStateIdFunc(resourceName),
			},
		},
	})
}

func testAccPreCheckAWSWafv2ScopeRegional(t *testing.T) {
	conn := testAccProvider.Meta().(*AWSClient).wafv2conn

	input := &wafv2.ListRuleGroupsInput{
		Scope: aws.String(wafv2.ScopeRegional),
	}

	_, err := conn.ListRuleGroups(input)

	if testAccPreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccCheckAwsWafv2RuleGroupDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_wafv2_rule_group" {
			continue
		}

		conn := testAccProvider.Meta().(*AWSClient).wafv2conn
		resp, err := conn.GetRuleGroup(
			&wafv2.GetRuleGroupInput{
				Id:    aws.String(rs.Primary.ID),
				Name:  aws.String(rs.Primary.Attributes["name"]),
				Scope: aws.String(rs.Primary.Attributes["scope"]),
			})

		if err == nil {
			if resp == nil || resp.RuleGroup == nil {
				return fmt.Errorf("Error getting WAFv2 RuleGroup")
			}

			if aws.StringValue(resp.RuleGroup.Id) == rs.Primary.ID {
				return fmt.Errorf("WAFv2 RuleGroup %s still exists", rs.Primary.ID)
			}

			return nil
		}

		// Return nil if the RuleGroup is already destroyed
		if isAWSErr(err, wafv2.ErrCodeWAFNonexistentItemException, "") {
			return nil
		}

		return err
	}

	return nil
}

func testAccCheckAwsWafv2RuleGroupExists(n string, v *wafv2.RuleGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No WAFv2 RuleGroup ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).wafv2conn
		resp, err := conn.GetRuleGroup(&wafv2.GetRuleGroupInput{
			Id:    aws.String(rs.Primary.ID),
			Name:  aws.String(rs.Primary.Attributes["name"]),
			Scope: aws.String(rs.Primary.Attributes["scope"]),
		})

		if err != nil {
			return err
		}

		if resp == nil || resp.RuleGroup == nil {
			return fmt.Errorf("Error getting WAFv2 RuleGroup")
		}

		if aws.StringValue(resp.RuleGroup.Id) == rs.Primary.ID {
			*v = *resp.RuleGroup
			return nil
		}

		return fmt.Errorf("WAFv2 RuleGroup (%s) not found", rs.Primary.ID)
	}
}

func testAccAwsWafv2RuleGroupConfig_Basic(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity    = 2
  name        = "%s"
  description = "%s"
  scope       = "REGIONAL"

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
`, name, name)
}

func testAccAwsWafv2RuleGroupConfig_BasicUpdate(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity    = 50
  name        = "%s"
  description = "Updated"
  scope       = "REGIONAL"

  rule {
    name     = "rule-1"
    priority = 1

    action {
      count {}
    }

    statement {
      geo_match_statement {
        country_codes = ["US", "NL"]
      }
    }

    visibility_config {
      cloudwatch_metrics_enabled = false
      metric_name                = "friendly-rule-metric-name"
      sampled_requests_enabled   = false
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
`, name)
}

func testAccAwsWafv2RuleGroupConfig_UpdateMultipleRules(name string, ruleName1, ruleName2 string, priority1, priority2 int) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity    = 50
  name        = "%[1]s"
  description = "Updated"
  scope       = "REGIONAL"

  rule {
    name     = "%[2]s"
    priority = %[3]d

    action {
      count {}
    }

    statement {
      geo_match_statement {
        country_codes = ["US", "NL"]
      }
    }

    visibility_config {
      cloudwatch_metrics_enabled = false
      metric_name                = "%[2]s"
      sampled_requests_enabled   = false
    }
  }

  rule {
    name     = "%[4]s"
    priority = %[5]d

    action {
      block {}
    }

    statement {
      size_constraint_statement {
        comparison_operator = "LT"
        size                = 50

        field_to_match {
          query_string {}
        }

        text_transformation {
          priority = 5
          type     = "NONE"
        }

        text_transformation {
          priority = 2
          type     = "CMD_LINE"
        }
      }
    }

    visibility_config {
      cloudwatch_metrics_enabled = false
      metric_name                = "%[4]s"
      sampled_requests_enabled   = false
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
`, name, ruleName1, priority1, ruleName2, priority2)
}

func testAccAwsWafv2RuleGroupConfig_UpdateCapacity(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity    = 3
  name        = "%s"
  description = "%s"
  scope       = "REGIONAL"

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
`, name, name)
}

func testAccAwsWafv2RuleGroupConfig_UpdateMetricName(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity    = 2
  name        = "%s"
  description = "%s"
  scope       = "REGIONAL"

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "updated-friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
`, name, name)
}

func testAccAwsWafv2RuleGroupConfig_Minimal(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity = 2
  name     = "%s"
  scope    = "REGIONAL"

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
`, name)
}

func testAccAwsWafv2RuleGroupConfig_RuleActionAllow(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity = 2
  name     = "%s"
  scope    = "REGIONAL"

  rule {
    name     = "rule-1"
    priority = 1

    action {
      allow {}
    }

    statement {
      geo_match_statement {
        country_codes = ["US", "NL"]
      }
    }

    visibility_config {
      cloudwatch_metrics_enabled = false
      metric_name                = "friendly-rule-metric-name"
      sampled_requests_enabled   = false
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
`, name)
}

func testAccAwsWafv2RuleGroupConfig_RuleActionBlock(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity = 2
  name     = "%s"
  scope    = "REGIONAL"

  rule {
    name     = "rule-1"
    priority = 1

    action {
      block {}
    }

    statement {
      geo_match_statement {
        country_codes = ["US", "NL"]
      }
    }

    visibility_config {
      cloudwatch_metrics_enabled = false
      metric_name                = "friendly-rule-metric-name"
      sampled_requests_enabled   = false
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
`, name)
}

func testAccAwsWafv2RuleGroupConfig_RuleActionCount(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity = 2
  name     = "%s"
  scope    = "REGIONAL"

  rule {
    name     = "rule-1"
    priority = 1

    action {
      count {}
    }

    statement {
      geo_match_statement {
        country_codes = ["US", "NL"]
      }
    }

    visibility_config {
      cloudwatch_metrics_enabled = false
      metric_name                = "friendly-rule-metric-name"
      sampled_requests_enabled   = false
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
`, name)
}

func testAccAwsWafv2RuleGroupConfig_ByteMatchStatement(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity = 300
  name     = "%s"
  scope    = "REGIONAL"

  rule {
    name     = "rule-1"
    priority = 1

    action {
      allow {}
    }

    statement {
      byte_match_statement {
        positional_constraint = "CONTAINS"
        search_string         = "word"

        field_to_match {
          all_query_arguments {}
        }

        text_transformation {
          priority = 5
          type     = "NONE"
        }

        text_transformation {
          priority = 2
          type     = "LOWERCASE"
        }
      }
    }

    visibility_config {
      cloudwatch_metrics_enabled = false
      metric_name                = "friendly-rule-metric-name"
      sampled_requests_enabled   = false
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
`, name)
}

func testAccAwsWafv2RuleGroupConfig_ByteMatchStatement_Update(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity = 30
  name     = "%s"
  scope    = "REGIONAL"

  rule {
    name     = "rule-1"
    priority = 1

    action {
      allow {}
    }

    statement {
      byte_match_statement {
        positional_constraint = "EXACTLY"
        search_string         = "sentence"

        field_to_match {
          all_query_arguments {}
        }

        text_transformation {
          priority = 3
          type     = "CMD_LINE"
        }
      }
    }

    visibility_config {
      cloudwatch_metrics_enabled = false
      metric_name                = "friendly-rule-metric-name"
      sampled_requests_enabled   = false
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
`, name)
}

func testAccAwsWafv2RuleGroupConfig_ByteMatchStatement_FieldToMatchAllQueryArguments(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity = 30
  name     = "%s"
  scope    = "REGIONAL"

  rule {
    name     = "rule-1"
    priority = 1

    action {
      allow {}
    }

    statement {
      byte_match_statement {
        positional_constraint = "CONTAINS"
        search_string         = "word"

        field_to_match {
          all_query_arguments {}
        }

        text_transformation {
          priority = 5
          type     = "NONE"
        }
      }
    }

    visibility_config {
      cloudwatch_metrics_enabled = false
      metric_name                = "friendly-rule-metric-name"
      sampled_requests_enabled   = false
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
`, name)
}

func testAccAwsWafv2RuleGroupConfig_ByteMatchStatement_FieldToMatchBody(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity = 15
  name     = "%s"
  scope    = "REGIONAL"

  rule {
    name     = "rule-1"
    priority = 1

    action {
      allow {}
    }

    statement {
      byte_match_statement {
        positional_constraint = "CONTAINS"
        search_string         = "word"

        field_to_match {
          body {}
        }

        text_transformation {
          priority = 1
          type     = "NONE"
        }
      }
    }

    visibility_config {
      cloudwatch_metrics_enabled = false
      metric_name                = "friendly-rule-metric-name"
      sampled_requests_enabled   = false
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
`, name)
}

func testAccAwsWafv2RuleGroupConfig_ByteMatchStatement_FieldToMatchMethod(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity = 15
  name     = "%s"
  scope    = "REGIONAL"

  rule {
    name     = "rule-1"
    priority = 1

    action {
      allow {}
    }

    statement {
      byte_match_statement {
        positional_constraint = "CONTAINS"
        search_string         = "word"

        field_to_match {
          method {}
        }

        text_transformation {
          priority = 1
          type     = "NONE"
        }
      }
    }

    visibility_config {
      cloudwatch_metrics_enabled = false
      metric_name                = "friendly-rule-metric-name"
      sampled_requests_enabled   = false
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
`, name)
}

func testAccAwsWafv2RuleGroupConfig_ByteMatchStatement_FieldToMatchQueryString(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity = 15
  name     = "%s"
  scope    = "REGIONAL"

  rule {
    name     = "rule-1"
    priority = 1

    action {
      allow {}
    }

    statement {
      byte_match_statement {
        positional_constraint = "CONTAINS"
        search_string         = "word"

        field_to_match {
          query_string {}
        }

        text_transformation {
          priority = 1
          type     = "NONE"
        }
      }
    }

    visibility_config {
      cloudwatch_metrics_enabled = false
      metric_name                = "friendly-rule-metric-name"
      sampled_requests_enabled   = false
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
`, name)
}

func testAccAwsWafv2RuleGroupConfig_ByteMatchStatement_FieldToMatchSingleHeader(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity = 15
  name     = "%s"
  scope    = "REGIONAL"

  rule {
    name     = "rule-1"
    priority = 1

    action {
      allow {}
    }

    statement {
      byte_match_statement {
        positional_constraint = "CONTAINS"
        search_string         = "word"

        field_to_match {
          single_header {
            name = "a-forty-character-long-header-name-40-40"
          }
        }

        text_transformation {
          priority = 1
          type     = "NONE"
        }
      }
    }

    visibility_config {
      cloudwatch_metrics_enabled = false
      metric_name                = "friendly-rule-metric-name"
      sampled_requests_enabled   = false
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
`, name)
}

func testAccAwsWafv2RuleGroupConfig_ByteMatchStatement_FieldToMatchSingleQueryArgument(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity = 30
  name     = "%s"
  scope    = "REGIONAL"

  rule {
    name     = "rule-1"
    priority = 1

    action {
      allow {}
    }

    statement {
      byte_match_statement {
        positional_constraint = "CONTAINS"
        search_string         = "word"

        field_to_match {
          single_query_argument {
            name = "a-max-30-characters-long-name-"
          }
        }

        text_transformation {
          priority = 1
          type     = "NONE"
        }
      }
    }

    visibility_config {
      cloudwatch_metrics_enabled = false
      metric_name                = "friendly-rule-metric-name"
      sampled_requests_enabled   = false
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
`, name)
}

func testAccAwsWafv2RuleGroupConfig_ByteMatchStatement_FieldToMatchUriPath(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity = 15
  name     = "%s"
  scope    = "REGIONAL"

  rule {
    name     = "rule-1"
    priority = 1

    action {
      allow {}
    }

    statement {
      byte_match_statement {
        positional_constraint = "CONTAINS"
        search_string         = "word"

        field_to_match {
          uri_path {}
        }

        text_transformation {
          priority = 1
          type     = "NONE"
        }
      }
    }

    visibility_config {
      cloudwatch_metrics_enabled = false
      metric_name                = "friendly-rule-metric-name"
      sampled_requests_enabled   = false
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
`, name)
}

func testAccAwsWafv2RuleGroupConfig_IpSetReferenceStatement(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_ip_set" "test" {
  name               = "ip-set-%s"
  scope              = "REGIONAL"
  ip_address_version = "IPV4"
  addresses          = ["1.1.1.1/32", "2.2.2.2/32"]
}

resource "aws_wafv2_rule_group" "test" {
  capacity = 2
  name     = "%s"
  scope    = "REGIONAL"

  rule {
    name     = "rule-1"
    priority = 1

    action {
      allow {}
    }

    statement {
      ip_set_reference_statement {
        arn = aws_wafv2_ip_set.test.arn
      }
    }

    visibility_config {
      cloudwatch_metrics_enabled = false
      metric_name                = "friendly-rule-metric-name"
      sampled_requests_enabled   = false
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
`, name, name)
}

func testAccAwsWafv2RuleGroupConfig_IpSetReferenceStatement_IPSetForwardedIPConfig(name, fallbackBehavior, headerName, position string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_ip_set" "test" {
  name               = "ip-set-%[1]s"
  scope              = "REGIONAL"
  ip_address_version = "IPV4"
  addresses          = ["1.1.1.1/32", "2.2.2.2/32"]
}

resource "aws_wafv2_rule_group" "test" {
  capacity = 5
  name     = "%[1]s"
  scope    = "REGIONAL"

  rule {
    name     = "rule-1"
    priority = 1

    action {
      allow {}
    }

    statement {
      ip_set_reference_statement {
        arn = aws_wafv2_ip_set.test.arn
        ip_set_forwarded_ip_config {
          fallback_behavior = "%[2]s"
          header_name       = "%[3]s"
          position          = "%[4]s"
        }
      }
    }

    visibility_config {
      cloudwatch_metrics_enabled = false
      metric_name                = "friendly-rule-metric-name"
      sampled_requests_enabled   = false
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
`, name, fallbackBehavior, headerName, position)
}

func testAccAwsWafv2RuleGroupConfig_GeoMatchStatement(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity = 2
  name     = "%s"
  scope    = "REGIONAL"

  rule {
    name     = "rule-1"
    priority = 1

    action {
      allow {}
    }

    statement {
      geo_match_statement {
        country_codes = ["US", "NL"]
      }
    }

    visibility_config {
      cloudwatch_metrics_enabled = false
      metric_name                = "friendly-rule-metric-name"
      sampled_requests_enabled   = false
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
`, name)
}

func testAccAwsWafv2RuleGroupConfig_GeoMatchStatement_ForwardedIPConfig(name, fallbackBehavior, headerName string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity = 2
  name     = "%s"
  scope    = "REGIONAL"

  rule {
    name     = "rule-1"
    priority = 1

    action {
      allow {}
    }

    statement {
      geo_match_statement {
        country_codes = ["US", "NL"]
        forwarded_ip_config {
          fallback_behavior = "%s"
          header_name       = "%s"
        }
      }
    }

    visibility_config {
      cloudwatch_metrics_enabled = false
      metric_name                = "friendly-rule-metric-name"
      sampled_requests_enabled   = false
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
`, name, fallbackBehavior, headerName)
}

func testAccAwsWafv2RuleGroupConfig_GeoMatchStatement_Update(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity = 2
  name     = "%s"
  scope    = "REGIONAL"

  rule {
    name     = "rule-1"
    priority = 1

    action {
      allow {}
    }

    statement {
      geo_match_statement {
        country_codes = ["ZM", "EE", "MM"]
      }
    }

    visibility_config {
      cloudwatch_metrics_enabled = false
      metric_name                = "friendly-rule-metric-name"
      sampled_requests_enabled   = false
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
`, name)
}

func testAccAwsWafv2RuleGroupConfig_LogicalRuleStatement_And(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity = 2
  name     = "%s"
  scope    = "REGIONAL"

  rule {
    name     = "rule-1"
    priority = 1

    action {
      allow {}
    }

    statement {
      and_statement {
        statement {
          geo_match_statement {
            country_codes = ["US"]
          }
        }

        statement {
          geo_match_statement {
            country_codes = ["NL"]
          }
        }
      }
    }

    visibility_config {
      cloudwatch_metrics_enabled = false
      metric_name                = "friendly-rule-metric-name"
      sampled_requests_enabled   = false
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
`, name)
}

func testAccAwsWafv2RuleGroupConfig_LogicalRuleStatement_NotAnd(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity = 2
  name     = "%s"
  scope    = "REGIONAL"

  rule {
    name     = "rule-1"
    priority = 1

    action {
      allow {}
    }

    statement {
      not_statement {
        statement {
          and_statement {
            statement {
              geo_match_statement {
                country_codes = ["US"]
              }
            }

            statement {
              geo_match_statement {
                country_codes = ["NL"]
              }
            }
          }
        }
      }
    }

    visibility_config {
      cloudwatch_metrics_enabled = false
      metric_name                = "friendly-rule-metric-name"
      sampled_requests_enabled   = false
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
`, name)
}

func testAccAwsWafv2RuleGroupConfig_LogicalRuleStatement_OrNotAnd(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity = 3
  name     = "%s"
  scope    = "REGIONAL"

  rule {
    name     = "rule-1"
    priority = 1

    action {
      allow {}
    }

    statement {
      or_statement {
        statement {
          not_statement {
            statement {
              geo_match_statement {
                country_codes = ["DE"]
              }
            }
          }
        }

        statement {
          and_statement {
            statement {
              geo_match_statement {
                country_codes = ["US"]
              }
            }

            statement {
              geo_match_statement {
                country_codes = ["NL"]
              }
            }
          }
        }
      }
    }

    visibility_config {
      cloudwatch_metrics_enabled = false
      metric_name                = "friendly-rule-metric-name"
      sampled_requests_enabled   = false
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
`, name)
}

func testAccAwsWafv2RuleGroupConfig_RegexPatternSetReferenceStatement(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_regex_pattern_set" "test" {
  name  = "regex-pattern-set-%s"
  scope = "REGIONAL"

  regular_expression {
    regex_string = "one"
  }
}

resource "aws_wafv2_rule_group" "test" {
  capacity = 50
  name     = "%s"
  scope    = "REGIONAL"

  rule {
    name     = "rule-1"
    priority = 1

    action {
      allow {}
    }

    statement {
      regex_pattern_set_reference_statement {
        arn = aws_wafv2_regex_pattern_set.test.arn

        field_to_match {
          body {}
        }

        text_transformation {
          priority = 2
          type     = "NONE"
        }
      }
    }

    visibility_config {
      cloudwatch_metrics_enabled = false
      metric_name                = "friendly-rule-metric-name"
      sampled_requests_enabled   = false
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
`, name, name)
}

func testAccAwsWafv2RuleGroupConfig_SizeConstraintStatement(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity = 30
  name     = "%s"
  scope    = "REGIONAL"

  rule {
    name     = "rule-1"
    priority = 1

    action {
      allow {}
    }

    statement {
      size_constraint_statement {
        comparison_operator = "GT"
        size                = 100

        field_to_match {
          method {}
        }

        text_transformation {
          priority = 5
          type     = "NONE"
        }
      }
    }

    visibility_config {
      cloudwatch_metrics_enabled = false
      metric_name                = "friendly-rule-metric-name"
      sampled_requests_enabled   = false
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
`, name)
}

func testAccAwsWafv2RuleGroupConfig_SizeConstraintStatement_Update(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity = 30
  name     = "%s"
  scope    = "REGIONAL"

  rule {
    name     = "rule-1"
    priority = 1

    action {
      allow {}
    }

    statement {
      size_constraint_statement {
        comparison_operator = "LT"
        size                = 50

        field_to_match {
          query_string {}
        }

        text_transformation {
          priority = 5
          type     = "NONE"
        }

        text_transformation {
          priority = 2
          type     = "CMD_LINE"
        }
      }
    }

    visibility_config {
      cloudwatch_metrics_enabled = false
      metric_name                = "friendly-rule-metric-name"
      sampled_requests_enabled   = false
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
`, name)
}

func testAccAwsWafv2RuleGroupConfig_SqliMatchStatement(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity = 300
  name     = "%s"
  scope    = "REGIONAL"

  rule {
    name     = "rule-1"
    priority = 1

    action {
      allow {}
    }

    statement {
      sqli_match_statement {
        field_to_match {
          all_query_arguments {}
        }

        text_transformation {
          priority = 5
          type     = "URL_DECODE"
        }

        text_transformation {
          priority = 2
          type     = "LOWERCASE"
        }
      }
    }

    visibility_config {
      cloudwatch_metrics_enabled = false
      metric_name                = "friendly-rule-metric-name"
      sampled_requests_enabled   = false
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
`, name)
}

func testAccAwsWafv2RuleGroupConfig_SqliMatchStatement_Update(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity = 300
  name     = "%s"
  scope    = "REGIONAL"

  rule {
    name     = "rule-1"
    priority = 1

    action {
      allow {}
    }

    statement {
      sqli_match_statement {
        field_to_match {
          body {}
        }

        text_transformation {
          priority = 5
          type     = "URL_DECODE"
        }

        text_transformation {
          priority = 4
          type     = "HTML_ENTITY_DECODE"
        }

        text_transformation {
          priority = 3
          type     = "COMPRESS_WHITE_SPACE"
        }
      }
    }

    visibility_config {
      cloudwatch_metrics_enabled = false
      metric_name                = "friendly-rule-metric-name"
      sampled_requests_enabled   = false
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
`, name)
}

func testAccAwsWafv2RuleGroupConfig_XssMatchStatement(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity = 300
  name     = "%s"
  scope    = "REGIONAL"

  rule {
    name     = "rule-1"
    priority = 1

    action {
      block {}
    }

    statement {
      xss_match_statement {
        field_to_match {
          body {}
        }

        text_transformation {
          priority = 2
          type     = "NONE"
        }
      }
    }

    visibility_config {
      cloudwatch_metrics_enabled = false
      metric_name                = "friendly-rule-metric-name"
      sampled_requests_enabled   = false
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
`, name)
}

func testAccAwsWafv2RuleGroupConfig_XssMatchStatement_Update(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity = 300
  name     = "%s"
  scope    = "REGIONAL"

  rule {
    name     = "rule-1"
    priority = 1

    action {
      allow {}
    }

    statement {
      xss_match_statement {
        field_to_match {
          body {}
        }

        text_transformation {
          priority = 2
          type     = "URL_DECODE"
        }
      }
    }

    visibility_config {
      cloudwatch_metrics_enabled = false
      metric_name                = "friendly-rule-metric-name"
      sampled_requests_enabled   = false
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
`, name)
}

func testAccAwsWafv2RuleGroupConfig_OneTag(name, tagKey, tagValue string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity    = 2
  name        = "%s"
  description = "%s"
  scope       = "REGIONAL"

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }

  tags = {
    "%s" = "%s"
  }
}
`, name, name, tagKey, tagValue)
}

func testAccAwsWafv2RuleGroupConfig_TwoTags(name, tag1Key, tag1Value, tag2Key, tag2Value string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity    = 2
  name        = "%s"
  description = "%s"
  scope       = "REGIONAL"

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }

  tags = {
    "%s" = "%s"
    "%s" = "%s"
  }
}
`, name, name, tag1Key, tag1Value, tag2Key, tag2Value)
}

func testAccAwsWafv2RuleGroupImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return fmt.Sprintf("%s/%s/%s", rs.Primary.ID, rs.Primary.Attributes["name"], rs.Primary.Attributes["scope"]), nil
	}
}
