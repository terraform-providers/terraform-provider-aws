package aws

import (
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/waf"
	"github.com/aws/aws-sdk-go/service/wafregional"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	resource.AddTestSweepers("aws_wafregional_rate_based_rule", &resource.Sweeper{
		Name: "aws_wafregional_rate_based_rule",
		F:    testSweepWafRegionalRateBasedRules,
		Dependencies: []string{
			"aws_wafregional_web_acl",
		},
	})
}

func testSweepWafRegionalRateBasedRules(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).wafregionalconn

	input := &waf.ListRateBasedRulesInput{}

	for {
		output, err := conn.ListRateBasedRules(input)

		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping WAF Regional Rate-Based Rule sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing WAF Regional Rate-Based Rules: %s", err)
		}

		for _, rule := range output.Rules {
			deleteInput := &waf.DeleteRateBasedRuleInput{
				RuleId: rule.RuleId,
			}
			id := aws.StringValue(rule.RuleId)
			wr := newWafRegionalRetryer(conn, region)

			_, err := wr.RetryWithToken(func(token *string) (interface{}, error) {
				deleteInput.ChangeToken = token
				log.Printf("[INFO] Deleting WAF Regional Rate-Based Rule: %s", id)
				return conn.DeleteRateBasedRule(deleteInput)
			})

			if isAWSErr(err, wafregional.ErrCodeWAFNonEmptyEntityException, "") {
				getRateBasedRuleInput := &waf.GetRateBasedRuleInput{
					RuleId: rule.RuleId,
				}

				getRateBasedRuleOutput, getRateBasedRuleErr := conn.GetRateBasedRule(getRateBasedRuleInput)

				if getRateBasedRuleErr != nil {
					return fmt.Errorf("error getting WAF Regional Rate-Based Rule (%s): %s", id, getRateBasedRuleErr)
				}

				var updates []*waf.RuleUpdate
				updateRateBasedRuleInput := &waf.UpdateRateBasedRuleInput{
					RateLimit: getRateBasedRuleOutput.Rule.RateLimit,
					RuleId:    rule.RuleId,
					Updates:   updates,
				}

				for _, predicate := range getRateBasedRuleOutput.Rule.MatchPredicates {
					update := &waf.RuleUpdate{
						Action:    aws.String(waf.ChangeActionDelete),
						Predicate: predicate,
					}

					updateRateBasedRuleInput.Updates = append(updateRateBasedRuleInput.Updates, update)
				}

				_, updateWebACLErr := wr.RetryWithToken(func(token *string) (interface{}, error) {
					updateRateBasedRuleInput.ChangeToken = token
					log.Printf("[INFO] Removing Predicates from WAF Regional Rate-Based Rule: %s", id)
					return conn.UpdateRateBasedRule(updateRateBasedRuleInput)
				})

				if updateWebACLErr != nil {
					return fmt.Errorf("error removing predicates from WAF Regional Rate-Based Rule (%s): %s", id, updateWebACLErr)
				}

				_, err = wr.RetryWithToken(func(token *string) (interface{}, error) {
					deleteInput.ChangeToken = token
					log.Printf("[INFO] Deleting WAF Regional Rate-Based Rule: %s", id)
					return conn.DeleteRateBasedRule(deleteInput)
				})
			}

			if err != nil {
				return fmt.Errorf("error deleting WAF Regional Rate-Based Rule (%s): %s", id, err)
			}
		}

		if aws.StringValue(output.NextMarker) == "" {
			break
		}

		input.NextMarker = output.NextMarker
	}

	return nil
}

func TestAccAWSWafRegionalRateBasedRule_basic(t *testing.T) {
	var v waf.RateBasedRule
	resourceName := "aws_wafregional_rate_based_rule.wafrule"
	wafRuleName := fmt.Sprintf("wafrule%s", acctest.RandString(5))
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(wafregional.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafRegionalRateBasedRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafRegionalRateBasedRuleConfig(wafRuleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafRegionalRateBasedRuleExists(resourceName, &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "waf-regional", regexp.MustCompile(`ratebasedrule/.+`)),
					resource.TestCheckResourceAttr(resourceName, "name", wafRuleName),
					resource.TestCheckResourceAttr(resourceName, "predicate.#", "1"),
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

func TestAccAWSWafRegionalRateBasedRule_tags(t *testing.T) {
	var v waf.RateBasedRule
	resourceName := "aws_wafregional_rate_based_rule.wafrule"
	wafRuleName := fmt.Sprintf("wafrule%s", acctest.RandString(5))
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(wafregional.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafRegionalRateBasedRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafRegionalRateBasedRuleConfigTags1(wafRuleName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafRegionalRateBasedRuleExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSWafRegionalRateBasedRuleConfigTags2(wafRuleName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafRegionalRateBasedRuleExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSWafRegionalRateBasedRuleConfigTags1(wafRuleName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafRegionalRateBasedRuleExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccAWSWafRegionalRateBasedRule_changeNameForceNew(t *testing.T) {
	var before, after waf.RateBasedRule
	resourceName := "aws_wafregional_rate_based_rule.wafrule"
	wafRuleName := fmt.Sprintf("wafrule%s", acctest.RandString(5))
	wafRuleNewName := fmt.Sprintf("wafrulenew%s", acctest.RandString(5))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(wafregional.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafRegionalRateBasedRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafRegionalRateBasedRuleConfig(wafRuleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafRegionalRateBasedRuleExists(resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, "name", wafRuleName),
					resource.TestCheckResourceAttr(resourceName, "predicate.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "metric_name", wafRuleName),
				),
			},
			{
				Config: testAccAWSWafRegionalRateBasedRuleConfigChangeName(wafRuleNewName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafRegionalRateBasedRuleExists(resourceName, &after),
					testAccCheckAWSWafRateBasedRuleIdDiffers(&before, &after),
					resource.TestCheckResourceAttr(resourceName, "name", wafRuleNewName),
					resource.TestCheckResourceAttr(resourceName, "predicate.#", "1"),
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

func TestAccAWSWafRegionalRateBasedRule_disappears(t *testing.T) {
	var v waf.RateBasedRule
	resourceName := "aws_wafregional_rate_based_rule.wafrule"
	wafRuleName := fmt.Sprintf("wafrule%s", acctest.RandString(5))
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(wafregional.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafRegionalRateBasedRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafRegionalRateBasedRuleConfig(wafRuleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafRegionalRateBasedRuleExists(resourceName, &v),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsWafRegionalRateBasedRule(), resourceName),
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
	resourceName := "aws_wafregional_rate_based_rule.wafrule"
	ruleName := fmt.Sprintf("wafrule%s", acctest.RandString(5))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(wafregional.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafRegionalRateBasedRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafRegionalRateBasedRuleConfig(ruleName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSWafRegionalIPSetExists("aws_wafregional_ipset.ipset", &ipset),
					testAccCheckAWSWafRegionalRateBasedRuleExists(resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, "name", ruleName),
					resource.TestCheckResourceAttr(resourceName, "predicate.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "predicate.*", map[string]string{
						"negated": "false",
						"type":    "IPMatch",
					}),
				),
			},
			{
				Config: testAccAWSWafRegionalRateBasedRuleConfig_changePredicates(ruleName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSWafRegionalByteMatchSetExists("aws_wafregional_byte_match_set.set", &byteMatchSet),
					testAccCheckAWSWafRegionalRateBasedRuleExists(resourceName, &after),
					resource.TestCheckResourceAttr(resourceName, "name", ruleName),
					resource.TestCheckResourceAttr(resourceName, "predicate.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "predicate.*", map[string]string{
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

func TestAccAWSWafRegionalRateBasedRule_changeRateLimit(t *testing.T) {
	var before, after waf.RateBasedRule
	resourceName := "aws_wafregional_rate_based_rule.wafrule"
	ruleName := fmt.Sprintf("wafrule%s", acctest.RandString(5))
	rateLimitBefore := "2000"
	rateLimitAfter := "2001"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(wafregional.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafRegionalRateBasedRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafRegionalRateBasedRuleWithRateLimitConfig(ruleName, rateLimitBefore),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSWafRegionalRateBasedRuleExists(resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, "name", ruleName),
					resource.TestCheckResourceAttr(resourceName, "rate_limit", rateLimitBefore),
				),
			},
			{
				Config: testAccAWSWafRegionalRateBasedRuleWithRateLimitConfig(ruleName, rateLimitAfter),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSWafRegionalRateBasedRuleExists(resourceName, &after),
					resource.TestCheckResourceAttr(resourceName, "name", ruleName),
					resource.TestCheckResourceAttr(resourceName, "rate_limit", rateLimitAfter),
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

func TestAccAWSWafRegionalRateBasedRule_noPredicates(t *testing.T) {
	var rule waf.RateBasedRule
	resourceName := "aws_wafregional_rate_based_rule.wafrule"
	ruleName := fmt.Sprintf("wafrule%s", acctest.RandString(5))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(wafregional.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafRegionalRateBasedRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafRegionalRateBasedRuleConfig_noPredicates(ruleName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSWafRegionalRateBasedRuleExists(resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "name", ruleName),
					resource.TestCheckResourceAttr(resourceName, "predicate.#", "0"),
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

func testAccCheckAWSWafRateBasedRuleIdDiffers(before, after *waf.RateBasedRule) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if *before.RuleId == *after.RuleId {
			return fmt.Errorf("Expected different IDs, given %q for both rules", *before.RuleId)
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
  name = %[1]q

  ip_set_descriptor {
    type  = "IPV4"
    value = "192.0.7.0/24"
  }
}

resource "aws_wafregional_rate_based_rule" "wafrule" {
  name        = %[1]q
  metric_name = %[1]q
  rate_key    = "IP"
  rate_limit  = 2000

  predicate {
    data_id = aws_wafregional_ipset.ipset.id
    negated = false
    type    = "IPMatch"
  }
}
`, name)
}

func testAccAWSWafRegionalRateBasedRuleConfigTags1(name, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_wafregional_ipset" "ipset" {
  name = %[1]q

  ip_set_descriptor {
    type  = "IPV4"
    value = "192.0.7.0/24"
  }
}

resource "aws_wafregional_rate_based_rule" "wafrule" {
  name        = %[1]q
  metric_name = %[1]q
  rate_key    = "IP"
  rate_limit  = 2000

  predicate {
    data_id = aws_wafregional_ipset.ipset.id
    negated = false
    type    = "IPMatch"
  }

  tags = {
    %[2]q = %[3]q
  }
}
`, name, tagKey1, tagValue1)
}

func testAccAWSWafRegionalRateBasedRuleConfigTags2(name, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_wafregional_ipset" "ipset" {
  name = %[1]q

  ip_set_descriptor {
    type  = "IPV4"
    value = "192.0.7.0/24"
  }
}

resource "aws_wafregional_rate_based_rule" "wafrule" {
  name        = %[1]q
  metric_name = %[1]q
  rate_key    = "IP"
  rate_limit  = 2000

  predicate {
    data_id = aws_wafregional_ipset.ipset.id
    negated = false
    type    = "IPMatch"
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, name, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccAWSWafRegionalRateBasedRuleConfigChangeName(name string) string {
	return fmt.Sprintf(`
resource "aws_wafregional_ipset" "ipset" {
  name = %[1]q

  ip_set_descriptor {
    type  = "IPV4"
    value = "192.0.7.0/24"
  }
}

resource "aws_wafregional_rate_based_rule" "wafrule" {
  name        = %[1]q
  metric_name = %[1]q
  rate_key    = "IP"
  rate_limit  = 2000

  predicate {
    data_id = aws_wafregional_ipset.ipset.id
    negated = false
    type    = "IPMatch"
  }
}
`, name)
}

func testAccAWSWafRegionalRateBasedRuleConfig_changePredicates(name string) string {
	return fmt.Sprintf(`
resource "aws_wafregional_ipset" "ipset" {
  name = %[1]q

  ip_set_descriptor {
    type  = "IPV4"
    value = "192.0.7.0/24"
  }
}

resource "aws_wafregional_byte_match_set" "set" {
  name = %[1]q

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

resource "aws_wafregional_rate_based_rule" "wafrule" {
  name        = %[1]q
  metric_name = %[1]q
  rate_key    = "IP"
  rate_limit  = 2000

  predicate {
    data_id = aws_wafregional_byte_match_set.set.id
    negated = true
    type    = "ByteMatch"
  }
}
`, name)
}

func testAccAWSWafRegionalRateBasedRuleConfig_noPredicates(name string) string {
	return fmt.Sprintf(`
resource "aws_wafregional_rate_based_rule" "wafrule" {
  name        = %[1]q
  metric_name = %[1]q
  rate_key    = "IP"
  rate_limit  = 2000
}
`, name)
}

func testAccAWSWafRegionalRateBasedRuleWithRateLimitConfig(name string, limit string) string {
	return fmt.Sprintf(`
resource "aws_wafregional_rate_based_rule" "wafrule" {
  name        = "%s"
  metric_name = "%s"
  rate_key    = "IP"
  rate_limit  = %s
}
`, name, name, limit)
}
