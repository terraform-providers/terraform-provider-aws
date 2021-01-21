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
	resource.AddTestSweepers("aws_wafv2_web_acl", &resource.Sweeper{
		Name: "aws_wafv2_web_acl",
		F:    testSweepWafv2WebAcls,
	})
}

func testSweepWafv2WebAcls(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).wafv2conn

	var sweeperErrs *multierror.Error

	input := &wafv2.ListWebACLsInput{
		Scope: aws.String(wafv2.ScopeRegional),
	}

	err = lister.ListWebACLsPages(conn, input, func(page *wafv2.ListWebACLsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, webAcl := range page.WebACLs {
			id := aws.StringValue(webAcl.Id)

			r := resourceAwsWafv2WebACL()
			d := r.Data(nil)
			d.SetId(id)
			d.Set("lock_token", webAcl.LockToken)
			d.Set("name", webAcl.Name)
			d.Set("scope", input.Scope)
			err := r.Delete(d, client)

			if err != nil {
				sweeperErr := fmt.Errorf("error deleting WAFv2 Web ACL (%s): %w", id, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}
		}

		return !lastPage
	})

	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping WAFv2 Web ACL sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error describing WAFv2 Web ACLs: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func TestAccAwsWafv2WebACL_basic(t *testing.T) {
	var v wafv2.WebACL
	webACLName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_wafv2_web_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSWafv2ScopeRegional(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsWafv2WebACLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsWafv2WebACLConfig_Basic(webACLName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2WebACLExists(resourceName, &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "name", webACLName),
					resource.TestCheckResourceAttr(resourceName, "description", webACLName),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "scope", wafv2.ScopeRegional),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.allow.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.block.#", "0"),
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
				ImportStateIdFunc: testAccAwsWafv2WebACLImportStateIdFunc(resourceName),
			},
		},
	})
}

func TestAccAwsWafv2WebACL_updateRule(t *testing.T) {
	var v wafv2.WebACL
	webACLName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_wafv2_web_acl.test"
	ruleName1 := fmt.Sprintf("%s-1", webACLName)
	ruleName2 := fmt.Sprintf("%s-2", webACLName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSWafv2ScopeRegional(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsWafv2WebACLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsWafv2WebACLConfig_BasicRule(webACLName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2WebACLExists(resourceName, &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "name", webACLName),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated"),
					resource.TestCheckResourceAttr(resourceName, "scope", wafv2.ScopeRegional),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.allow.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.block.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.cloudwatch_metrics_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.metric_name", "friendly-metric-name"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.sampled_requests_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"name":                ruleName1,
						"priority":            "10",
						"action.#":            "1",
						"action.0.allow.#":    "0",
						"action.0.block.#":    "0",
						"action.0.count.#":    "1",
						"visibility_config.#": "1",
						"visibility_config.0.cloudwatch_metrics_enabled": "false",
						"visibility_config.0.metric_name":                fmt.Sprintf("%s-metric-name-1", webACLName),
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
				// Test step to verify additional rule block with first rule block unchanged
				Config: testAccAwsWafv2WebACLConfig_UpdateRuleNamePriorityMetric(webACLName, ruleName1, ruleName2, 10, 5),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2WebACLExists(resourceName, &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "name", webACLName),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated"),
					resource.TestCheckResourceAttr(resourceName, "scope", wafv2.ScopeRegional),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.allow.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.block.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.cloudwatch_metrics_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.metric_name", "friendly-metric-name"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.sampled_requests_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"name":                ruleName1,
						"priority":            "10",
						"action.#":            "1",
						"action.0.allow.#":    "0",
						"action.0.block.#":    "0",
						"action.0.count.#":    "1",
						"visibility_config.#": "1",
						"visibility_config.0.cloudwatch_metrics_enabled": "false",
						"visibility_config.0.metric_name":                ruleName1,
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
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"name":                ruleName2,
						"priority":            "5",
						"action.#":            "1",
						"action.0.allow.#":    "1",
						"action.0.block.#":    "0",
						"action.0.count.#":    "0",
						"visibility_config.#": "1",
						"visibility_config.0.cloudwatch_metrics_enabled": "false",
						"visibility_config.0.metric_name":                ruleName2,
						"visibility_config.0.sampled_requests_enabled":   "false",
						"statement.#":                                       "1",
						"statement.0.geo_match_statement.#":                 "1",
						"statement.0.geo_match_statement.0.country_codes.#": "2",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccAwsWafv2WebACLImportStateIdFunc(resourceName),
			},
		},
	})
}

func TestAccAwsWafv2WebACL_UpdateRuleProperties(t *testing.T) {
	var v wafv2.WebACL
	webACLName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_wafv2_web_acl.test"
	ruleName1 := fmt.Sprintf("%s-1", webACLName)
	ruleName2 := fmt.Sprintf("%s-2", webACLName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSWafv2ScopeRegional(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsWafv2WebACLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsWafv2WebACLConfig_UpdateRuleNamePriorityMetric(webACLName, ruleName1, ruleName2, 5, 10),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2WebACLExists(resourceName, &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "name", webACLName),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated"),
					resource.TestCheckResourceAttr(resourceName, "scope", wafv2.ScopeRegional),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.allow.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.block.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.cloudwatch_metrics_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.metric_name", "friendly-metric-name"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.sampled_requests_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"name":                ruleName1,
						"priority":            "5",
						"action.#":            "1",
						"action.0.allow.#":    "0",
						"action.0.block.#":    "0",
						"action.0.count.#":    "1",
						"visibility_config.#": "1",
						"visibility_config.0.cloudwatch_metrics_enabled": "false",
						"visibility_config.0.metric_name":                ruleName1,
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
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"name":                ruleName2,
						"priority":            "10",
						"action.#":            "1",
						"action.0.allow.#":    "1",
						"action.0.block.#":    "0",
						"action.0.count.#":    "0",
						"visibility_config.#": "1",
						"visibility_config.0.cloudwatch_metrics_enabled": "false",
						"visibility_config.0.metric_name":                ruleName2,
						"visibility_config.0.sampled_requests_enabled":   "false",
						"statement.#":                                       "1",
						"statement.0.geo_match_statement.#":                 "1",
						"statement.0.geo_match_statement.0.country_codes.#": "2",
					}),
				),
			},
			{
				Config: testAccAwsWafv2WebACLConfig_UpdateRuleNamePriorityMetric(webACLName, ruleName1, ruleName2, 10, 5),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2WebACLExists(resourceName, &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "name", webACLName),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated"),
					resource.TestCheckResourceAttr(resourceName, "scope", wafv2.ScopeRegional),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.allow.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.block.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.cloudwatch_metrics_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.metric_name", "friendly-metric-name"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.sampled_requests_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"name":                ruleName1,
						"priority":            "10",
						"action.#":            "1",
						"action.0.allow.#":    "0",
						"action.0.block.#":    "0",
						"action.0.count.#":    "1",
						"visibility_config.#": "1",
						"visibility_config.0.cloudwatch_metrics_enabled": "false",
						"visibility_config.0.metric_name":                ruleName1,
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
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"name":                ruleName2,
						"priority":            "5",
						"action.#":            "1",
						"action.0.allow.#":    "1",
						"action.0.block.#":    "0",
						"action.0.count.#":    "0",
						"visibility_config.#": "1",
						"visibility_config.0.cloudwatch_metrics_enabled": "false",
						"visibility_config.0.metric_name":                ruleName2,
						"visibility_config.0.sampled_requests_enabled":   "false",
						"statement.#":                                       "1",
						"statement.0.geo_match_statement.#":                 "1",
						"statement.0.geo_match_statement.0.country_codes.#": "2",
					}),
				),
			},
			{
				Config: testAccAwsWafv2WebACLConfig_UpdateRuleNamePriorityMetric(webACLName, ruleName1, "updated", 10, 5),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2WebACLExists(resourceName, &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "name", webACLName),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated"),
					resource.TestCheckResourceAttr(resourceName, "scope", wafv2.ScopeRegional),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.allow.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.block.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.cloudwatch_metrics_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.metric_name", "friendly-metric-name"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.sampled_requests_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"name":                ruleName1,
						"priority":            "10",
						"action.#":            "1",
						"action.0.allow.#":    "0",
						"action.0.block.#":    "0",
						"action.0.count.#":    "1",
						"visibility_config.#": "1",
						"visibility_config.0.cloudwatch_metrics_enabled": "false",
						"visibility_config.0.metric_name":                ruleName1,
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
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"name":                "updated",
						"priority":            "5",
						"action.#":            "1",
						"action.0.allow.#":    "1",
						"action.0.block.#":    "0",
						"action.0.count.#":    "0",
						"visibility_config.#": "1",
						"visibility_config.0.cloudwatch_metrics_enabled": "false",
						"visibility_config.0.metric_name":                "updated",
						"visibility_config.0.sampled_requests_enabled":   "false",
						"statement.#":                                       "1",
						"statement.0.geo_match_statement.#":                 "1",
						"statement.0.geo_match_statement.0.country_codes.#": "2",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccAwsWafv2WebACLImportStateIdFunc(resourceName),
			},
		},
	})
}

func TestAccAwsWafv2WebACL_ChangeNameForceNew(t *testing.T) {
	var before, after wafv2.WebACL
	webACLName := acctest.RandomWithPrefix("tf-acc-test")
	ruleGroupNewName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_wafv2_web_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSWafv2ScopeRegional(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsWafv2WebACLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsWafv2WebACLConfig_Basic(webACLName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2WebACLExists(resourceName, &before),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "name", webACLName),
					resource.TestCheckResourceAttr(resourceName, "description", webACLName),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "scope", wafv2.ScopeRegional),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.allow.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.block.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.cloudwatch_metrics_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.metric_name", "friendly-metric-name"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.sampled_requests_enabled", "false"),
				),
			},
			{
				Config: testAccAwsWafv2WebACLConfig_Basic(ruleGroupNewName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2WebACLExists(resourceName, &after),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "name", ruleGroupNewName),
					resource.TestCheckResourceAttr(resourceName, "description", ruleGroupNewName),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "scope", wafv2.ScopeRegional),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.allow.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.block.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.cloudwatch_metrics_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.metric_name", "friendly-metric-name"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.sampled_requests_enabled", "false"),
				),
			},
		},
	})
}

func TestAccAwsWafv2WebACL_Disappears(t *testing.T) {
	var v wafv2.WebACL
	webACLName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_wafv2_web_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSWafv2ScopeRegional(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsWafv2WebACLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsWafv2WebACLConfig_Minimal(webACLName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2WebACLExists(resourceName, &v),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsWafv2WebACL(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAwsWafv2WebACL_ManagedRuleGroupStatement(t *testing.T) {
	var v wafv2.WebACL
	webACLName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_wafv2_web_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSWafv2ScopeRegional(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsWafv2WebACLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsWafv2WebACLConfig_ManagedRuleGroupStatement(webACLName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2WebACLExists(resourceName, &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "name", webACLName),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"name":                      "rule-1",
						"action.#":                  "0",
						"override_action.#":         "1",
						"override_action.0.count.#": "0",
						"override_action.0.none.#":  "1",
						"statement.#":               "1",
						"statement.0.managed_rule_group_statement.#":                 "1",
						"statement.0.managed_rule_group_statement.0.name":            "AWSManagedRulesCommonRuleSet",
						"statement.0.managed_rule_group_statement.0.vendor_name":     "AWS",
						"statement.0.managed_rule_group_statement.0.excluded_rule.#": "0",
					}),
				),
			},
			{
				Config: testAccAwsWafv2WebACLConfig_ManagedRuleGroupStatement_Update(webACLName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2WebACLExists(resourceName, &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "name", webACLName),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"name":                      "rule-1",
						"action.#":                  "0",
						"override_action.#":         "1",
						"override_action.0.count.#": "1",
						"override_action.0.none.#":  "0",
						"statement.#":               "1",
						"statement.0.managed_rule_group_statement.#":                      "1",
						"statement.0.managed_rule_group_statement.0.name":                 "AWSManagedRulesCommonRuleSet",
						"statement.0.managed_rule_group_statement.0.vendor_name":          "AWS",
						"statement.0.managed_rule_group_statement.0.excluded_rule.#":      "2",
						"statement.0.managed_rule_group_statement.0.excluded_rule.0.name": "SizeRestrictions_QUERYSTRING",
						"statement.0.managed_rule_group_statement.0.excluded_rule.1.name": "NoUserAgent_HEADER",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccAwsWafv2WebACLImportStateIdFunc(resourceName),
			},
		},
	})
}

func TestAccAwsWafv2WebACL_Minimal(t *testing.T) {
	var v wafv2.WebACL
	webACLName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_wafv2_web_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSWafv2ScopeRegional(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsWafv2WebACLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsWafv2WebACLConfig_Minimal(webACLName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2WebACLExists(resourceName, &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "name", webACLName),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "scope", wafv2.ScopeRegional),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.allow.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.block.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.cloudwatch_metrics_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.metric_name", "friendly-metric-name"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.sampled_requests_enabled", "false"),
				),
			},
		},
	})
}

func TestAccAwsWafv2WebACL_RateBasedStatement(t *testing.T) {
	var v wafv2.WebACL
	webACLName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_wafv2_web_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSWafv2ScopeRegional(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsWafv2WebACLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsWafv2WebACLConfig_RateBasedStatement(webACLName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2WebACLExists(resourceName, &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "name", webACLName),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"name":                               "rule-1",
						"action.#":                           "1",
						"action.0.allow.#":                   "0",
						"action.0.block.#":                   "0",
						"action.0.count.#":                   "1",
						"statement.#":                        "1",
						"statement.0.rate_based_statement.#": "1",
						"statement.0.rate_based_statement.0.aggregate_key_type":     "IP",
						"statement.0.rate_based_statement.0.forwarded_ip_config.#":  "0",
						"statement.0.rate_based_statement.0.limit":                  "50000",
						"statement.0.rate_based_statement.0.scope_down_statement.#": "0",
					}),
				),
			},
			{
				Config: testAccAwsWafv2WebACLConfig_RateBasedStatement_Update(webACLName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2WebACLExists(resourceName, &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "name", webACLName),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"name":                               "rule-1",
						"action.#":                           "1",
						"action.0.allow.#":                   "0",
						"action.0.block.#":                   "0",
						"action.0.count.#":                   "1",
						"statement.#":                        "1",
						"statement.0.rate_based_statement.#": "1",
						"statement.0.rate_based_statement.0.aggregate_key_type":                                           "IP",
						"statement.0.rate_based_statement.0.forwarded_ip_config.#":                                        "0",
						"statement.0.rate_based_statement.0.limit":                                                        "10000",
						"statement.0.rate_based_statement.0.scope_down_statement.#":                                       "1",
						"statement.0.rate_based_statement.0.scope_down_statement.0.geo_match_statement.#":                 "1",
						"statement.0.rate_based_statement.0.scope_down_statement.0.geo_match_statement.0.country_codes.#": "2",
						"statement.0.rate_based_statement.0.scope_down_statement.0.geo_match_statement.0.country_codes.0": "US",
						"statement.0.rate_based_statement.0.scope_down_statement.0.geo_match_statement.0.country_codes.1": "NL",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccAwsWafv2WebACLImportStateIdFunc(resourceName),
			},
		},
	})
}

func TestAccAwsWafv2WebACL_GeoMatchStatement(t *testing.T) {
	var v wafv2.WebACL
	webACLName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_wafv2_web_acl.test"
	countryCode := fmt.Sprintf("%q", "US")
	countryCodes := fmt.Sprintf("%s, %q", countryCode, "CA")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSWafv2ScopeRegional(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsWafv2WebACLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsWafv2WebACLConfig_GeoMatchStatement(webACLName, countryCode),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2WebACLExists(resourceName, &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "name", webACLName),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.allow.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.block.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "scope", "REGIONAL"),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"name":                              "rule-1",
						"action.#":                          "1",
						"action.0.allow.#":                  "0",
						"action.0.block.#":                  "1",
						"action.0.count.#":                  "0",
						"priority":                          "1",
						"statement.#":                       "1",
						"statement.0.geo_match_statement.#": "1",
						"statement.0.geo_match_statement.0.country_codes.#":       "1",
						"statement.0.geo_match_statement.0.country_codes.0":       "US",
						"statement.0.geo_match_statement.0.forwarded_ip_config.#": "0",
						"visibility_config.#":                            "1",
						"visibility_config.0.cloudwatch_metrics_enabled": "false",
						"visibility_config.0.metric_name":                "friendly-rule-metric-name",
						"visibility_config.0.sampled_requests_enabled":   "false",
					}),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.cloudwatch_metrics_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.metric_name", "friendly-metric-name"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.sampled_requests_enabled", "false"),
				),
			},
			{
				Config: testAccAwsWafv2WebACLConfig_GeoMatchStatement(webACLName, countryCodes),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2WebACLExists(resourceName, &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "name", webACLName),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.allow.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.block.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "scope", "REGIONAL"),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"name":                              "rule-1",
						"action.#":                          "1",
						"action.0.allow.#":                  "0",
						"action.0.block.#":                  "1",
						"action.0.count.#":                  "0",
						"priority":                          "1",
						"statement.#":                       "1",
						"statement.0.geo_match_statement.#": "1",
						"statement.0.geo_match_statement.0.country_codes.#":       "2",
						"statement.0.geo_match_statement.0.country_codes.0":       "US",
						"statement.0.geo_match_statement.0.country_codes.1":       "CA",
						"statement.0.geo_match_statement.0.forwarded_ip_config.#": "0",
						"visibility_config.#":                                     "1",
						"visibility_config.0.cloudwatch_metrics_enabled":          "false",
						"visibility_config.0.metric_name":                         "friendly-rule-metric-name",
						"visibility_config.0.sampled_requests_enabled":            "false",
					}),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.cloudwatch_metrics_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.metric_name", "friendly-metric-name"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.sampled_requests_enabled", "false"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccAwsWafv2WebACLImportStateIdFunc(resourceName),
			},
		},
	})
}

func TestAccAwsWafv2WebACL_GeoMatchStatement_ForwardedIPConfig(t *testing.T) {
	var v wafv2.WebACL
	webACLName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_wafv2_web_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSWafv2ScopeRegional(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsWafv2WebACLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsWafv2WebACLConfig_GeoMatchStatement_ForwardedIPConfig(webACLName, "MATCH", "X-Forwarded-For"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2WebACLExists(resourceName, &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "name", webACLName),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.#":                            "1",
						"statement.0.or_statement.#":             "1",
						"statement.0.or_statement.0.statement.#": "2",
						"statement.0.or_statement.0.statement.0.geo_match_statement.#":                                         "1",
						"statement.0.or_statement.0.statement.0.geo_match_statement.0.country_codes.#":                         "1",
						"statement.0.or_statement.0.statement.0.geo_match_statement.0.forwarded_ip_config.#":                   "0",
						"statement.0.or_statement.0.statement.1.geo_match_statement.#":                                         "1",
						"statement.0.or_statement.0.statement.1.geo_match_statement.0.country_codes.#":                         "1",
						"statement.0.or_statement.0.statement.1.geo_match_statement.0.forwarded_ip_config.#":                   "1",
						"statement.0.or_statement.0.statement.1.geo_match_statement.0.forwarded_ip_config.0.fallback_behavior": "MATCH",
						"statement.0.or_statement.0.statement.1.geo_match_statement.0.forwarded_ip_config.0.header_name":       "X-Forwarded-For",
					}),
				),
			},
			{
				Config: testAccAwsWafv2WebACLConfig_GeoMatchStatement_ForwardedIPConfig(webACLName, "NO_MATCH", "Updated"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2WebACLExists(resourceName, &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "name", webACLName),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.#":                            "1",
						"statement.0.or_statement.#":             "1",
						"statement.0.or_statement.0.statement.#": "2",
						"statement.0.or_statement.0.statement.0.geo_match_statement.#":                                         "1",
						"statement.0.or_statement.0.statement.0.geo_match_statement.0.country_codes.#":                         "1",
						"statement.0.or_statement.0.statement.0.geo_match_statement.0.forwarded_ip_config.#":                   "0",
						"statement.0.or_statement.0.statement.1.geo_match_statement.#":                                         "1",
						"statement.0.or_statement.0.statement.1.geo_match_statement.0.country_codes.#":                         "1",
						"statement.0.or_statement.0.statement.1.geo_match_statement.0.forwarded_ip_config.#":                   "1",
						"statement.0.or_statement.0.statement.1.geo_match_statement.0.forwarded_ip_config.0.fallback_behavior": "NO_MATCH",
						"statement.0.or_statement.0.statement.1.geo_match_statement.0.forwarded_ip_config.0.header_name":       "Updated",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccAwsWafv2WebACLImportStateIdFunc(resourceName),
			},
		},
	})
}

func TestAccAwsWafv2WebACL_IPSetReferenceStatement(t *testing.T) {
	var v wafv2.WebACL
	webACLName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_wafv2_web_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSWafv2ScopeRegional(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsWafv2WebACLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsWafv2WebACLConfig_IPSetReferenceStatement(webACLName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2WebACLExists(resourceName, &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "name", webACLName),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.#": "1",
						"statement.0.ip_set_reference_statement.#":                              "1",
						"statement.0.ip_set_reference_statement.0.ip_set_forwarded_ip_config.#": "0",
						"visibility_config.#":                            "1",
						"visibility_config.0.cloudwatch_metrics_enabled": "false",
						"visibility_config.0.metric_name":                "friendly-rule-metric-name",
						"visibility_config.0.sampled_requests_enabled":   "false",
					}),
					resource.TestMatchTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]*regexp.Regexp{
						"statement.0.ip_set_reference_statement.0.arn": regexp.MustCompile(`regional/ipset/.+$`),
					}),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.cloudwatch_metrics_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.metric_name", "friendly-metric-name"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.sampled_requests_enabled", "false"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccAwsWafv2WebACLImportStateIdFunc(resourceName),
			},
		},
	})
}

func TestAccAwsWafv2WebACL_IPSetReferenceStatement_IPSetForwardedIPConfig(t *testing.T) {
	var v wafv2.WebACL
	webACLName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_wafv2_web_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSWafv2ScopeRegional(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsWafv2WebACLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsWafv2WebACLConfig_IPSetReferenceStatement_IPSetForwardedIPConfig(webACLName, "MATCH", "X-Forwarded-For", "FIRST"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2WebACLExists(resourceName, &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "name", webACLName),
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
				Config: testAccAwsWafv2WebACLConfig_IPSetReferenceStatement_IPSetForwardedIPConfig(webACLName, "NO_MATCH", "X-Forwarded-For", "LAST"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2WebACLExists(resourceName, &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "name", webACLName),
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
				Config: testAccAwsWafv2WebACLConfig_IPSetReferenceStatement_IPSetForwardedIPConfig(webACLName, "MATCH", "Updated", "ANY"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2WebACLExists(resourceName, &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "name", webACLName),
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
				Config: testAccAwsWafv2WebACLConfig_IPSetReferenceStatement(webACLName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2WebACLExists(resourceName, &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "name", webACLName),
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
				ImportStateIdFunc: testAccAwsWafv2WebACLImportStateIdFunc(resourceName),
			},
		},
	})
}

func TestAccAwsWafv2WebACL_RateBasedStatement_ForwardedIPConfig(t *testing.T) {
	var v wafv2.WebACL
	webACLName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_wafv2_web_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSWafv2ScopeRegional(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsWafv2WebACLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsWafv2WebACLConfig_RateBasedStatement_ForwardedIPConfig(webACLName, "MATCH", "X-Forwarded-For"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2WebACLExists(resourceName, &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "name", webACLName),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"name":                               "rule-1",
						"action.#":                           "1",
						"action.0.allow.#":                   "0",
						"action.0.block.#":                   "0",
						"action.0.count.#":                   "1",
						"statement.#":                        "1",
						"statement.0.rate_based_statement.#": "1",
						"statement.0.rate_based_statement.0.aggregate_key_type":                      "FORWARDED_IP",
						"statement.0.rate_based_statement.0.forwarded_ip_config.#":                   "1",
						"statement.0.rate_based_statement.0.forwarded_ip_config.0.fallback_behavior": "MATCH",
						"statement.0.rate_based_statement.0.forwarded_ip_config.0.header_name":       "X-Forwarded-For",
						"statement.0.rate_based_statement.0.limit":                                   "50000",
						"statement.0.rate_based_statement.0.scope_down_statement.#":                  "0",
					}),
				),
			},
			{
				Config: testAccAwsWafv2WebACLConfig_RateBasedStatement_ForwardedIPConfig(webACLName, "NO_MATCH", "Updated"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2WebACLExists(resourceName, &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "name", webACLName),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"name":                               "rule-1",
						"action.#":                           "1",
						"action.0.allow.#":                   "0",
						"action.0.block.#":                   "0",
						"action.0.count.#":                   "1",
						"statement.#":                        "1",
						"statement.0.rate_based_statement.#": "1",
						"statement.0.rate_based_statement.0.aggregate_key_type":                      "FORWARDED_IP",
						"statement.0.rate_based_statement.0.forwarded_ip_config.#":                   "1",
						"statement.0.rate_based_statement.0.forwarded_ip_config.0.fallback_behavior": "NO_MATCH",
						"statement.0.rate_based_statement.0.forwarded_ip_config.0.header_name":       "Updated",
						"statement.0.rate_based_statement.0.limit":                                   "50000",
						"statement.0.rate_based_statement.0.scope_down_statement.#":                  "0",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccAwsWafv2WebACLImportStateIdFunc(resourceName),
			},
		},
	})
}

func TestAccAwsWafv2WebACL_RuleGroupReferenceStatement(t *testing.T) {
	var v wafv2.WebACL
	webACLName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_wafv2_web_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSWafv2ScopeRegional(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsWafv2WebACLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsWafv2WebACLConfig_RuleGroupReferenceStatement(webACLName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2WebACLExists(resourceName, &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "name", webACLName),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"name":                      "rule-1",
						"override_action.#":         "1",
						"override_action.0.count.#": "1",
						"override_action.0.none.#":  "0",
						"statement.#":               "1",
						"statement.0.rule_group_reference_statement.#":                 "1",
						"statement.0.rule_group_reference_statement.0.excluded_rule.#": "0",
					}),
					resource.TestMatchTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]*regexp.Regexp{
						"statement.0.rule_group_reference_statement.0.arn": regexp.MustCompile(`regional/rulegroup/.+$`),
					}),
				),
			},
			{
				Config: testAccAwsWafv2WebACLConfig_RuleGroupReferenceStatement_Update(webACLName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2WebACLExists(resourceName, &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "name", webACLName),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"name":                      "rule-1",
						"override_action.#":         "1",
						"override_action.0.count.#": "1",
						"override_action.0.none.#":  "0",
						"statement.#":               "1",
						"statement.0.rule_group_reference_statement.#":                      "1",
						"statement.0.rule_group_reference_statement.0.excluded_rule.#":      "2",
						"statement.0.rule_group_reference_statement.0.excluded_rule.0.name": "rule-to-exclude-b",
						"statement.0.rule_group_reference_statement.0.excluded_rule.1.name": "rule-to-exclude-a",
					}),
					resource.TestMatchTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]*regexp.Regexp{
						"statement.0.rule_group_reference_statement.0.arn": regexp.MustCompile(`regional/rulegroup/.+$`),
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccAwsWafv2WebACLImportStateIdFunc(resourceName),
			},
		},
	})
}

func TestAccAwsWafv2WebACL_Tags(t *testing.T) {
	var v wafv2.WebACL
	webACLName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_wafv2_web_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSWafv2ScopeRegional(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsWafv2WebACLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsWafv2WebACLConfig_OneTag(webACLName, "Tag1", "Value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2WebACLExists(resourceName, &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Tag1", "Value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccAwsWafv2WebACLImportStateIdFunc(resourceName),
			},
			{
				Config: testAccAwsWafv2WebACLConfig_TwoTags(webACLName, "Tag1", "Value1Updated", "Tag2", "Value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2WebACLExists(resourceName, &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.Tag1", "Value1Updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.Tag2", "Value2"),
				),
			},
			{
				Config: testAccAwsWafv2WebACLConfig_OneTag(webACLName, "Tag2", "Value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2WebACLExists(resourceName, &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Tag2", "Value2"),
				),
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/13862
func TestAccAwsWafv2WebACL_MaxNestedRateBasedStatements(t *testing.T) {
	var v wafv2.WebACL
	webACLName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_wafv2_web_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSWafv2ScopeRegional(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsWafv2WebACLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsWafv2WebACLConfig_multipleNestedRateBasedStatements(webACLName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2WebACLExists(resourceName, &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "name", webACLName),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.#":                                                                                                      "1",
						"statement.0.rate_based_statement.#":                                                                               "1",
						"statement.0.rate_based_statement.0.limit":                                                                         "300",
						"statement.0.rate_based_statement.0.aggregate_key_type":                                                            "IP",
						"statement.0.rate_based_statement.0.scope_down_statement.#":                                                        "1",
						"statement.0.rate_based_statement.0.scope_down_statement.0.not_statement.#":                                        "1",
						"statement.0.rate_based_statement.0.scope_down_statement.0.not_statement.0.statement.#":                            "1",
						"statement.0.rate_based_statement.0.scope_down_statement.0.not_statement.0.statement.0.or_statement.#":             "1",
						"statement.0.rate_based_statement.0.scope_down_statement.0.not_statement.0.statement.0.or_statement.0.statement.#": "2",
						"statement.0.rate_based_statement.0.scope_down_statement.0.not_statement.0.statement.0.or_statement.0.statement.0.regex_pattern_set_reference_statement.#": "1",
						"statement.0.rate_based_statement.0.scope_down_statement.0.not_statement.0.statement.0.or_statement.0.statement.1.ip_set_reference_statement.#":            "1",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccAwsWafv2WebACLImportStateIdFunc(resourceName),
			},
		},
	})
}

func TestAccAwsWafv2WebACL_MaxNestedOperatorStatements(t *testing.T) {
	var v wafv2.WebACL
	webACLName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_wafv2_web_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSWafv2ScopeRegional(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsWafv2WebACLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsWafv2WebACLConfig_multipleNestedOperatorStatements(webACLName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2WebACLExists(resourceName, &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "name", webACLName),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.#":                                                                                    "1",
						"statement.0.and_statement.#":                                                                    "1",
						"statement.0.and_statement.0.statement.#":                                                        "2",
						"statement.0.and_statement.0.statement.0.not_statement.#":                                        "1",
						"statement.0.and_statement.0.statement.0.not_statement.0.statement.#":                            "1",
						"statement.0.and_statement.0.statement.0.not_statement.0.statement.0.or_statement.#":             "1",
						"statement.0.and_statement.0.statement.0.not_statement.0.statement.0.or_statement.0.statement.#": "2",
						"statement.0.and_statement.0.statement.0.not_statement.0.statement.0.or_statement.0.statement.0.regex_pattern_set_reference_statement.#": "1",
						"statement.0.and_statement.0.statement.0.not_statement.0.statement.0.or_statement.0.statement.1.ip_set_reference_statement.#":            "1",
						"statement.0.and_statement.0.statement.1.geo_match_statement.#":                                                                          "1",
						"statement.0.and_statement.0.statement.1.geo_match_statement.0.country_codes.0":                                                          "NL",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccAwsWafv2WebACLImportStateIdFunc(resourceName),
			},
		},
	})
}

func testAccCheckAwsWafv2WebACLDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_wafv2_web_acl" {
			continue
		}

		conn := testAccProvider.Meta().(*AWSClient).wafv2conn
		resp, err := conn.GetWebACL(
			&wafv2.GetWebACLInput{
				Id:    aws.String(rs.Primary.ID),
				Name:  aws.String(rs.Primary.Attributes["name"]),
				Scope: aws.String(rs.Primary.Attributes["scope"]),
			})

		if err == nil {
			if resp == nil || resp.WebACL == nil {
				return fmt.Errorf("Error getting WAFv2 WebACL")
			}
			if aws.StringValue(resp.WebACL.Id) == rs.Primary.ID {
				return fmt.Errorf("WAFv2 WebACL %s still exists", rs.Primary.ID)
			}
		}

		// Return nil if the WebACL is already destroyed
		if isAWSErr(err, wafv2.ErrCodeWAFNonexistentItemException, "") {
			return nil
		}

		return err
	}

	return nil
}

func testAccCheckAwsWafv2WebACLExists(n string, v *wafv2.WebACL) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No WAFv2 WebACL ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).wafv2conn
		resp, err := conn.GetWebACL(&wafv2.GetWebACLInput{
			Id:    aws.String(rs.Primary.ID),
			Name:  aws.String(rs.Primary.Attributes["name"]),
			Scope: aws.String(rs.Primary.Attributes["scope"]),
		})

		if err != nil {
			return err
		}

		if resp == nil || resp.WebACL == nil {
			return fmt.Errorf("Error getting WAFv2 WebACL")
		}

		if aws.StringValue(resp.WebACL.Id) == rs.Primary.ID {
			*v = *resp.WebACL
			return nil
		}

		return fmt.Errorf("WAFv2 WebACL (%s) not found", rs.Primary.ID)
	}
}

func testAccAwsWafv2WebACLConfig_Basic(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_web_acl" "test" {
  name        = "%[1]s"
  description = "%[1]s"
  scope       = "REGIONAL"

  default_action {
    allow {}
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
`, name)
}

func testAccAwsWafv2WebACLConfig_BasicRule(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_web_acl" "test" {
  name        = "%[1]s"
  description = "Updated"
  scope       = "REGIONAL"

  default_action {
    block {}
  }

  rule {
    name     = "%[1]s-1"
    priority = 10

    action {
      count {}
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
      metric_name                = "%[1]s-metric-name-1"
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

func testAccAwsWafv2WebACLConfig_UpdateRuleNamePriorityMetric(name, ruleName1, ruleName2 string, priority1, priority2 int) string {
	return fmt.Sprintf(`
resource "aws_wafv2_web_acl" "test" {
  name        = "%[1]s"
  description = "Updated"
  scope       = "REGIONAL"

  default_action {
    block {}
  }

  rule {
    name     = "%[2]s"
    priority = %[3]d

    action {
      count {}
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
      metric_name                = "%[2]s"
      sampled_requests_enabled   = false
    }
  }

  rule {
    name     = "%[4]s"
    priority = %[5]d

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

func testAccAwsWafv2WebACLConfig_GeoMatchStatement(name, countryCodes string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_web_acl" "test" {
  name        = "%[1]s"
  description = "%[1]s"
  scope       = "REGIONAL"

  default_action {
    allow {}
  }

  rule {
    name     = "rule-1"
    priority = 1

    action {
      block {}
    }

    statement {
      geo_match_statement {
        country_codes = [%[2]s]
      }
    }

    visibility_config {
      cloudwatch_metrics_enabled = false
      metric_name                = "friendly-rule-metric-name"
      sampled_requests_enabled   = false
    }
  }

  tags = {
    Tag1 = "Value1"
    Tag2 = "Value2"
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
`, name, countryCodes)
}

func testAccAwsWafv2WebACLConfig_GeoMatchStatement_ForwardedIPConfig(name, fallbackBehavior, headerName string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_web_acl" "test" {
  name        = "%[1]s"
  description = "%[1]s"
  scope       = "REGIONAL"

  default_action {
    block {}
  }

  rule {
    name     = "rule-1"
    priority = 1

    action {
      block {}
    }

    statement {
      or_statement {
        statement {
          geo_match_statement {
            country_codes = ["US"]
          }
        }

        statement {
          geo_match_statement {
            country_codes = ["CA"]
            forwarded_ip_config {
              fallback_behavior = "%s"
              header_name       = "%s"
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
`, name, fallbackBehavior, headerName)
}

func testAccAwsWafv2WebACLConfig_IPSetReferenceStatement(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_ip_set" "test" {
  name               = "ip-set-%[1]s"
  scope              = "REGIONAL"
  ip_address_version = "IPV4"
  addresses          = ["1.1.1.1/32", "2.2.2.2/32"]
}

resource "aws_wafv2_web_acl" "test" {
  name        = "%[1]s"
  description = "%[1]s"
  scope       = "REGIONAL"

  default_action {
    block {}
  }

  rule {
    name     = "rule-1"
    priority = 1

    action {
      block {}
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
`, name)
}

func testAccAwsWafv2WebACLConfig_IPSetReferenceStatement_IPSetForwardedIPConfig(name, fallbackBehavior, headerName, position string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_ip_set" "test" {
  name               = "ip-set-%[1]s"
  scope              = "REGIONAL"
  ip_address_version = "IPV4"
  addresses          = ["1.1.1.1/32", "2.2.2.2/32"]
}

resource "aws_wafv2_web_acl" "test" {
  name        = "%[1]s"
  description = "%[1]s"
  scope       = "REGIONAL"

  default_action {
    block {}
  }

  rule {
    name     = "rule-1"
    priority = 1

    action {
      block {}
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

func testAccAwsWafv2WebACLConfig_ManagedRuleGroupStatement(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_web_acl" "test" {
  name        = "%[1]s"
  description = "%[1]s"
  scope       = "REGIONAL"

  default_action {
    allow {}
  }

  rule {
    name     = "rule-1"
    priority = 1

    override_action {
      none {}
    }

    statement {
      managed_rule_group_statement {
        name        = "AWSManagedRulesCommonRuleSet"
        vendor_name = "AWS"
      }
    }

    visibility_config {
      cloudwatch_metrics_enabled = false
      metric_name                = "friendly-rule-metric-name"
      sampled_requests_enabled   = false
    }
  }

  tags = {
    Tag1 = "Value1"
    Tag2 = "Value2"
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
`, name)
}

func testAccAwsWafv2WebACLConfig_ManagedRuleGroupStatement_Update(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_web_acl" "test" {
  name        = "%[1]s"
  description = "%[1]s"
  scope       = "REGIONAL"

  default_action {
    allow {}
  }

  rule {
    name     = "rule-1"
    priority = 1

    override_action {
      count {}
    }

    statement {
      managed_rule_group_statement {
        name        = "AWSManagedRulesCommonRuleSet"
        vendor_name = "AWS"

        excluded_rule {
          name = "SizeRestrictions_QUERYSTRING"
        }

        excluded_rule {
          name = "NoUserAgent_HEADER"
        }
      }
    }

    visibility_config {
      cloudwatch_metrics_enabled = false
      metric_name                = "friendly-rule-metric-name"
      sampled_requests_enabled   = false
    }
  }

  tags = {
    Tag1 = "Value1"
    Tag2 = "Value2"
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
`, name)
}

func testAccAwsWafv2WebACLConfig_RateBasedStatement(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_web_acl" "test" {
  name        = "%[1]s"
  description = "%[1]s"
  scope       = "REGIONAL"

  default_action {
    block {}
  }

  rule {
    name     = "rule-1"
    priority = 1

    action {
      count {}
    }

    statement {
      rate_based_statement {
        limit = 50000
      }
    }

    visibility_config {
      cloudwatch_metrics_enabled = false
      metric_name                = "friendly-rule-metric-name"
      sampled_requests_enabled   = false
    }
  }

  tags = {
    Tag1 = "Value1"
    Tag2 = "Value2"
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
`, name)
}

func testAccAwsWafv2WebACLConfig_RateBasedStatement_ForwardedIPConfig(name, fallbackBehavior, headerName string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_web_acl" "test" {
  name        = "%[1]s"
  description = "%[1]s"
  scope       = "REGIONAL"

  default_action {
    block {}
  }

  rule {
    name     = "rule-1"
    priority = 1

    action {
      count {}
    }

    statement {
      rate_based_statement {
        aggregate_key_type = "FORWARDED_IP"
        forwarded_ip_config {
          fallback_behavior = "%s"
          header_name       = "%s"
        }
        limit = 50000
      }
    }

    visibility_config {
      cloudwatch_metrics_enabled = false
      metric_name                = "friendly-rule-metric-name"
      sampled_requests_enabled   = false
    }
  }

  tags = {
    Tag1 = "Value1"
    Tag2 = "Value2"
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
`, name, fallbackBehavior, headerName)
}

func testAccAwsWafv2WebACLConfig_RateBasedStatement_Update(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_web_acl" "test" {
  name        = "%[1]s"
  description = "%[1]s"
  scope       = "REGIONAL"

  default_action {
    block {}
  }

  rule {
    name     = "rule-1"
    priority = 1

    action {
      count {}
    }

    statement {
      rate_based_statement {
        limit              = 10000
        aggregate_key_type = "IP"

        scope_down_statement {
          geo_match_statement {
            country_codes = ["US", "NL"]
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

  tags = {
    Tag1 = "Value1"
    Tag2 = "Value2"
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
`, name)
}

func testAccAwsWafv2WebACLConfig_RuleGroupReferenceStatement(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity = 10
  name     = "rule-group-%[1]s"
  scope    = "REGIONAL"

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}

resource "aws_wafv2_web_acl" "test" {
  name  = "%[1]s"
  scope = "REGIONAL"

  default_action {
    block {}
  }

  rule {
    name     = "rule-1"
    priority = 1

    override_action {
      count {}
    }

    statement {
      rule_group_reference_statement {
        arn = aws_wafv2_rule_group.test.arn
      }
    }

    visibility_config {
      cloudwatch_metrics_enabled = false
      metric_name                = "friendly-rule-metric-name"
      sampled_requests_enabled   = false
    }
  }

  tags = {
    Tag1 = "Value1"
    Tag2 = "Value2"
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
`, name)
}

func testAccAwsWafv2WebACLConfig_RuleGroupReferenceStatement_Update(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity = 10
  name     = "rule-group-%[1]s"
  scope    = "REGIONAL"

  rule {
    name     = "rule-1"
    priority = 1

    action {
      count {}
    }

    statement {
      geo_match_statement {
        country_codes = ["NL"]
      }
    }

    visibility_config {
      cloudwatch_metrics_enabled = false
      metric_name                = "friendly-rule-metric-name"
      sampled_requests_enabled   = false
    }
  }

  rule {
    name     = "rule-to-exclude-a"
    priority = 10

    action {
      allow {}
    }

    statement {
      geo_match_statement {
        country_codes = ["US"]
      }
    }

    visibility_config {
      cloudwatch_metrics_enabled = false
      metric_name                = "friendly-rule-metric-name"
      sampled_requests_enabled   = false
    }
  }

  rule {
    name     = "rule-to-exclude-b"
    priority = 15

    action {
      allow {}
    }

    statement {
      geo_match_statement {
        country_codes = ["GB"]
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

resource "aws_wafv2_web_acl" "test" {
  name  = "%[1]s"
  scope = "REGIONAL"

  default_action {
    block {}
  }

  rule {
    name     = "rule-1"
    priority = 1

    override_action {
      count {}
    }

    statement {
      rule_group_reference_statement {
        arn = aws_wafv2_rule_group.test.arn

        excluded_rule {
          name = "rule-to-exclude-b"
        }

        excluded_rule {
          name = "rule-to-exclude-a"
        }
      }
    }

    visibility_config {
      cloudwatch_metrics_enabled = false
      metric_name                = "friendly-rule-metric-name"
      sampled_requests_enabled   = false
    }
  }

  tags = {
    Tag1 = "Value1"
    Tag2 = "Value2"
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
`, name)
}

func testAccAwsWafv2WebACLConfig_Minimal(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_web_acl" "test" {
  name  = "%s"
  scope = "REGIONAL"

  default_action {
    allow {}
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
`, name)
}

func testAccAwsWafv2WebACLConfig_OneTag(name, tagKey, tagValue string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_web_acl" "test" {
  name        = "%[1]s"
  description = "%[1]s"
  scope       = "REGIONAL"

  default_action {
    allow {}
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }

  tags = {
    "%s" = "%s"
  }
}
`, name, tagKey, tagValue)
}

func testAccAwsWafv2WebACLConfig_TwoTags(name, tag1Key, tag1Value, tag2Key, tag2Value string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_web_acl" "test" {
  name        = "%[1]s"
  description = "%[1]s"
  scope       = "REGIONAL"

  default_action {
    allow {}
  }

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
`, name, tag1Key, tag1Value, tag2Key, tag2Value)
}

func testAccAwsWafv2WebACLConfig_multipleNestedRateBasedStatements(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_regex_pattern_set" "test" {
  name  = "%[1]s"
  scope = "REGIONAL"

  regular_expression {
    regex_string = "one"
  }
}

resource "aws_wafv2_ip_set" "test" {
  name               = "%[1]s"
  scope              = "REGIONAL"
  ip_address_version = "IPV4"
  addresses          = ["1.2.3.4/32", "5.6.7.8/32"]
}

resource "aws_wafv2_web_acl" "test" {
  name        = "%[1]s"
  description = "%[1]s"
  scope       = "REGIONAL"

  default_action {
    allow {}
  }

  rule {
    name     = "rule"
    priority = 0

    action {
      block {}
    }

    statement {
      rate_based_statement {
        limit              = 300
        aggregate_key_type = "IP"

        scope_down_statement {
          not_statement {
            statement {
              or_statement {
                statement {
                  regex_pattern_set_reference_statement {
                    arn = aws_wafv2_regex_pattern_set.test.arn

                    field_to_match {
                      uri_path {}
                    }

                    text_transformation {
                      type     = "LOWERCASE"
                      priority = 1
                    }
                  }
                }

                statement {
                  ip_set_reference_statement {
                    arn = aws_wafv2_ip_set.test.arn
                  }
                }
              }
            }
          }
        }
      }
    }

    visibility_config {
      cloudwatch_metrics_enabled = false
      metric_name                = "rule"
      sampled_requests_enabled   = false
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "waf"
    sampled_requests_enabled   = false
  }
}
`, name)
}

func testAccAwsWafv2WebACLConfig_multipleNestedOperatorStatements(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_regex_pattern_set" "test" {
  name  = "%[1]s"
  scope = "REGIONAL"

  regular_expression {
    regex_string = "one"
  }
}

resource "aws_wafv2_ip_set" "test" {
  name               = "%[1]s"
  scope              = "REGIONAL"
  ip_address_version = "IPV4"
  addresses          = ["1.2.3.4/32", "5.6.7.8/32"]
}

resource "aws_wafv2_web_acl" "test" {
  name        = "%[1]s"
  description = "%[1]s"
  scope       = "REGIONAL"

  default_action {
    allow {}
  }

  rule {
    name     = "rule"
    priority = 0

    action {
      block {}
    }

    statement {
      and_statement {
        statement {
          not_statement {
            statement {
              or_statement {
                statement {
                  regex_pattern_set_reference_statement {
                    arn = aws_wafv2_regex_pattern_set.test.arn

                    field_to_match {
                      uri_path {}
                    }

                    text_transformation {
                      type     = "LOWERCASE"
                      priority = 1
                    }
                  }
                }

                statement {
                  ip_set_reference_statement {
                    arn = aws_wafv2_ip_set.test.arn
                  }
                }
              }
            }
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
      metric_name                = "rule"
      sampled_requests_enabled   = false
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "waf"
    sampled_requests_enabled   = false
  }
}
`, name)
}

func testAccAwsWafv2WebACLImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return fmt.Sprintf("%s/%s/%s", rs.Primary.ID, rs.Primary.Attributes["name"], rs.Primary.Attributes["scope"]), nil
	}
}
