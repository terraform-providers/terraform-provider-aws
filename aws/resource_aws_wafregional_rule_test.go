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

func TestAccAWSWafRegionalRule_basic(t *testing.T) {
	var v waf.Rule
	wafRuleName := fmt.Sprintf("wafrule%s", acctest.RandString(5))
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafRegionalRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafRegionalRuleConfig(wafRuleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafRegionalRuleExists("aws_wafregional_rule.wafrule", &v),
					resource.TestCheckResourceAttr(
						"aws_wafregional_rule.wafrule", "name", wafRuleName),
					resource.TestCheckResourceAttr(
						"aws_wafregional_rule.wafrule", "predicate.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_wafregional_rule.wafrule", "metric_name", wafRuleName),
				),
			},
		},
	})
}

func TestAccAWSWafRegionalRule_changeNameForceNew(t *testing.T) {
	var before, after waf.Rule
	wafRuleName := fmt.Sprintf("wafrule%s", acctest.RandString(5))
	wafRuleNewName := fmt.Sprintf("wafrulenew%s", acctest.RandString(5))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafRegionalIPSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafRegionalRuleConfig(wafRuleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafRegionalRuleExists("aws_wafregional_rule.wafrule", &before),
					resource.TestCheckResourceAttr(
						"aws_wafregional_rule.wafrule", "name", wafRuleName),
					resource.TestCheckResourceAttr(
						"aws_wafregional_rule.wafrule", "predicate.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_wafregional_rule.wafrule", "metric_name", wafRuleName),
				),
			},
			{
				Config: testAccAWSWafRegionalRuleConfigChangeName(wafRuleNewName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafRegionalRuleExists("aws_wafregional_rule.wafrule", &after),
					resource.TestCheckResourceAttr(
						"aws_wafregional_rule.wafrule", "name", wafRuleNewName),
					resource.TestCheckResourceAttr(
						"aws_wafregional_rule.wafrule", "predicate.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_wafregional_rule.wafrule", "metric_name", wafRuleNewName),
				),
			},
		},
	})
}

func TestAccAWSWafRegionalRule_disappears(t *testing.T) {
	var v waf.Rule
	wafRuleName := fmt.Sprintf("wafrule%s", acctest.RandString(5))
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafRegionalRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafRegionalRuleConfig(wafRuleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafRegionalRuleExists("aws_wafregional_rule.wafrule", &v),
					testAccCheckAWSWafRegionalRuleDisappears(&v),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSWafRegionalRule_noPredicates(t *testing.T) {
	var v waf.Rule
	wafRuleName := fmt.Sprintf("wafrule%s", acctest.RandString(5))
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafRegionalRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafRegionalRule_noPredicates(wafRuleName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSWafRegionalRuleExists("aws_wafregional_rule.wafrule", &v),
					resource.TestCheckResourceAttr(
						"aws_wafregional_rule.wafrule", "name", wafRuleName),
					resource.TestCheckResourceAttr(
						"aws_wafregional_rule.wafrule", "predicate.#", "0"),
				),
			},
		},
	})
}

func TestAccAWSWafRegionalRule_changePredicates(t *testing.T) {
	var ipset waf.IPSet
	var xssMatchSet waf.XssMatchSet

	var before, after waf.Rule
	var idx int
	ruleName := fmt.Sprintf("wafrule%s", acctest.RandString(5))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafRegionalRuleConfig(ruleName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSWafRegionalIPSetExists("aws_wafregional_ipset.ipset", &ipset),
					testAccCheckAWSWafRegionalRuleExists("aws_wafregional_rule.wafrule", &before),
					resource.TestCheckResourceAttr("aws_wafregional_rule.wafrule", "name", ruleName),
					resource.TestCheckResourceAttr("aws_wafregional_rule.wafrule", "predicate.#", "1"),
					computeWafRegionalRulePredicate(&ipset.IPSetId, false, "IPMatch", &idx),
					testCheckResourceAttrWithIndexesAddr("aws_wafregional_rule.wafrule", "predicate.%d.negated", &idx, "false"),
					testCheckResourceAttrWithIndexesAddr("aws_wafregional_rule.wafrule", "predicate.%d.type", &idx, "IPMatch"),
				),
			},
			{
				Config: testAccAWSWafRegionalRule_changePredicates(ruleName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSWafRegionalXssMatchSetExists("aws_wafregional_xss_match_set.xss_match_set", &xssMatchSet),
					testAccCheckAWSWafRegionalRuleExists("aws_wafregional_rule.wafrule", &after),
					resource.TestCheckResourceAttr("aws_wafregional_rule.wafrule", "name", ruleName),
					resource.TestCheckResourceAttr("aws_wafregional_rule.wafrule", "predicate.#", "2"),
					computeWafRegionalRulePredicate(&xssMatchSet.XssMatchSetId, true, "XssMatch", &idx),
					testCheckResourceAttrWithIndexesAddr("aws_wafregional_rule.wafrule", "predicate.%d.negated", &idx, "true"),
					testCheckResourceAttrWithIndexesAddr("aws_wafregional_rule.wafrule", "predicate.%d.type", &idx, "XssMatch"),
					computeWafRegionalRulePredicate(&ipset.IPSetId, true, "IPMatch", &idx),
					testCheckResourceAttrWithIndexesAddr("aws_wafregional_rule.wafrule", "predicate.%d.negated", &idx, "true"),
					testCheckResourceAttrWithIndexesAddr("aws_wafregional_rule.wafrule", "predicate.%d.type", &idx, "IPMatch"),
				),
			},
		},
	})
}

// Calculates the index which isn't static because dataId is generated as part of the test
func computeWafRegionalRulePredicate(dataId **string, negated bool, pType string, idx *int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		predicateResource := resourceAwsWafRegionalRule().Schema["predicate"].Elem.(*schema.Resource)
		m := map[string]interface{}{
			"data_id": **dataId,
			"negated": negated,
			"type":    pType,
		}

		f := schema.HashResource(predicateResource)
		*idx = f(m)

		return nil
	}
}

func testAccCheckAWSWafRegionalRuleDisappears(v *waf.Rule) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).wafregionalconn
		region := testAccProvider.Meta().(*AWSClient).region

		wr := newWafRegionalRetryer(conn, region)
		_, err := wr.RetryWithToken(func(token *string) (interface{}, error) {
			req := &waf.UpdateRuleInput{
				ChangeToken: token,
				RuleId:      v.RuleId,
			}

			for _, predicate := range v.Predicates {
				predicate := &waf.RuleUpdate{
					Action: aws.String("DELETE"),
					Predicate: &waf.Predicate{
						Negated: predicate.Negated,
						Type:    predicate.Type,
						DataId:  predicate.DataId,
					},
				}
				req.Updates = append(req.Updates, predicate)
			}

			return conn.UpdateRule(req)
		})
		if err != nil {
			return fmt.Errorf("Error Updating WAF Rule: %s", err)
		}

		_, err = wr.RetryWithToken(func(token *string) (interface{}, error) {
			opts := &waf.DeleteRuleInput{
				ChangeToken: token,
				RuleId:      v.RuleId,
			}
			return conn.DeleteRule(opts)
		})
		if err != nil {
			return fmt.Errorf("Error Deleting WAF Rule: %s", err)
		}
		return nil
	}
}

func testAccCheckAWSWafRegionalRuleDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_wafregional_rule" {
			continue
		}

		conn := testAccProvider.Meta().(*AWSClient).wafregionalconn
		resp, err := conn.GetRule(
			&waf.GetRuleInput{
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

func testAccCheckAWSWafRegionalRuleExists(n string, v *waf.Rule) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No WAF Rule ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).wafregionalconn
		resp, err := conn.GetRule(&waf.GetRuleInput{
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

func testAccAWSWafRegionalRuleConfig(name string) string {
	return fmt.Sprintf(`
resource "aws_wafregional_ipset" "ipset" {
  name = "%s"

  ip_set_descriptor {
    type = "IPV4"
    value = "192.0.7.0/24"
  }
}

resource "aws_wafregional_rule" "wafrule" {
  name = "%s"
  metric_name = "%s"

  predicate {
    data_id = "${aws_wafregional_ipset.ipset.id}"
    negated = false
    type = "IPMatch"
  }
}`, name, name, name)
}

func testAccAWSWafRegionalRuleConfigChangeName(name string) string {
	return fmt.Sprintf(`
resource "aws_wafregional_ipset" "ipset" {
  name = "%s"

  ip_set_descriptor {
    type = "IPV4"
    value = "192.0.7.0/24"
  }
}

resource "aws_wafregional_rule" "wafrule" {
  name = "%s"
  metric_name = "%s"

  predicate {
    data_id = "${aws_wafregional_ipset.ipset.id}"
    negated = false
    type = "IPMatch"
  }
}`, name, name, name)
}

func testAccAWSWafRegionalRule_noPredicates(name string) string {
	return fmt.Sprintf(`
resource "aws_wafregional_rule" "wafrule" {
	name = "%s"
	metric_name = "%s"
}
`, name, name)
}

func testAccAWSWafRegionalRule_changePredicates(name string) string {
	return fmt.Sprintf(`
resource "aws_wafregional_ipset" "ipset" {
  name = "%s"

  ip_set_descriptor {
    type = "IPV4"
    value = "192.0.7.0/24"
  }
}

resource "aws_wafregional_xss_match_set" "xss_match_set" {
  name = "%s"
  xss_match_tuple {
	text_transformation = "NONE"
	field_to_match {
	  type = "URI"
    }
  }
}

resource "aws_wafregional_rule" "wafrule" {
  name = "%s"
  metric_name = "%s"

  predicate {
    data_id = "${aws_wafregional_xss_match_set.xss_match_set.id}"
    negated = true
    type = "XssMatch"
  }

  predicate {
    data_id = "${aws_wafregional_ipset.ipset.id}"
    negated = true
    type = "IPMatch"
  }
}`, name, name, name, name)
}
