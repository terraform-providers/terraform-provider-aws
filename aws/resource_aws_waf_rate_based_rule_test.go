package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/waf"
	"github.com/hashicorp/terraform/helper/acctest"
)

func TestAccAWSWafRateBasedRule_basic(t *testing.T) {
	var v waf.RateBasedRule
	wafRuleName := fmt.Sprintf("wafrule%s", acctest.RandString(5))
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafRateBasedRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafRateBasedRuleConfig(wafRuleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafRateBasedRuleExists("aws_waf_rate_based_rule.wafrule", &v),
					resource.TestCheckResourceAttr(
						"aws_waf_rate_based_rule.wafrule", "name", wafRuleName),
					resource.TestCheckResourceAttr(
						"aws_waf_rate_based_rule.wafrule", "predicates.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_waf_rate_based_rule.wafrule", "metric_name", wafRuleName),
				),
			},
		},
	})
}

func TestAccAWSWafRateBasedRule_changeNameForceNew(t *testing.T) {
	var before, after waf.RateBasedRule
	wafRuleName := fmt.Sprintf("wafrule%s", acctest.RandString(5))
	wafRuleNewName := fmt.Sprintf("wafrulenew%s", acctest.RandString(5))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafIPSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafRateBasedRuleConfig(wafRuleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafRateBasedRuleExists("aws_waf_rate_based_rule.wafrule", &before),
					resource.TestCheckResourceAttr(
						"aws_waf_rate_based_rule.wafrule", "name", wafRuleName),
					resource.TestCheckResourceAttr(
						"aws_waf_rate_based_rule.wafrule", "predicates.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_waf_rate_based_rule.wafrule", "metric_name", wafRuleName),
				),
			},
			{
				Config: testAccAWSWafRateBasedRuleConfigChangeName(wafRuleNewName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafRateBasedRuleExists("aws_waf_rate_based_rule.wafrule", &after),
					resource.TestCheckResourceAttr(
						"aws_waf_rate_based_rule.wafrule", "name", wafRuleNewName),
					resource.TestCheckResourceAttr(
						"aws_waf_rate_based_rule.wafrule", "predicates.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_waf_rate_based_rule.wafrule", "metric_name", wafRuleNewName),
				),
			},
		},
	})
}

func TestAccAWSWafRateBasedRule_disappears(t *testing.T) {
	var v waf.RateBasedRule
	wafRuleName := fmt.Sprintf("wafrule%s", acctest.RandString(5))
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafRateBasedRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafRateBasedRuleConfig(wafRuleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafRateBasedRuleExists("aws_waf_rate_based_rule.wafrule", &v),
					testAccCheckAWSWafRateBasedRuleDisappears(&v),
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
	var idx int
	ruleName := fmt.Sprintf("wafrule%s", acctest.RandString(5))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafRateBasedRuleConfig(ruleName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSWafIPSetExists("aws_waf_ipset.ipset", &ipset),
					testAccCheckAWSWafRateBasedRuleExists("aws_waf_rate_based_rule.wafrule", &before),
					resource.TestCheckResourceAttr("aws_waf_rate_based_rule.wafrule", "name", ruleName),
					resource.TestCheckResourceAttr("aws_waf_rate_based_rule.wafrule", "predicates.#", "1"),
					computeWafRateBasedRulePredicateWithIpSet(&ipset, false, "IPMatch", &idx),
					testCheckResourceAttrWithIndexesAddr("aws_waf_rate_based_rule.wafrule", "predicates.%d.negated", &idx, "false"),
					testCheckResourceAttrWithIndexesAddr("aws_waf_rate_based_rule.wafrule", "predicates.%d.type", &idx, "IPMatch"),
				),
			},
			{
				Config: testAccAWSWafRateBasedRuleConfig_changePredicates(ruleName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSWafByteMatchSetExists("aws_waf_byte_match_set.set", &byteMatchSet),
					testAccCheckAWSWafRateBasedRuleExists("aws_waf_rate_based_rule.wafrule", &after),
					resource.TestCheckResourceAttr("aws_waf_rate_based_rule.wafrule", "name", ruleName),
					resource.TestCheckResourceAttr("aws_waf_rate_based_rule.wafrule", "predicates.#", "1"),
					computeWafRateBasedRulePredicateWithByteMatchSet(&byteMatchSet, true, "ByteMatch", &idx),
					testCheckResourceAttrWithIndexesAddr("aws_waf_rate_based_rule.wafrule", "predicates.%d.negated", &idx, "true"),
					testCheckResourceAttrWithIndexesAddr("aws_waf_rate_based_rule.wafrule", "predicates.%d.type", &idx, "ByteMatch"),
				),
			},
		},
	})
}

// computeWafRateBasedRulePredicateWithIpSet calculates index
// which isn't static because dataId is generated as part of the test
func computeWafRateBasedRulePredicateWithIpSet(ipSet *waf.IPSet, negated bool, pType string, idx *int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		predicateResource := resourceAwsWafRateBasedRule().Schema["predicates"].Elem.(*schema.Resource)

		m := map[string]interface{}{
			"data_id": *ipSet.IPSetId,
			"negated": negated,
			"type":    pType,
		}

		f := schema.HashResource(predicateResource)
		*idx = f(m)

		return nil
	}
}

// computeWafRateBasedRulePredicateWithByteMatchSet calculates index
// which isn't static because dataId is generated as part of the test
func computeWafRateBasedRulePredicateWithByteMatchSet(set *waf.ByteMatchSet, negated bool, pType string, idx *int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		predicateResource := resourceAwsWafRateBasedRule().Schema["predicates"].Elem.(*schema.Resource)

		m := map[string]interface{}{
			"data_id": *set.ByteMatchSetId,
			"negated": negated,
			"type":    pType,
		}

		f := schema.HashResource(predicateResource)
		*idx = f(m)

		return nil
	}
}

func TestAccAWSWafRateBasedRule_noPredicates(t *testing.T) {
	var rule waf.RateBasedRule
	ruleName := fmt.Sprintf("wafrule%s", acctest.RandString(5))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafRateBasedRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafRateBasedRuleConfig_noPredicates(ruleName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSWafRateBasedRuleExists("aws_waf_rate_based_rule.wafrule", &rule),
					resource.TestCheckResourceAttr(
						"aws_waf_rate_based_rule.wafrule", "name", ruleName),
					resource.TestCheckResourceAttr(
						"aws_waf_rate_based_rule.wafrule", "predicates.#", "0"),
				),
			},
		},
	})
}

func testAccCheckAWSWafRateBasedRuleDisappears(v *waf.RateBasedRule) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).wafconn

		wr := newWafRetryer(conn)
		_, err := wr.RetryWithToken(func(token *string) (interface{}, error) {
			req := &waf.UpdateRateBasedRuleInput{
				ChangeToken: token,
				RuleId:      v.RuleId,
				RateLimit:   v.RateLimit,
			}

			for _, Predicate := range v.MatchPredicates {
				Predicate := &waf.RuleUpdate{
					Action: aws.String("DELETE"),
					Predicate: &waf.Predicate{
						Negated: Predicate.Negated,
						Type:    Predicate.Type,
						DataId:  Predicate.DataId,
					},
				}
				req.Updates = append(req.Updates, Predicate)
			}

			return conn.UpdateRateBasedRule(req)
		})
		if err != nil {
			return fmt.Errorf("Error Updating WAF Rule: %s", err)
		}

		_, err = wr.RetryWithToken(func(token *string) (interface{}, error) {
			opts := &waf.DeleteRateBasedRuleInput{
				ChangeToken: token,
				RuleId:      v.RuleId,
			}
			return conn.DeleteRateBasedRule(opts)
		})
		if err != nil {
			return fmt.Errorf("Error Deleting WAF Rule: %s", err)
		}
		return nil
	}
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
			if awsErr.Code() == "WAFNonexistentItemException" {
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
    type = "IPV4"
    value = "192.0.7.0/24"
  }
}

resource "aws_waf_rate_based_rule" "wafrule" {
  depends_on = ["aws_waf_ipset.ipset"]
  name = "%s"
  metric_name = "%s"
  rate_key = "IP"
  rate_limit = 2000
  predicates {
    data_id = "${aws_waf_ipset.ipset.id}"
    negated = false
    type = "IPMatch"
  }
}`, name, name, name)
}

func testAccAWSWafRateBasedRuleConfigChangeName(name string) string {
	return fmt.Sprintf(`
resource "aws_waf_ipset" "ipset" {
  name = "%s"
  ip_set_descriptors {
    type = "IPV4"
    value = "192.0.7.0/24"
  }
}

resource "aws_waf_rate_based_rule" "wafrule" {
  depends_on = ["aws_waf_ipset.ipset"]
  name = "%s"
  metric_name = "%s"
  rate_key = "IP"
  rate_limit = 2000
  predicates {
    data_id = "${aws_waf_ipset.ipset.id}"
    negated = false
    type = "IPMatch"
  }
}`, name, name, name)
}

func testAccAWSWafRateBasedRuleConfig_changePredicates(name string) string {
	return fmt.Sprintf(`
resource "aws_waf_ipset" "ipset" {
  name = "%s"
  ip_set_descriptors {
    type = "IPV4"
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
  name = "%s"
  metric_name = "%s"
  rate_key = "IP"
  rate_limit = 2000
  predicates {
    data_id = "${aws_waf_byte_match_set.set.id}"
    negated = true
    type = "ByteMatch"
  }
}`, name, name, name, name)
}

func testAccAWSWafRateBasedRuleConfig_noPredicates(name string) string {
	return fmt.Sprintf(`
resource "aws_waf_rate_based_rule" "wafrule" {
  name = "%s"
  metric_name = "%s"
  rate_key = "IP"
  rate_limit = 2000
}`, name, name)
}
