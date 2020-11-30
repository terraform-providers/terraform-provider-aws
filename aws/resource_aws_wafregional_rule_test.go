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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	resource.AddTestSweepers("aws_wafregional_rule", &resource.Sweeper{
		Name: "aws_wafregional_rule",
		F:    testSweepWafRegionalRules,
		Dependencies: []string{
			"aws_wafregional_web_acl",
		},
	})
}

func testSweepWafRegionalRules(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).wafregionalconn

	input := &waf.ListRulesInput{}

	for {
		output, err := conn.ListRules(input)

		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping WAF Regional Rule sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing WAF Regional Rules: %s", err)
		}

		for _, rule := range output.Rules {
			deleteInput := &waf.DeleteRuleInput{
				RuleId: rule.RuleId,
			}
			id := aws.StringValue(rule.RuleId)
			wr := newWafRegionalRetryer(conn, region)

			_, err := wr.RetryWithToken(func(token *string) (interface{}, error) {
				deleteInput.ChangeToken = token
				log.Printf("[INFO] Deleting WAF Regional Rule: %s", id)
				return conn.DeleteRule(deleteInput)
			})

			if isAWSErr(err, wafregional.ErrCodeWAFNonEmptyEntityException, "") {
				getRuleInput := &waf.GetRuleInput{
					RuleId: rule.RuleId,
				}

				getRuleOutput, getRuleErr := conn.GetRule(getRuleInput)

				if getRuleErr != nil {
					return fmt.Errorf("error getting WAF Regional Rule (%s): %s", id, getRuleErr)
				}

				var updates []*waf.RuleUpdate
				updateRuleInput := &waf.UpdateRuleInput{
					RuleId:  rule.RuleId,
					Updates: updates,
				}

				for _, predicate := range getRuleOutput.Rule.Predicates {
					update := &waf.RuleUpdate{
						Action:    aws.String(waf.ChangeActionDelete),
						Predicate: predicate,
					}

					updateRuleInput.Updates = append(updateRuleInput.Updates, update)
				}

				_, updateWebACLErr := wr.RetryWithToken(func(token *string) (interface{}, error) {
					updateRuleInput.ChangeToken = token
					log.Printf("[INFO] Removing Predicates from WAF Regional Rule: %s", id)
					return conn.UpdateRule(updateRuleInput)
				})

				if updateWebACLErr != nil {
					return fmt.Errorf("error removing predicates from WAF Regional Rule (%s): %s", id, updateWebACLErr)
				}

				_, err = wr.RetryWithToken(func(token *string) (interface{}, error) {
					deleteInput.ChangeToken = token
					log.Printf("[INFO] Deleting WAF Regional Rule: %s", id)
					return conn.DeleteRule(deleteInput)
				})
			}

			if err != nil {
				return fmt.Errorf("error deleting WAF Regional Rule (%s): %s", id, err)
			}
		}

		if aws.StringValue(output.NextMarker) == "" {
			break
		}

		input.NextMarker = output.NextMarker
	}

	return nil
}

func TestAccAWSWafRegionalRule_basic(t *testing.T) {
	var v waf.Rule
	wafRuleName := fmt.Sprintf("wafrule%s", acctest.RandString(5))
	resourceName := "aws_wafregional_rule.wafrule"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(wafregional.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafRegionalRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafRegionalRuleConfig(wafRuleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafRegionalRuleExists(resourceName, &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "waf-regional", regexp.MustCompile(`rule/.+`)),
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

func TestAccAWSWafRegionalRule_tags(t *testing.T) {
	var v waf.Rule
	wafRuleName := fmt.Sprintf("wafrule%s", acctest.RandString(5))
	resourceName := "aws_wafregional_rule.wafrule"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(wafregional.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafRegionalRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafRegionalRuleConfigTags1(wafRuleName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafRegionalRuleExists(resourceName, &v),
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
				Config: testAccAWSWafRegionalRuleConfigTags2(wafRuleName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafRegionalRuleExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSWafRegionalRuleConfigTags1(wafRuleName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafRegionalRuleExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccAWSWafRegionalRule_changeNameForceNew(t *testing.T) {
	var before, after waf.Rule
	wafRuleName := fmt.Sprintf("wafrule%s", acctest.RandString(5))
	wafRuleNewName := fmt.Sprintf("wafrulenew%s", acctest.RandString(5))
	resourceName := "aws_wafregional_rule.wafrule"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(wafregional.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafRegionalIPSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafRegionalRuleConfig(wafRuleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafRegionalRuleExists(resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, "name", wafRuleName),
					resource.TestCheckResourceAttr(resourceName, "predicate.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "metric_name", wafRuleName),
				),
			},
			{
				Config: testAccAWSWafRegionalRuleConfigChangeName(wafRuleNewName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafRegionalRuleExists(resourceName, &after),
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

func TestAccAWSWafRegionalRule_disappears(t *testing.T) {
	var v waf.Rule
	wafRuleName := fmt.Sprintf("wafrule%s", acctest.RandString(5))
	resourceName := "aws_wafregional_rule.wafrule"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(wafregional.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafRegionalRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafRegionalRuleConfig(wafRuleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafRegionalRuleExists(resourceName, &v),
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
	resourceName := "aws_wafregional_rule.wafrule"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(wafregional.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafRegionalRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafRegionalRule_noPredicates(wafRuleName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSWafRegionalRuleExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "name", wafRuleName),
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

func TestAccAWSWafRegionalRule_changePredicates(t *testing.T) {
	var ipset waf.IPSet
	var xssMatchSet waf.XssMatchSet

	var before, after waf.Rule
	var idx int
	ruleName := fmt.Sprintf("wafrule%s", acctest.RandString(5))
	resourceName := "aws_wafregional_rule.wafrule"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(wafregional.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafRegionalRuleConfig(ruleName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSWafRegionalIPSetExists("aws_wafregional_ipset.ipset", &ipset),
					testAccCheckAWSWafRegionalRuleExists(resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, "name", ruleName),
					resource.TestCheckResourceAttr(resourceName, "predicate.#", "1"),
					computeWafRegionalRulePredicate(&ipset.IPSetId, false, "IPMatch", &idx),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "predicate.*", map[string]string{
						"negated": "false",
						"type":    "IPMatch",
					}),
				),
			},
			{
				Config: testAccAWSWafRegionalRule_changePredicates(ruleName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSWafRegionalXssMatchSetExists("aws_wafregional_xss_match_set.xss_match_set", &xssMatchSet),
					testAccCheckAWSWafRegionalRuleExists(resourceName, &after),
					resource.TestCheckResourceAttr(resourceName, "name", ruleName),
					resource.TestCheckResourceAttr(resourceName, "predicate.#", "2"),
					computeWafRegionalRulePredicate(&xssMatchSet.XssMatchSetId, true, "XssMatch", &idx),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "predicate.*", map[string]string{
						"negated": "true",
						"type":    "XssMatch",
					}),
					computeWafRegionalRulePredicate(&ipset.IPSetId, true, "IPMatch", &idx),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "predicate.*", map[string]string{
						"negated": "true",
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
  name = %[1]q

  ip_set_descriptor {
    type  = "IPV4"
    value = "192.0.7.0/24"
  }
}

resource "aws_wafregional_rule" "wafrule" {
  name        = %[1]q
  metric_name = %[1]q

  predicate {
    data_id = aws_wafregional_ipset.ipset.id
    negated = false
    type    = "IPMatch"
  }
}
`, name)
}

func testAccAWSWafRegionalRuleConfigTags1(name, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_wafregional_ipset" "ipset" {
  name = %[1]q

  ip_set_descriptor {
    type  = "IPV4"
    value = "192.0.7.0/24"
  }
}

resource "aws_wafregional_rule" "wafrule" {
  name        = %[1]q
  metric_name = %[1]q

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

func testAccAWSWafRegionalRuleConfigTags2(name, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_wafregional_ipset" "ipset" {
  name = %[1]q

  ip_set_descriptor {
    type  = "IPV4"
    value = "192.0.7.0/24"
  }
}

resource "aws_wafregional_rule" "wafrule" {
  name        = %[1]q
  metric_name = %[1]q

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

func testAccAWSWafRegionalRuleConfigChangeName(name string) string {
	return fmt.Sprintf(`
resource "aws_wafregional_ipset" "ipset" {
  name = %[1]q

  ip_set_descriptor {
    type  = "IPV4"
    value = "192.0.7.0/24"
  }
}

resource "aws_wafregional_rule" "wafrule" {
  name        = %[1]q
  metric_name = %[1]q

  predicate {
    data_id = aws_wafregional_ipset.ipset.id
    negated = false
    type    = "IPMatch"
  }
}
`, name)
}

func testAccAWSWafRegionalRule_noPredicates(name string) string {
	return fmt.Sprintf(`
resource "aws_wafregional_rule" "wafrule" {
  name        = %[1]q
  metric_name = %[1]q
}
`, name)
}

func testAccAWSWafRegionalRule_changePredicates(name string) string {
	return fmt.Sprintf(`
resource "aws_wafregional_ipset" "ipset" {
  name = %[1]q

  ip_set_descriptor {
    type  = "IPV4"
    value = "192.0.7.0/24"
  }
}

resource "aws_wafregional_xss_match_set" "xss_match_set" {
  name = %[1]q

  xss_match_tuple {
    text_transformation = "NONE"

    field_to_match {
      type = "URI"
    }
  }
}

resource "aws_wafregional_rule" "wafrule" {
  name        = %[1]q
  metric_name = %[1]q

  predicate {
    data_id = aws_wafregional_xss_match_set.xss_match_set.id
    negated = true
    type    = "XssMatch"
  }

  predicate {
    data_id = aws_wafregional_ipset.ipset.id
    negated = true
    type    = "IPMatch"
  }
}
`, name)
}
