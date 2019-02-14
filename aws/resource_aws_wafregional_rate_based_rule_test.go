package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/waf"
	"github.com/aws/aws-sdk-go/service/wafregional"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSWafRegionalRateBasedRule_basic(t *testing.T) {
	var v waf.RateBasedRule
	wafRuleName := fmt.Sprintf("wafrule%s", acctest.RandString(5))
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafRegionalRateBasedRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafRegionalRateBasedRuleConfig(wafRuleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafRegionalRateBasedRuleExists("aws_wafregional_rate_based_rule.wafrule", &v),
					resource.TestCheckResourceAttr(
						"aws_wafregional_rate_based_rule.wafrule", "name", wafRuleName),
					resource.TestCheckResourceAttr(
						"aws_wafregional_rate_based_rule.wafrule", "predicate.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_wafregional_rate_based_rule.wafrule", "metric_name", wafRuleName),
				),
			},
		},
	})
}

func TestAccAWSWafRegionalRateBasedRule_changeNameForceNew(t *testing.T) {
	var before, after waf.RateBasedRule
	wafRuleName := fmt.Sprintf("wafrule%s", acctest.RandString(5))
	wafRuleNewName := fmt.Sprintf("wafrulenew%s", acctest.RandString(5))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafRegionalRateBasedRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafRegionalRateBasedRuleConfig(wafRuleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafRegionalRateBasedRuleExists("aws_wafregional_rate_based_rule.wafrule", &before),
					resource.TestCheckResourceAttr(
						"aws_wafregional_rate_based_rule.wafrule", "name", wafRuleName),
					resource.TestCheckResourceAttr(
						"aws_wafregional_rate_based_rule.wafrule", "predicate.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_wafregional_rate_based_rule.wafrule", "metric_name", wafRuleName),
				),
			},
			{
				Config: testAccAWSWafRegionalRateBasedRuleConfigChangeName(wafRuleNewName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafRegionalRateBasedRuleExists("aws_wafregional_rate_based_rule.wafrule", &after),
					testAccCheckAWSWafRateBasedRuleIdDiffers(&before, &after),
					resource.TestCheckResourceAttr(
						"aws_wafregional_rate_based_rule.wafrule", "name", wafRuleNewName),
					resource.TestCheckResourceAttr(
						"aws_wafregional_rate_based_rule.wafrule", "predicate.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_wafregional_rate_based_rule.wafrule", "metric_name", wafRuleNewName),
				),
			},
		},
	})
}

func TestAccAWSWafRegionalRateBasedRule_disappears(t *testing.T) {
	var v waf.RateBasedRule
	wafRuleName := fmt.Sprintf("wafrule%s", acctest.RandString(5))
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafRegionalRateBasedRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafRegionalRateBasedRuleConfig(wafRuleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafRegionalRateBasedRuleExists("aws_wafregional_rate_based_rule.wafrule", &v),
					testAccCheckAWSWafRegionalRateBasedRuleDisappears(&v),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSWafRegionalRateBasedRule_changePredicates(t *testing.T) {
	var ipset waf.IPSet
	var byteMatchSet waf.ByteMatchSet

	var before, after waf.RateBasedRule
	var idx int
	ruleName := fmt.Sprintf("wafrule%s", acctest.RandString(5))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafRegionalRateBasedRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafRegionalRateBasedRuleConfig(ruleName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSWafRegionalIPSetExists("aws_wafregional_ipset.ipset", &ipset),
					testAccCheckAWSWafRegionalRateBasedRuleExists("aws_wafregional_rate_based_rule.wafrule", &before),
					resource.TestCheckResourceAttr("aws_wafregional_rate_based_rule.wafrule", "name", ruleName),
					resource.TestCheckResourceAttr("aws_wafregional_rate_based_rule.wafrule", "predicate.#", "1"),
					computeWafRegionalRateBasedRulePredicateWithIpSet(&ipset, false, "IPMatch", &idx),
					testCheckResourceAttrWithIndexesAddr("aws_wafregional_rate_based_rule.wafrule", "predicate.%d.negated", &idx, "false"),
					testCheckResourceAttrWithIndexesAddr("aws_wafregional_rate_based_rule.wafrule", "predicate.%d.type", &idx, "IPMatch"),
				),
			},
			{
				Config: testAccAWSWafRegionalRateBasedRuleConfig_changePredicates(ruleName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSWafRegionalByteMatchSetExists("aws_wafregional_byte_match_set.set", &byteMatchSet),
					testAccCheckAWSWafRegionalRateBasedRuleExists("aws_wafregional_rate_based_rule.wafrule", &after),
					resource.TestCheckResourceAttr("aws_wafregional_rate_based_rule.wafrule", "name", ruleName),
					resource.TestCheckResourceAttr("aws_wafregional_rate_based_rule.wafrule", "predicate.#", "1"),
					computeWafRegionalRateBasedRulePredicateWithByteMatchSet(&byteMatchSet, true, "ByteMatch", &idx),
					testCheckResourceAttrWithIndexesAddr("aws_wafregional_rate_based_rule.wafrule", "predicate.%d.negated", &idx, "true"),
					testCheckResourceAttrWithIndexesAddr("aws_wafregional_rate_based_rule.wafrule", "predicate.%d.type", &idx, "ByteMatch"),
				),
			},
		},
	})
}

func TestAccAWSWafRegionalRateBasedRule_changeRateLimit(t *testing.T) {
	var before, after waf.RateBasedRule
	ruleName := fmt.Sprintf("wafrule%s", acctest.RandString(5))
	rateLimitBefore := "2000"
	rateLimitAfter := "2001"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafRegionalRateBasedRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafRegionalRateBasedRuleWithRateLimitConfig(ruleName, rateLimitBefore),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSWafRegionalRateBasedRuleExists("aws_wafregional_rate_based_rule.wafrule", &before),
					resource.TestCheckResourceAttr("aws_wafregional_rate_based_rule.wafrule", "name", ruleName),
					resource.TestCheckResourceAttr("aws_wafregional_rate_based_rule.wafrule", "rate_limit", rateLimitBefore),
				),
			},
			{
				Config: testAccAWSWafRegionalRateBasedRuleWithRateLimitConfig(ruleName, rateLimitAfter),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSWafRegionalRateBasedRuleExists("aws_wafregional_rate_based_rule.wafrule", &after),
					resource.TestCheckResourceAttr("aws_wafregional_rate_based_rule.wafrule", "name", ruleName),
					resource.TestCheckResourceAttr("aws_wafregional_rate_based_rule.wafrule", "rate_limit", rateLimitAfter),
				),
			},
		},
	})
}

// computeWafRegionalRateBasedRulePredicateWithIpSet calculates index
// which isn't static because dataId is generated as part of the test
func computeWafRegionalRateBasedRulePredicateWithIpSet(ipSet *waf.IPSet, negated bool, pType string, idx *int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		predicateResource := resourceAwsWafRegionalRateBasedRule().Schema["predicate"].Elem.(*schema.Resource)

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

// computeWafRegionalRateBasedRulePredicateWithByteMatchSet calculates index
// which isn't static because dataId is generated as part of the test
func computeWafRegionalRateBasedRulePredicateWithByteMatchSet(set *waf.ByteMatchSet, negated bool, pType string, idx *int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		predicateResource := resourceAwsWafRegionalRateBasedRule().Schema["predicate"].Elem.(*schema.Resource)

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

func TestAccAWSWafRegionalRateBasedRule_noPredicates(t *testing.T) {
	var rule waf.RateBasedRule
	ruleName := fmt.Sprintf("wafrule%s", acctest.RandString(5))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafRegionalRateBasedRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafRegionalRateBasedRuleConfig_noPredicates(ruleName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSWafRegionalRateBasedRuleExists("aws_wafregional_rate_based_rule.wafrule", &rule),
					resource.TestCheckResourceAttr(
						"aws_wafregional_rate_based_rule.wafrule", "name", ruleName),
					resource.TestCheckResourceAttr(
						"aws_wafregional_rate_based_rule.wafrule", "predicate.#", "0"),
				),
			},
		},
	})
}

func testAccCheckAWSWafRateBasedRuleIdDiffers(before, after *waf.RateBasedRule) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if *before.RuleId == *after.RuleId {
			return fmt.Errorf("Expected different IDs, given %q for both rules", *before.RuleId)
		}
		return nil
	}
}

func testAccCheckAWSWafRegionalRateBasedRuleDisappears(v *waf.RateBasedRule) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).wafregionalconn
		region := testAccProvider.Meta().(*AWSClient).region

		wr := newWafRegionalRetryer(conn, region)
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

func testAccCheckAWSWafRegionalRateBasedRuleDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_wafregional_rate_based_rule" {
			continue
		}

		conn := testAccProvider.Meta().(*AWSClient).wafregionalconn
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
		if isAWSErr(err, wafregional.ErrCodeWAFNonexistentItemException, "") {
			return nil
		}

		return err
	}

	return nil
}

func testAccCheckAWSWafRegionalRateBasedRuleExists(n string, v *waf.RateBasedRule) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No WAF Rule ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).wafregionalconn
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

		return fmt.Errorf("WAF Regional Rule (%s) not found", rs.Primary.ID)
	}
}

func testAccAWSWafRegionalRateBasedRuleConfig(name string) string {
	return fmt.Sprintf(`
resource "aws_wafregional_ipset" "ipset" {
  name = "%s"
  ip_set_descriptor {
    type = "IPV4"
    value = "192.0.7.0/24"
  }
}

resource "aws_wafregional_rate_based_rule" "wafrule" {
  name = "%s"
  metric_name = "%s"
  rate_key = "IP"
  rate_limit = 2000
  predicate {
    data_id = "${aws_wafregional_ipset.ipset.id}"
    negated = false
    type = "IPMatch"
  }
}`, name, name, name)
}

func testAccAWSWafRegionalRateBasedRuleConfigChangeName(name string) string {
	return fmt.Sprintf(`
resource "aws_wafregional_ipset" "ipset" {
  name = "%s"
  ip_set_descriptor {
    type = "IPV4"
    value = "192.0.7.0/24"
  }
}

resource "aws_wafregional_rate_based_rule" "wafrule" {
  name = "%s"
  metric_name = "%s"
  rate_key = "IP"
  rate_limit = 2000
  predicate {
    data_id = "${aws_wafregional_ipset.ipset.id}"
    negated = false
    type = "IPMatch"
  }
}`, name, name, name)
}

func testAccAWSWafRegionalRateBasedRuleConfig_changePredicates(name string) string {
	return fmt.Sprintf(`
resource "aws_wafregional_ipset" "ipset" {
  name = "%s"
  ip_set_descriptor {
    type = "IPV4"
    value = "192.0.7.0/24"
  }
}

resource "aws_wafregional_byte_match_set" "set" {
  name = "%s"
  byte_match_tuple {
    text_transformation   = "NONE"
    target_string         = "badrefer1"
    positional_constraint = "CONTAINS"

    field_to_match {
      type = "HEADER"
      data = "referer"
    }
  }
}

resource "aws_wafregional_rate_based_rule" "wafrule" {
  name = "%s"
  metric_name = "%s"
  rate_key = "IP"
  rate_limit = 2000
  predicate {
    data_id = "${aws_wafregional_byte_match_set.set.id}"
    negated = true
    type = "ByteMatch"
  }
}`, name, name, name, name)
}

func testAccAWSWafRegionalRateBasedRuleConfig_noPredicates(name string) string {
	return fmt.Sprintf(`
resource "aws_wafregional_rate_based_rule" "wafrule" {
  name = "%s"
  metric_name = "%s"
  rate_key = "IP"
  rate_limit = 2000
}`, name, name)
}

func testAccAWSWafRegionalRateBasedRuleWithRateLimitConfig(name string, limit string) string {
	return fmt.Sprintf(`
resource "aws_wafregional_rate_based_rule" "wafrule" {
  name = "%s"
  metric_name = "%s"
  rate_key = "IP"
  rate_limit = %s
}`, name, name, limit)
}
