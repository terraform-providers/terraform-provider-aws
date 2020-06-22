package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/wafv2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccAwsWafv2RuleGroup_Basic(t *testing.T) {
	var v wafv2.RuleGroup
	ruleGroupName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_wafv2_rule_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
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
					resource.TestCheckResourceAttr(resourceName, "rule.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "rule.1162901930.name", "rule-2"),
					resource.TestCheckResourceAttr(resourceName, "rule.1162901930.priority", "10"),
					resource.TestCheckResourceAttr(resourceName, "rule.1162901930.action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.1162901930.action.0.allow.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "rule.1162901930.action.0.block.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.1162901930.action.0.count.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "rule.1162901930.statement.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.1162901930.statement.0.size_constraint_statement.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.1162901930.statement.0.size_constraint_statement.0.comparison_operator", "LT"),
					resource.TestCheckResourceAttr(resourceName, "rule.1162901930.statement.0.size_constraint_statement.0.field_to_match.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.1162901930.statement.0.size_constraint_statement.0.field_to_match.0.query_string.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.1162901930.statement.0.size_constraint_statement.0.size", "50"),
					resource.TestCheckResourceAttr(resourceName, "rule.1162901930.statement.0.size_constraint_statement.0.text_transformation.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "rule.1162901930.statement.0.size_constraint_statement.0.text_transformation.2212084700.priority", "2"),
					resource.TestCheckResourceAttr(resourceName, "rule.1162901930.statement.0.size_constraint_statement.0.text_transformation.2212084700.type", "CMD_LINE"),
					resource.TestCheckResourceAttr(resourceName, "rule.1162901930.statement.0.size_constraint_statement.0.text_transformation.4292479961.priority", "5"),
					resource.TestCheckResourceAttr(resourceName, "rule.1162901930.statement.0.size_constraint_statement.0.text_transformation.4292479961.type", "NONE"),
					resource.TestCheckResourceAttr(resourceName, "rule.3400404505.name", "rule-1"),
					resource.TestCheckResourceAttr(resourceName, "rule.3400404505.priority", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.3400404505.action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.3400404505.action.0.allow.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "rule.3400404505.action.0.block.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "rule.3400404505.action.0.count.#", "1"),
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
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsWafv2RuleGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsWafv2RuleGroupConfig_ByteMatchStatement(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2RuleGroupExists(resourceName, &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.3451531458.statement.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.3451531458.statement.0.byte_match_statement.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.3451531458.statement.0.byte_match_statement.0.positional_constraint", "CONTAINS"),
					resource.TestCheckResourceAttr(resourceName, "rule.3451531458.statement.0.byte_match_statement.0.search_string", "word"),
					resource.TestCheckResourceAttr(resourceName, "rule.3451531458.statement.0.byte_match_statement.0.text_transformation.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "rule.3451531458.statement.0.byte_match_statement.0.text_transformation.4292479961.priority", "5"),
					resource.TestCheckResourceAttr(resourceName, "rule.3451531458.statement.0.byte_match_statement.0.text_transformation.4292479961.type", "NONE"),
					resource.TestCheckResourceAttr(resourceName, "rule.3451531458.statement.0.byte_match_statement.0.text_transformation.2156930824.priority", "2"),
					resource.TestCheckResourceAttr(resourceName, "rule.3451531458.statement.0.byte_match_statement.0.text_transformation.2156930824.type", "LOWERCASE"),
				),
			},
			{
				Config: testAccAwsWafv2RuleGroupConfig_ByteMatchStatement_Update(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2RuleGroupExists(resourceName, &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.3446749900.statement.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.3446749900.statement.0.byte_match_statement.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.3446749900.statement.0.byte_match_statement.0.positional_constraint", "EXACTLY"),
					resource.TestCheckResourceAttr(resourceName, "rule.3446749900.statement.0.byte_match_statement.0.search_string", "sentence"),
					resource.TestCheckResourceAttr(resourceName, "rule.3446749900.statement.0.byte_match_statement.0.text_transformation.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.3446749900.statement.0.byte_match_statement.0.text_transformation.766585421.priority", "3"),
					resource.TestCheckResourceAttr(resourceName, "rule.3446749900.statement.0.byte_match_statement.0.text_transformation.766585421.type", "CMD_LINE"),
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
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsWafv2RuleGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsWafv2RuleGroupConfig_ByteMatchStatement_FieldToMatchAllQueryArguments(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2RuleGroupExists(resourceName, &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.2971982489.statement.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.2971982489.statement.0.byte_match_statement.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.2971982489.statement.0.byte_match_statement.0.field_to_match.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.2971982489.statement.0.byte_match_statement.0.field_to_match.0.all_query_arguments.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.2971982489.statement.0.byte_match_statement.0.field_to_match.0.body.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "rule.2971982489.statement.0.byte_match_statement.0.field_to_match.0.method.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "rule.2971982489.statement.0.byte_match_statement.0.field_to_match.0.query_string.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "rule.2971982489.statement.0.byte_match_statement.0.field_to_match.0.single_header.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "rule.2971982489.statement.0.byte_match_statement.0.field_to_match.0.single_query_argument.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "rule.2971982489.statement.0.byte_match_statement.0.field_to_match.0.uri_path.#", "0"),
				),
			},
			{
				Config: testAccAwsWafv2RuleGroupConfig_ByteMatchStatement_FieldToMatchBody(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2RuleGroupExists(resourceName, &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.4243272354.statement.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.4243272354.statement.0.byte_match_statement.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.4243272354.statement.0.byte_match_statement.0.field_to_match.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.4243272354.statement.0.byte_match_statement.0.field_to_match.0.all_query_arguments.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "rule.4243272354.statement.0.byte_match_statement.0.field_to_match.0.body.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.4243272354.statement.0.byte_match_statement.0.field_to_match.0.method.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "rule.4243272354.statement.0.byte_match_statement.0.field_to_match.0.query_string.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "rule.4243272354.statement.0.byte_match_statement.0.field_to_match.0.single_header.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "rule.4243272354.statement.0.byte_match_statement.0.field_to_match.0.single_query_argument.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "rule.4243272354.statement.0.byte_match_statement.0.field_to_match.0.uri_path.#", "0"),
				),
			},
			{
				Config: testAccAwsWafv2RuleGroupConfig_ByteMatchStatement_FieldToMatchMethod(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2RuleGroupExists(resourceName, &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.1903944126.statement.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.1903944126.statement.0.byte_match_statement.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.1903944126.statement.0.byte_match_statement.0.field_to_match.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.1903944126.statement.0.byte_match_statement.0.field_to_match.0.all_query_arguments.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "rule.1903944126.statement.0.byte_match_statement.0.field_to_match.0.body.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "rule.1903944126.statement.0.byte_match_statement.0.field_to_match.0.method.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.1903944126.statement.0.byte_match_statement.0.field_to_match.0.query_string.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "rule.1903944126.statement.0.byte_match_statement.0.field_to_match.0.single_header.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "rule.1903944126.statement.0.byte_match_statement.0.field_to_match.0.single_query_argument.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "rule.1903944126.statement.0.byte_match_statement.0.field_to_match.0.uri_path.#", "0"),
				),
			},
			{
				Config: testAccAwsWafv2RuleGroupConfig_ByteMatchStatement_FieldToMatchQueryString(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2RuleGroupExists(resourceName, &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.95492546.statement.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.95492546.statement.0.byte_match_statement.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.95492546.statement.0.byte_match_statement.0.field_to_match.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.95492546.statement.0.byte_match_statement.0.field_to_match.0.all_query_arguments.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "rule.95492546.statement.0.byte_match_statement.0.field_to_match.0.body.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "rule.95492546.statement.0.byte_match_statement.0.field_to_match.0.method.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "rule.95492546.statement.0.byte_match_statement.0.field_to_match.0.query_string.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.95492546.statement.0.byte_match_statement.0.field_to_match.0.single_header.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "rule.95492546.statement.0.byte_match_statement.0.field_to_match.0.single_query_argument.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "rule.95492546.statement.0.byte_match_statement.0.field_to_match.0.uri_path.#", "0"),
				),
			},
			{
				Config: testAccAwsWafv2RuleGroupConfig_ByteMatchStatement_FieldToMatchSingleHeader(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2RuleGroupExists(resourceName, &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.2127391528.statement.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.2127391528.statement.0.byte_match_statement.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.2127391528.statement.0.byte_match_statement.0.field_to_match.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.2127391528.statement.0.byte_match_statement.0.field_to_match.0.all_query_arguments.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "rule.2127391528.statement.0.byte_match_statement.0.field_to_match.0.body.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "rule.2127391528.statement.0.byte_match_statement.0.field_to_match.0.method.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "rule.2127391528.statement.0.byte_match_statement.0.field_to_match.0.query_string.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "rule.2127391528.statement.0.byte_match_statement.0.field_to_match.0.single_header.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.2127391528.statement.0.byte_match_statement.0.field_to_match.0.single_header.0.name", "a-forty-character-long-header-name-40-40"),
					resource.TestCheckResourceAttr(resourceName, "rule.2127391528.statement.0.byte_match_statement.0.field_to_match.0.single_query_argument.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "rule.2127391528.statement.0.byte_match_statement.0.field_to_match.0.uri_path.#", "0"),
				),
			},
			{
				Config: testAccAwsWafv2RuleGroupConfig_ByteMatchStatement_FieldToMatchSingleQueryArgument(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2RuleGroupExists(resourceName, &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.492520910.statement.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.492520910.statement.0.byte_match_statement.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.492520910.statement.0.byte_match_statement.0.field_to_match.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.492520910.statement.0.byte_match_statement.0.field_to_match.0.all_query_arguments.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "rule.492520910.statement.0.byte_match_statement.0.field_to_match.0.body.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "rule.492520910.statement.0.byte_match_statement.0.field_to_match.0.method.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "rule.492520910.statement.0.byte_match_statement.0.field_to_match.0.query_string.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "rule.492520910.statement.0.byte_match_statement.0.field_to_match.0.single_header.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "rule.492520910.statement.0.byte_match_statement.0.field_to_match.0.single_query_argument.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.492520910.statement.0.byte_match_statement.0.field_to_match.0.single_query_argument.0.name", "a-max-30-characters-long-name-"),
					resource.TestCheckResourceAttr(resourceName, "rule.492520910.statement.0.byte_match_statement.0.field_to_match.0.uri_path.#", "0"),
				),
			},
			{
				Config: testAccAwsWafv2RuleGroupConfig_ByteMatchStatement_FieldToMatchUriPath(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2RuleGroupExists(resourceName, &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.984121555.statement.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.984121555.statement.0.byte_match_statement.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.984121555.statement.0.byte_match_statement.0.field_to_match.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.984121555.statement.0.byte_match_statement.0.field_to_match.0.all_query_arguments.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "rule.984121555.statement.0.byte_match_statement.0.field_to_match.0.body.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "rule.984121555.statement.0.byte_match_statement.0.field_to_match.0.method.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "rule.984121555.statement.0.byte_match_statement.0.field_to_match.0.query_string.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "rule.984121555.statement.0.byte_match_statement.0.field_to_match.0.single_header.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "rule.984121555.statement.0.byte_match_statement.0.field_to_match.0.single_query_argument.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "rule.984121555.statement.0.byte_match_statement.0.field_to_match.0.uri_path.#", "1"),
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
		PreCheck:     func() { testAccPreCheck(t) },
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
		PreCheck:     func() { testAccPreCheck(t) },
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
		PreCheck:     func() { testAccPreCheck(t) },
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
		PreCheck:     func() { testAccPreCheck(t) },
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
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsWafv2RuleGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsWafv2RuleGroupConfig_GeoMatchStatement(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2RuleGroupExists(resourceName, &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.2064993648.statement.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.2064993648.statement.0.geo_match_statement.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.2064993648.statement.0.geo_match_statement.0.country_codes.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "rule.2064993648.statement.0.geo_match_statement.0.country_codes.0", "US"),
					resource.TestCheckResourceAttr(resourceName, "rule.2064993648.statement.0.geo_match_statement.0.country_codes.1", "NL"),
				),
			},
			{
				Config: testAccAwsWafv2RuleGroupConfig_GeoMatchStatement_Update(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2RuleGroupExists(resourceName, &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.4215509823.statement.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.4215509823.statement.0.geo_match_statement.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.4215509823.statement.0.geo_match_statement.0.country_codes.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "rule.4215509823.statement.0.geo_match_statement.0.country_codes.0", "ZM"),
					resource.TestCheckResourceAttr(resourceName, "rule.4215509823.statement.0.geo_match_statement.0.country_codes.1", "EE"),
					resource.TestCheckResourceAttr(resourceName, "rule.4215509823.statement.0.geo_match_statement.0.country_codes.2", "MM"),
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
	var idx int
	ruleGroupName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_wafv2_rule_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsWafv2RuleGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsWafv2RuleGroupConfig_IpSetReferenceStatement(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2RuleGroupExists(resourceName, &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					computeWafv2IpSetRefStatementIndex(&v, &idx),
					testCheckResourceAttrWithIndexesAddr(resourceName, "rule.%d.statement.#", &idx, "1"),
					testCheckResourceAttrWithIndexesAddr(resourceName, "rule.%d.statement.0.ip_set_reference_statement.#", &idx, "1"),
					testAccMatchResourceAttrArnWithIndexesAddr(resourceName, "rule.%d.statement.0.ip_set_reference_statement.0.arn", &idx, regexp.MustCompile(`regional/ipset/.+$`)),
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
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsWafv2RuleGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsWafv2RuleGroupConfig_LogicalRuleStatement_And(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2RuleGroupExists(resourceName, &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.867936606.statement.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.867936606.statement.0.and_statement.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.867936606.statement.0.and_statement.0.statement.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "rule.867936606.statement.0.and_statement.0.statement.0.geo_match_statement.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.867936606.statement.0.and_statement.0.statement.1.geo_match_statement.#", "1"),
				),
			},
			{
				Config: testAccAwsWafv2RuleGroupConfig_LogicalRuleStatement_NotAnd(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2RuleGroupExists(resourceName, &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.2966210839.statement.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.2966210839.statement.0.not_statement.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.2966210839.statement.0.not_statement.0.statement.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.2966210839.statement.0.not_statement.0.statement.0.and_statement.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.2966210839.statement.0.not_statement.0.statement.0.and_statement.0.statement.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "rule.2966210839.statement.0.not_statement.0.statement.0.and_statement.0.statement.0.geo_match_statement.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.2966210839.statement.0.not_statement.0.statement.0.and_statement.0.statement.1.geo_match_statement.#", "1"),
				),
			},
			{
				Config: testAccAwsWafv2RuleGroupConfig_LogicalRuleStatement_OrNotAnd(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2RuleGroupExists(resourceName, &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.464720012.statement.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.464720012.statement.0.or_statement.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.464720012.statement.0.or_statement.0.statement.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "rule.464720012.statement.0.or_statement.0.statement.0.not_statement.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.464720012.statement.0.or_statement.0.statement.0.not_statement.0.statement.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.464720012.statement.0.or_statement.0.statement.0.not_statement.0.statement.0.geo_match_statement.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.464720012.statement.0.or_statement.0.statement.1.and_statement.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.464720012.statement.0.or_statement.0.statement.1.and_statement.0.statement.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "rule.464720012.statement.0.or_statement.0.statement.1.and_statement.0.statement.0.geo_match_statement.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.464720012.statement.0.or_statement.0.statement.1.and_statement.0.statement.1.geo_match_statement.#", "1"),
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
		PreCheck:     func() { testAccPreCheck(t) },
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
	var idx int
	ruleGroupName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_wafv2_rule_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsWafv2RuleGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsWafv2RuleGroupConfig_RegexPatternSetReferenceStatement(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2RuleGroupExists(resourceName, &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					computeWafv2RegexSetRefStatementIndex(&v, &idx),
					testCheckResourceAttrWithIndexesAddr(resourceName, "rule.%d.statement.#", &idx, "1"),
					testCheckResourceAttrWithIndexesAddr(resourceName, "rule.%d.statement.0.regex_pattern_set_reference_statement.#", &idx, "1"),
					testCheckResourceAttrWithIndexesAddr(resourceName, "rule.%d.statement.0.regex_pattern_set_reference_statement.0.field_to_match.#", &idx, "1"),
					testCheckResourceAttrWithIndexesAddr(resourceName, "rule.%d.statement.0.regex_pattern_set_reference_statement.0.text_transformation.#", &idx, "1"),
					testAccMatchResourceAttrArnWithIndexesAddr(resourceName, "rule.%d.statement.0.regex_pattern_set_reference_statement.0.arn", &idx, regexp.MustCompile(`regional/regexpatternset/.+$`)),
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
		PreCheck:     func() { testAccPreCheck(t) },
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
					resource.TestCheckResourceAttr(resourceName, "rule.2064993648.action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.2064993648.action.0.allow.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.2064993648.action.0.block.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "rule.2064993648.action.0.count.#", "0"),
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
					resource.TestCheckResourceAttr(resourceName, "rule.2490412960.action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.2490412960.action.0.allow.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "rule.2490412960.action.0.block.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.2490412960.action.0.count.#", "0"),
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
					resource.TestCheckResourceAttr(resourceName, "rule.3400404505.action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.3400404505.action.0.allow.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "rule.3400404505.action.0.block.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "rule.3400404505.action.0.count.#", "1"),
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
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsWafv2RuleGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsWafv2RuleGroupConfig_SizeConstraintStatement(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2RuleGroupExists(resourceName, &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.204038473.statement.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.204038473.statement.0.size_constraint_statement.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.204038473.statement.0.size_constraint_statement.0.comparison_operator", "GT"),
					resource.TestCheckResourceAttr(resourceName, "rule.204038473.statement.0.size_constraint_statement.0.size", "100"),
					resource.TestCheckResourceAttr(resourceName, "rule.204038473.statement.0.size_constraint_statement.0.field_to_match.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.204038473.statement.0.size_constraint_statement.0.field_to_match.0.method.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.204038473.statement.0.size_constraint_statement.0.text_transformation.#", "1"),
				),
			},
			{
				Config: testAccAwsWafv2RuleGroupConfig_SizeConstraintStatement_Update(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2RuleGroupExists(resourceName, &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.3233602303.statement.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.3233602303.statement.0.size_constraint_statement.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.3233602303.statement.0.size_constraint_statement.0.comparison_operator", "LT"),
					resource.TestCheckResourceAttr(resourceName, "rule.3233602303.statement.0.size_constraint_statement.0.size", "50"),
					resource.TestCheckResourceAttr(resourceName, "rule.3233602303.statement.0.size_constraint_statement.0.field_to_match.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.3233602303.statement.0.size_constraint_statement.0.field_to_match.0.query_string.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.3233602303.statement.0.size_constraint_statement.0.text_transformation.#", "2"),
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
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsWafv2RuleGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsWafv2RuleGroupConfig_SqliMatchStatement(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2RuleGroupExists(resourceName, &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.1169024249.statement.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.1169024249.statement.0.sqli_match_statement.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.1169024249.statement.0.sqli_match_statement.0.field_to_match.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.1169024249.statement.0.sqli_match_statement.0.field_to_match.0.all_query_arguments.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.1169024249.statement.0.sqli_match_statement.0.text_transformation.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "rule.1169024249.statement.0.sqli_match_statement.0.text_transformation.1247795808.priority", "5"),
					resource.TestCheckResourceAttr(resourceName, "rule.1169024249.statement.0.sqli_match_statement.0.text_transformation.1247795808.type", "URL_DECODE"),
					resource.TestCheckResourceAttr(resourceName, "rule.1169024249.statement.0.sqli_match_statement.0.text_transformation.2156930824.priority", "2"),
					resource.TestCheckResourceAttr(resourceName, "rule.1169024249.statement.0.sqli_match_statement.0.text_transformation.2156930824.type", "LOWERCASE"),
				),
			},
			{
				Config: testAccAwsWafv2RuleGroupConfig_SqliMatchStatement_Update(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2RuleGroupExists(resourceName, &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.2934928563.statement.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.2934928563.statement.0.sqli_match_statement.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.2934928563.statement.0.sqli_match_statement.0.field_to_match.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.2934928563.statement.0.sqli_match_statement.0.field_to_match.0.body.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.2934928563.statement.0.sqli_match_statement.0.text_transformation.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "rule.2934928563.statement.0.sqli_match_statement.0.text_transformation.1247795808.priority", "5"),
					resource.TestCheckResourceAttr(resourceName, "rule.2934928563.statement.0.sqli_match_statement.0.text_transformation.1247795808.type", "URL_DECODE"),
					resource.TestCheckResourceAttr(resourceName, "rule.2934928563.statement.0.sqli_match_statement.0.text_transformation.4120379776.priority", "4"),
					resource.TestCheckResourceAttr(resourceName, "rule.2934928563.statement.0.sqli_match_statement.0.text_transformation.4120379776.type", "HTML_ENTITY_DECODE"),
					resource.TestCheckResourceAttr(resourceName, "rule.2934928563.statement.0.sqli_match_statement.0.text_transformation.1521630275.priority", "3"),
					resource.TestCheckResourceAttr(resourceName, "rule.2934928563.statement.0.sqli_match_statement.0.text_transformation.1521630275.type", "COMPRESS_WHITE_SPACE"),
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
		PreCheck:     func() { testAccPreCheck(t) },
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
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsWafv2RuleGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsWafv2RuleGroupConfig_XssMatchStatement(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2RuleGroupExists(resourceName, &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.377378487.statement.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.377378487.statement.0.xss_match_statement.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.377378487.statement.0.xss_match_statement.0.field_to_match.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.377378487.statement.0.xss_match_statement.0.field_to_match.0.body.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.377378487.statement.0.xss_match_statement.0.text_transformation.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.377378487.statement.0.xss_match_statement.0.text_transformation.2336416342.priority", "2"),
					resource.TestCheckResourceAttr(resourceName, "rule.377378487.statement.0.xss_match_statement.0.text_transformation.2336416342.type", "NONE"),
				),
			},
			{
				Config: testAccAwsWafv2RuleGroupConfig_XssMatchStatement_Update(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2RuleGroupExists(resourceName, &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.3452423088.statement.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.3452423088.statement.0.xss_match_statement.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.3452423088.statement.0.xss_match_statement.0.field_to_match.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.3452423088.statement.0.xss_match_statement.0.field_to_match.0.body.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.3452423088.statement.0.xss_match_statement.0.text_transformation.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.3452423088.statement.0.xss_match_statement.0.text_transformation.2876362756.priority", "2"),
					resource.TestCheckResourceAttr(resourceName, "rule.3452423088.statement.0.xss_match_statement.0.text_transformation.2876362756.type", "URL_DECODE"),
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

func testAccMatchResourceAttrArnWithIndexesAddr(name, format string, idx *int, arnResourceRegexp *regexp.Regexp) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		return testAccMatchResourceAttrRegionalARN(name, fmt.Sprintf(format, *idx), "wafv2", arnResourceRegexp)(s)
	}
}

// Calculates the index which isn't static because ARN is generated as part of the test
func computeWafv2IpSetRefStatementIndex(r *wafv2.RuleGroup, idx *int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ruleResource := resourceAwsWafv2RuleGroup().Schema["rule"].Elem.(*schema.Resource)

		rule := map[string]interface{}{
			"name":     "rule-1",
			"priority": 1,
			"action": []interface{}{
				map[string]interface{}{
					"allow": make([]interface{}, 1),
					"block": []interface{}{},
					"count": []interface{}{},
				},
			},
			"statement": []interface{}{
				map[string]interface{}{
					"and_statement":        []interface{}{},
					"byte_match_statement": []interface{}{},
					"geo_match_statement":  []interface{}{},
					"ip_set_reference_statement": []interface{}{
						map[string]interface{}{
							"arn": aws.StringValue(r.Rules[0].Statement.IPSetReferenceStatement.ARN),
						},
					},
					"not_statement":                         []interface{}{},
					"or_statement":                          []interface{}{},
					"regex_pattern_set_reference_statement": []interface{}{},
					"size_constraint_statement":             []interface{}{},
					"sqli_match_statement":                  []interface{}{},
					"xss_match_statement":                   []interface{}{},
				},
			},
			"visibility_config": []interface{}{
				map[string]interface{}{
					"cloudwatch_metrics_enabled": false,
					"metric_name":                "friendly-rule-metric-name",
					"sampled_requests_enabled":   false,
				},
			},
		}

		f := schema.HashResource(ruleResource)
		*idx = f(rule)

		return nil
	}
}

// Calculates the index which isn't static because ARN is generated as part of the test
func computeWafv2RegexSetRefStatementIndex(r *wafv2.RuleGroup, idx *int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ruleResource := resourceAwsWafv2RuleGroup().Schema["rule"].Elem.(*schema.Resource)
		textTransformationResource := ruleResource.Schema["statement"].Elem.(*schema.Resource).Schema["regex_pattern_set_reference_statement"].Elem.(*schema.Resource).Schema["text_transformation"].Elem.(*schema.Resource)
		rule := map[string]interface{}{
			"name":     "rule-1",
			"priority": 1,
			"action": []interface{}{
				map[string]interface{}{
					"allow": make([]interface{}, 1),
					"block": []interface{}{},
					"count": []interface{}{},
				},
			},
			"statement": []interface{}{
				map[string]interface{}{
					"and_statement":              []interface{}{},
					"byte_match_statement":       []interface{}{},
					"geo_match_statement":        []interface{}{},
					"ip_set_reference_statement": []interface{}{},
					"not_statement":              []interface{}{},
					"or_statement":               []interface{}{},
					"regex_pattern_set_reference_statement": []interface{}{
						map[string]interface{}{
							"arn": aws.StringValue(r.Rules[0].Statement.RegexPatternSetReferenceStatement.ARN),
							"field_to_match": []interface{}{
								map[string]interface{}{
									"all_query_arguments":   []interface{}{},
									"body":                  make([]interface{}, 1),
									"method":                []interface{}{},
									"query_string":          []interface{}{},
									"single_header":         []interface{}{},
									"single_query_argument": []interface{}{},
									"uri_path":              []interface{}{},
								},
							},
							"text_transformation": schema.NewSet(schema.HashResource(textTransformationResource), []interface{}{
								map[string]interface{}{
									"priority": 2,
									"type":     "NONE",
								},
							}),
						},
					},
					"size_constraint_statement": []interface{}{},
					"sqli_match_statement":      []interface{}{},
					"xss_match_statement":       []interface{}{},
				},
			},
			"visibility_config": []interface{}{
				map[string]interface{}{
					"cloudwatch_metrics_enabled": false,
					"metric_name":                "friendly-rule-metric-name",
					"sampled_requests_enabled":   false,
				},
			},
		}

		f := schema.HashResource(ruleResource)
		*idx = f(rule)

		return nil
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
    name     = "rule-2"
    priority = 10

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
      metric_name                = "friendly-rule-metric-name"
      sampled_requests_enabled   = false
    }
  }

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
