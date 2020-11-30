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
	resource.AddTestSweepers("aws_wafregional_rule_group", &resource.Sweeper{
		Name: "aws_wafregional_rule_group",
		F:    testSweepWafRegionalRuleGroups,
	})
}

func testSweepWafRegionalRuleGroups(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).wafregionalconn

	req := &waf.ListRuleGroupsInput{}
	resp, err := conn.ListRuleGroups(req)
	if err != nil {
		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping WAF Regional Rule Group sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("Error describing WAF Regional Rule Groups: %s", err)
	}

	if len(resp.RuleGroups) == 0 {
		log.Print("[DEBUG] No AWS WAF Regional Rule Groups to sweep")
		return nil
	}

	for _, group := range resp.RuleGroups {
		rResp, err := conn.ListActivatedRulesInRuleGroup(&waf.ListActivatedRulesInRuleGroupInput{
			RuleGroupId: group.RuleGroupId,
		})
		if err != nil {
			return err
		}
		oldRules := flattenWafActivatedRules(rResp.ActivatedRules)
		err = deleteWafRegionalRuleGroup(*group.RuleGroupId, oldRules, conn, region)
		if err != nil {
			return err
		}
	}

	return nil
}

func TestAccAWSWafRegionalRuleGroup_basic(t *testing.T) {
	var rule waf.Rule
	var group waf.RuleGroup
	var idx int

	ruleName := fmt.Sprintf("tfacc%s", acctest.RandString(5))
	groupName := fmt.Sprintf("tfacc%s", acctest.RandString(5))
	resourceName := "aws_wafregional_rule_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(wafregional.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafRegionalRuleGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafRegionalRuleGroupConfig(ruleName, groupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafRegionalRuleExists("aws_wafregional_rule.test", &rule),
					testAccCheckAWSWafRegionalRuleGroupExists(resourceName, &group),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "waf-regional", regexp.MustCompile(`rulegroup/.+`)),
					resource.TestCheckResourceAttr(resourceName, "name", groupName),
					resource.TestCheckResourceAttr(resourceName, "activated_rule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "metric_name", groupName),
					computeWafActivatedRuleWithRuleId(&rule, "COUNT", 50, &idx),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "activated_rule.*", map[string]string{
						"action.0.type": "COUNT",
						"priority":      "50",
						"type":          waf.WafRuleTypeRegular,
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

func TestAccAWSWafRegionalRuleGroup_tags(t *testing.T) {
	var rule waf.Rule
	var group waf.RuleGroup

	ruleName := fmt.Sprintf("tfacc%s", acctest.RandString(5))
	groupName := fmt.Sprintf("tfacc%s", acctest.RandString(5))
	resourceName := "aws_wafregional_rule_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(wafregional.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafRegionalRuleGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafRegionalRuleGroupConfigTags1(ruleName, groupName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafRegionalRuleExists("aws_wafregional_rule.test", &rule),
					testAccCheckAWSWafRegionalRuleGroupExists(resourceName, &group),
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
				Config: testAccAWSWafRegionalRuleGroupConfigTags2(ruleName, groupName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafRegionalRuleExists("aws_wafregional_rule.test", &rule),
					testAccCheckAWSWafRegionalRuleGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSWafRegionalRuleGroupConfigTags1(ruleName, groupName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafRegionalRuleExists("aws_wafregional_rule.test", &rule),
					testAccCheckAWSWafRegionalRuleGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccAWSWafRegionalRuleGroup_changeNameForceNew(t *testing.T) {
	var before, after waf.RuleGroup

	ruleName := fmt.Sprintf("tfacc%s", acctest.RandString(5))
	groupName := fmt.Sprintf("tfacc%s", acctest.RandString(5))
	newGroupName := fmt.Sprintf("tfacc%s", acctest.RandString(5))
	resourceName := "aws_wafregional_rule_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(wafregional.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafRegionalRuleGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafRegionalRuleGroupConfig(ruleName, groupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafRegionalRuleGroupExists(resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, "name", groupName),
					resource.TestCheckResourceAttr(resourceName, "activated_rule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "metric_name", groupName),
				),
			},
			{
				Config: testAccAWSWafRegionalRuleGroupConfig(ruleName, newGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafRegionalRuleGroupExists(resourceName, &after),
					resource.TestCheckResourceAttr(resourceName, "name", newGroupName),
					resource.TestCheckResourceAttr(resourceName, "activated_rule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "metric_name", newGroupName),
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

func TestAccAWSWafRegionalRuleGroup_disappears(t *testing.T) {
	var group waf.RuleGroup
	ruleName := fmt.Sprintf("tfacc%s", acctest.RandString(5))
	groupName := fmt.Sprintf("tfacc%s", acctest.RandString(5))
	resourceName := "aws_wafregional_rule_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(wafregional.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafRegionalRuleGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafRegionalRuleGroupConfig(ruleName, groupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafRegionalRuleGroupExists(resourceName, &group),
					testAccCheckAWSWafRegionalRuleGroupDisappears(&group),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSWafRegionalRuleGroup_changeActivatedRules(t *testing.T) {
	var rule0, rule1, rule2, rule3 waf.Rule
	var groupBefore, groupAfter waf.RuleGroup
	var idx0, idx1, idx2, idx3 int

	groupName := fmt.Sprintf("tfacc%s", acctest.RandString(5))
	ruleName1 := fmt.Sprintf("tfacc%s", acctest.RandString(5))
	ruleName2 := fmt.Sprintf("tfacc%s", acctest.RandString(5))
	ruleName3 := fmt.Sprintf("tfacc%s", acctest.RandString(5))
	resourceName := "aws_wafregional_rule_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(wafregional.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafRegionalRuleGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafRegionalRuleGroupConfig(ruleName1, groupName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSWafRegionalRuleExists("aws_wafregional_rule.test", &rule0),
					testAccCheckAWSWafRegionalRuleGroupExists(resourceName, &groupBefore),
					resource.TestCheckResourceAttr(resourceName, "name", groupName),
					resource.TestCheckResourceAttr(resourceName, "activated_rule.#", "1"),
					computeWafActivatedRuleWithRuleId(&rule0, "COUNT", 50, &idx0),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "activated_rule.*", map[string]string{
						"action.0.type": "COUNT",
						"priority":      "50",
						"type":          waf.WafRuleTypeRegular,
					}),
				),
			},
			{
				Config: testAccAWSWafRegionalRuleGroupConfig_changeActivatedRules(ruleName1, ruleName2, ruleName3, groupName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", groupName),
					resource.TestCheckResourceAttr(resourceName, "activated_rule.#", "3"),
					testAccCheckAWSWafRegionalRuleGroupExists(resourceName, &groupAfter),

					testAccCheckAWSWafRegionalRuleExists("aws_wafregional_rule.test", &rule1),
					computeWafActivatedRuleWithRuleId(&rule1, "BLOCK", 10, &idx1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "activated_rule.*", map[string]string{
						"action.0.type": "BLOCK",
						"priority":      "10",
						"type":          waf.WafRuleTypeRegular,
					}),

					testAccCheckAWSWafRegionalRuleExists("aws_wafregional_rule.test2", &rule2),
					computeWafActivatedRuleWithRuleId(&rule2, "COUNT", 1, &idx2),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "activated_rule.*", map[string]string{
						"action.0.type": "COUNT",
						"priority":      "1",
						"type":          waf.WafRuleTypeRegular,
					}),

					testAccCheckAWSWafRegionalRuleExists("aws_wafregional_rule.test3", &rule3),
					computeWafActivatedRuleWithRuleId(&rule3, "BLOCK", 15, &idx3),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "activated_rule.*", map[string]string{
						"action.0.type": "BLOCK",
						"priority":      "15",
						"type":          waf.WafRuleTypeRegular,
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

func TestAccAWSWafRegionalRuleGroup_noActivatedRules(t *testing.T) {
	var group waf.RuleGroup
	groupName := fmt.Sprintf("tfacc%s", acctest.RandString(5))
	resourceName := "aws_wafregional_rule_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(wafregional.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafRegionalRuleGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafRegionalRuleGroupConfig_noActivatedRules(groupName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSWafRegionalRuleGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "name", groupName),
					resource.TestCheckResourceAttr(resourceName, "activated_rule.#", "0"),
				),
			},
		},
	})
}

func testAccCheckAWSWafRegionalRuleGroupDisappears(group *waf.RuleGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).wafregionalconn
		region := testAccProvider.Meta().(*AWSClient).region

		rResp, err := conn.ListActivatedRulesInRuleGroup(&waf.ListActivatedRulesInRuleGroupInput{
			RuleGroupId: group.RuleGroupId,
		})
		if err != nil {
			return fmt.Errorf("error listing activated rules in WAF Regional Rule Group (%s): %s", aws.StringValue(group.RuleGroupId), err)
		}

		wr := newWafRegionalRetryer(conn, region)
		_, err = wr.RetryWithToken(func(token *string) (interface{}, error) {
			req := &waf.UpdateRuleGroupInput{
				ChangeToken: token,
				RuleGroupId: group.RuleGroupId,
			}

			for _, rule := range rResp.ActivatedRules {
				rule := &waf.RuleGroupUpdate{
					Action:        aws.String("DELETE"),
					ActivatedRule: rule,
				}
				req.Updates = append(req.Updates, rule)
			}

			return conn.UpdateRuleGroup(req)
		})
		if err != nil {
			return fmt.Errorf("Error Updating WAF Regional Rule Group: %s", err)
		}

		_, err = wr.RetryWithToken(func(token *string) (interface{}, error) {
			opts := &waf.DeleteRuleGroupInput{
				ChangeToken: token,
				RuleGroupId: group.RuleGroupId,
			}
			return conn.DeleteRuleGroup(opts)
		})
		if err != nil {
			return fmt.Errorf("Error Deleting WAF Regional Rule Group: %s", err)
		}
		return nil
	}
}

func testAccCheckAWSWafRegionalRuleGroupDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_wafregional_rule_group" {
			continue
		}

		conn := testAccProvider.Meta().(*AWSClient).wafregionalconn
		resp, err := conn.GetRuleGroup(&waf.GetRuleGroupInput{
			RuleGroupId: aws.String(rs.Primary.ID),
		})

		if err == nil {
			if *resp.RuleGroup.RuleGroupId == rs.Primary.ID {
				return fmt.Errorf("WAF Regional Rule Group %s still exists", rs.Primary.ID)
			}
		}

		if isAWSErr(err, wafregional.ErrCodeWAFNonexistentItemException, "") {
			return nil
		}

		return err
	}

	return nil
}

func testAccCheckAWSWafRegionalRuleGroupExists(n string, group *waf.RuleGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No WAF Regional Rule Group ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).wafregionalconn
		resp, err := conn.GetRuleGroup(&waf.GetRuleGroupInput{
			RuleGroupId: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return err
		}

		if *resp.RuleGroup.RuleGroupId == rs.Primary.ID {
			*group = *resp.RuleGroup
			return nil
		}

		return fmt.Errorf("WAF Regional Rule Group (%s) not found", rs.Primary.ID)
	}
}

func testAccAWSWafRegionalRuleGroupConfig(ruleName, groupName string) string {
	return fmt.Sprintf(`
resource "aws_wafregional_rule" "test" {
  name        = "%[1]s"
  metric_name = "%[1]s"
}

resource "aws_wafregional_rule_group" "test" {
  name        = "%[2]s"
  metric_name = "%[2]s"

  activated_rule {
    action {
      type = "COUNT"
    }

    priority = 50
    rule_id  = aws_wafregional_rule.test.id
  }
}
`, ruleName, groupName)
}

func testAccAWSWafRegionalRuleGroupConfigTags1(ruleName, groupName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_wafregional_rule" "test" {
  name        = %[1]q
  metric_name = %[1]q
}

resource "aws_wafregional_rule_group" "test" {
  name        = %[2]q
  metric_name = %[2]q

  activated_rule {
    action {
      type = "COUNT"
    }

    priority = 50
    rule_id  = aws_wafregional_rule.test.id
  }

  tags = {
    %[3]q = %[4]q
  }
}
`, ruleName, groupName, tagKey1, tagValue1)
}

func testAccAWSWafRegionalRuleGroupConfigTags2(ruleName, groupName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_wafregional_rule" "test" {
  name        = %[1]q
  metric_name = %[1]q
}

resource "aws_wafregional_rule_group" "test" {
  name        = %[2]q
  metric_name = %[2]q

  activated_rule {
    action {
      type = "COUNT"
    }

    priority = 50
    rule_id  = aws_wafregional_rule.test.id
  }

  tags = {
    %[3]q = %[4]q
    %[5]q = %[6]q
  }
}
`, ruleName, groupName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccAWSWafRegionalRuleGroupConfig_changeActivatedRules(ruleName1, ruleName2, ruleName3, groupName string) string {
	return fmt.Sprintf(`
resource "aws_wafregional_rule" "test" {
  name        = "%[1]s"
  metric_name = "%[1]s"
}

resource "aws_wafregional_rule" "test2" {
  name        = "%[2]s"
  metric_name = "%[2]s"
}

resource "aws_wafregional_rule" "test3" {
  name        = "%[3]s"
  metric_name = "%[3]s"
}

resource "aws_wafregional_rule_group" "test" {
  name        = "%[4]s"
  metric_name = "%[4]s"

  activated_rule {
    action {
      type = "BLOCK"
    }

    priority = 10
    rule_id  = aws_wafregional_rule.test.id
  }

  activated_rule {
    action {
      type = "COUNT"
    }

    priority = 1
    rule_id  = aws_wafregional_rule.test2.id
  }

  activated_rule {
    action {
      type = "BLOCK"
    }

    priority = 15
    rule_id  = aws_wafregional_rule.test3.id
  }
}
`, ruleName1, ruleName2, ruleName3, groupName)
}

func testAccAWSWafRegionalRuleGroupConfig_noActivatedRules(groupName string) string {
	return fmt.Sprintf(`
resource "aws_wafregional_rule_group" "test" {
  name        = "%[1]s"
  metric_name = "%[1]s"
}
`, groupName)
}
