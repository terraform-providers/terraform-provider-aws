package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/waf"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfawsresource"
)

func TestAccAWSWafRateBasedRule_basic(t *testing.T) {
	var v waf.RateBasedRule
	wafRuleName := fmt.Sprintf("wafrule%s", acctest.RandString(5))
	resourceName := "aws_waf_rate_based_rule.wafrule"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSWaf(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafRateBasedRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafRateBasedRuleConfig(wafRuleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafRateBasedRuleExists(resourceName, &v),
					testAccMatchResourceAttrGlobalARN(resourceName, "arn", "waf", regexp.MustCompile(`ratebasedrule/.+`)),
					resource.TestCheckResourceAttr(resourceName, "name", wafRuleName),
					resource.TestCheckResourceAttr(resourceName, "predicates.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "metric_name", wafRuleName),
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

func TestAccAWSWafRateBasedRule_changeNameForceNew(t *testing.T) {
	var before, after waf.RateBasedRule
	wafRuleName := fmt.Sprintf("wafrule%s", acctest.RandString(5))
	wafRuleNewName := fmt.Sprintf("wafrulenew%s", acctest.RandString(5))
	resourceName := "aws_waf_rate_based_rule.wafrule"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSWaf(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafIPSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafRateBasedRuleConfig(wafRuleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafRateBasedRuleExists(resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, "name", wafRuleName),
					resource.TestCheckResourceAttr(resourceName, "predicates.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "metric_name", wafRuleName),
				),
			},
			{
				Config: testAccAWSWafRateBasedRuleConfigChangeName(wafRuleNewName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafRateBasedRuleExists(resourceName, &after),
					resource.TestCheckResourceAttr(resourceName, "name", wafRuleNewName),
					resource.TestCheckResourceAttr(resourceName, "predicates.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "metric_name", wafRuleNewName),
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

func TestAccAWSWafRateBasedRule_disappears(t *testing.T) {
	var v waf.RateBasedRule
	wafRuleName := fmt.Sprintf("wafrule%s", acctest.RandString(5))
	resourceName := "aws_waf_rate_based_rule.wafrule"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSWaf(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafRateBasedRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafRateBasedRuleConfig(wafRuleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafRateBasedRuleExists(resourceName, &v),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsWafRateBasedRule(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSWafRateBasedRule_changePredicates(t *testing.T) {
	var ipset waf.IPSet
	var byteMatchSet waf.ByteMatchSet

	var before, after waf.RateBasedRule
	ruleName := fmt.Sprintf("wafrule%s", acctest.RandString(5))
	resourceName := "aws_waf_rate_based_rule.wafrule"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSWaf(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafRateBasedRuleConfig(ruleName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSWafIPSetExists("aws_waf_ipset.ipset", &ipset),
					testAccCheckAWSWafRateBasedRuleExists(resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, "name", ruleName),
					resource.TestCheckResourceAttr(resourceName, "predicates.#", "1"),
					tfawsresource.TestCheckTypeSetElemNestedAttrs(resourceName, "predicates.*", map[string]string{
						"negated": "false",
						"type":    "IPMatch",
					}),
				),
			},
			{
				Config: testAccAWSWafRateBasedRuleConfig_changePredicates(ruleName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSWafByteMatchSetExists("aws_waf_byte_match_set.set", &byteMatchSet),
					testAccCheckAWSWafRateBasedRuleExists(resourceName, &after),
					resource.TestCheckResourceAttr(resourceName, "name", ruleName),
					resource.TestCheckResourceAttr(resourceName, "predicates.#", "1"),
					tfawsresource.TestCheckTypeSetElemNestedAttrs(resourceName, "predicates.*", map[string]string{
						"negated": "true",
						"type":    "ByteMatch",
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

// Reference: https://github.com/terraform-providers/terraform-provider-aws/issues/9659
func TestAccAWSWafRateBasedRule_changeRateLimit(t *testing.T) {
	var ipset waf.IPSet
	var before, after waf.RateBasedRule
	ruleName := fmt.Sprintf("wafrule%s", acctest.RandString(5))
	resourceName := "aws_waf_rate_based_rule.wafrule"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSWaf(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafRateBasedRuleConfig_changeRateLimit(ruleName, 4000),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSWafIPSetExists("aws_waf_ipset.ipset", &ipset),
					testAccCheckAWSWafRateBasedRuleExists(resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, "name", ruleName),
					resource.TestCheckResourceAttr(resourceName, "rate_limit", "4000"),
					resource.TestCheckResourceAttr(resourceName, "predicates.#", "1"),
					tfawsresource.TestCheckTypeSetElemNestedAttrs(resourceName, "predicates.*", map[string]string{
						"negated": "false",
						"type":    "IPMatch",
					}),
				),
			},
			{
				Config: testAccAWSWafRateBasedRuleConfig_changeRateLimit(ruleName, 3000),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSWafIPSetExists("aws_waf_ipset.ipset", &ipset),
					testAccCheckAWSWafRateBasedRuleExists(resourceName, &after),
					resource.TestCheckResourceAttr(resourceName, "name", ruleName),
					resource.TestCheckResourceAttr(resourceName, "rate_limit", "3000"),
					resource.TestCheckResourceAttr(resourceName, "predicates.#", "1"),
					tfawsresource.TestCheckTypeSetElemNestedAttrs(resourceName, "predicates.*", map[string]string{
						"negated": "false",
						"type":    "IPMatch",
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

func TestAccAWSWafRateBasedRule_noPredicates(t *testing.T) {
	var rule waf.RateBasedRule
	ruleName := fmt.Sprintf("wafrule%s", acctest.RandString(5))
	resourceName := "aws_waf_rate_based_rule.wafrule"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSWaf(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafRateBasedRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafRateBasedRuleConfig_noPredicates(ruleName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSWafRateBasedRuleExists(resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "name", ruleName),
					resource.TestCheckResourceAttr(resourceName, "predicates.#", "0"),
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

func TestAccAWSWafRateBasedRule_Tags(t *testing.T) {
	var rule waf.RateBasedRule
	ruleName := fmt.Sprintf("wafrule%s", acctest.RandString(5))
	resourceName := "aws_waf_rate_based_rule.wafrule"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSWaf(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafRateBasedRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafRateBasedRuleConfigTags1(ruleName, "key1", "value1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSWafRateBasedRuleExists(resourceName, &rule),
					testAccMatchResourceAttrGlobalARN(resourceName, "arn", "waf", regexp.MustCompile(`ratebasedrule/.+`)),
					resource.TestCheckResourceAttr(resourceName, "name", ruleName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				Config: testAccAWSWafRateBasedRuleConfigTags2(ruleName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSWafRateBasedRuleExists(resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "name", ruleName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSWafRateBasedRuleConfigTags1(ruleName, "key2", "value2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSWafRateBasedRuleExists(resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "name", ruleName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
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

func testAccCheckAWSWafRateBasedRuleDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_waf_rate_based_rule" {
			continue
		}

		conn := testAccProvider.Meta().(*AWSClient).wafconn
		resp, err := conn.GetRateBasedRule(
			&waf.GetRateBasedRuleInput{
				RuleId: aws.String(rs.Primary.ID),
			})

		if err == nil {
			if *resp.Rule.RuleId == rs.Primary.ID {
				return fmt.Errorf("WAF Rule %s still exists", rs.Primary.ID)
			}
		}

		// Return nil if the Rule is already destroyed
		if awsErr, ok := err.(awserr.Error); ok {
			if awsErr.Code() == waf.ErrCodeNonexistentItemException {
				return nil
			}
		}

		return err
	}

	return nil
}

func testAccCheckAWSWafRateBasedRuleExists(n string, v *waf.RateBasedRule) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No WAF Rule ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).wafconn
		resp, err := conn.GetRateBasedRule(&waf.GetRateBasedRuleInput{
			RuleId: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return err
		}

		if *resp.Rule.RuleId == rs.Primary.ID {
			*v = *resp.Rule
			return nil
		}

		return fmt.Errorf("WAF Rule (%s) not found", rs.Primary.ID)
	}
}

func testAccAWSWafRateBasedRuleConfig(name string) string {
	return fmt.Sprintf(`
resource "aws_waf_ipset" "ipset" {
  name = "%s"

  ip_set_descriptors {
    type  = "IPV4"
    value = "192.0.7.0/24"
  }
}

resource "aws_waf_rate_based_rule" "wafrule" {
  depends_on  = [aws_waf_ipset.ipset]
  name        = "%s"
  metric_name = "%s"
  rate_key    = "IP"
  rate_limit  = 2000

  predicates {
    data_id = aws_waf_ipset.ipset.id
    negated = false
    type    = "IPMatch"
  }
}
`, name, name, name)
}

func testAccAWSWafRateBasedRuleConfig_changeRateLimit(name string, rateLimit int) string {
	return fmt.Sprintf(`
resource "aws_waf_ipset" "ipset" {
  name = "%s"

  ip_set_descriptors {
    type  = "IPV4"
    value = "192.0.7.0/24"
  }
}

resource "aws_waf_rate_based_rule" "wafrule" {
  depends_on  = [aws_waf_ipset.ipset]
  name        = "%[1]s"
  metric_name = "%[1]s"
  rate_key    = "IP"
  rate_limit  = %[2]d

  predicates {
    data_id = aws_waf_ipset.ipset.id
    negated = false
    type    = "IPMatch"
  }
}
`, name, rateLimit)
}

func testAccAWSWafRateBasedRuleConfigChangeName(name string) string {
	return fmt.Sprintf(`
resource "aws_waf_ipset" "ipset" {
  name = "%s"

  ip_set_descriptors {
    type  = "IPV4"
    value = "192.0.7.0/24"
  }
}

resource "aws_waf_rate_based_rule" "wafrule" {
  depends_on  = [aws_waf_ipset.ipset]
  name        = "%s"
  metric_name = "%s"
  rate_key    = "IP"
  rate_limit  = 2000

  predicates {
    data_id = aws_waf_ipset.ipset.id
    negated = false
    type    = "IPMatch"
  }
}
`, name, name, name)
}

func testAccAWSWafRateBasedRuleConfig_changePredicates(name string) string {
	return fmt.Sprintf(`
resource "aws_waf_ipset" "ipset" {
  name = "%s"

  ip_set_descriptors {
    type  = "IPV4"
    value = "192.0.7.0/24"
  }
}

resource "aws_waf_byte_match_set" "set" {
  name = "%s"

  byte_match_tuples {
    text_transformation   = "NONE"
    target_string         = "badrefer1"
    positional_constraint = "CONTAINS"

    field_to_match {
      type = "HEADER"
      data = "referer"
    }
  }
}

resource "aws_waf_rate_based_rule" "wafrule" {
  name        = "%s"
  metric_name = "%s"
  rate_key    = "IP"
  rate_limit  = 2000

  predicates {
    data_id = aws_waf_byte_match_set.set.id
    negated = true
    type    = "ByteMatch"
  }
}
`, name, name, name, name)
}

func testAccAWSWafRateBasedRuleConfig_noPredicates(name string) string {
	return fmt.Sprintf(`
resource "aws_waf_rate_based_rule" "wafrule" {
  name        = "%s"
  metric_name = "%s"
  rate_key    = "IP"
  rate_limit  = 2000
}
`, name, name)
}

func testAccAWSWafRateBasedRuleConfigTags1(name, tag1Key, tag1Value string) string {
	return fmt.Sprintf(`
resource "aws_waf_rate_based_rule" "wafrule" {
  name        = "%s"
  metric_name = "%s"
  rate_key    = "IP"
  rate_limit  = 2000

  tags = {
    %q = %q
  }
}
`, name, name, tag1Key, tag1Value)
}

func testAccAWSWafRateBasedRuleConfigTags2(name, tag1Key, tag1Value, tag2Key, tag2Value string) string {
	return fmt.Sprintf(`
resource "aws_waf_rate_based_rule" "wafrule" {
  name        = "%s"
  metric_name = "%s"
  rate_key    = "IP"
  rate_limit  = 2000

  tags = {
    %q = %q
    %q = %q
  }
}
`, name, name, tag1Key, tag1Value, tag2Key, tag2Value)
}
