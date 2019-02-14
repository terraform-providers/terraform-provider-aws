package aws

import (
	"fmt"
	"log"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/waf"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

func init() {
	resource.AddTestSweepers("aws_waf_rule_group", &resource.Sweeper{
		Name: "aws_waf_rule_group",
		F:    testSweepWafRuleGroups,
	})
}

func testSweepWafRuleGroups(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).wafconn

	req := &waf.ListRuleGroupsInput{}
	resp, err := conn.ListRuleGroups(req)
	if err != nil {
		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping WAF Rule Group sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("Error describing WAF Rule Groups: %s", err)
	}

	if len(resp.RuleGroups) == 0 {
		log.Print("[DEBUG] No AWS WAF Rule Groups to sweep")
		return nil
	}

	for _, group := range resp.RuleGroups {
		if !strings.HasPrefix(*group.Name, "tfacc") {
			continue
		}

		rResp, err := conn.ListActivatedRulesInRuleGroup(&waf.ListActivatedRulesInRuleGroupInput{
			RuleGroupId: group.RuleGroupId,
		})
		if err != nil {
			return err
		}
		oldRules := flattenWafActivatedRules(rResp.ActivatedRules)
		err = deleteWafRuleGroup(*group.RuleGroupId, oldRules, conn)
		if err != nil {
			return err
		}
	}

	return nil
}

func TestAccAWSWafRuleGroup_basic(t *testing.T) {
	var rule waf.Rule
	var group waf.RuleGroup
	var idx int

	ruleName := fmt.Sprintf("tfacc%s", acctest.RandString(5))
	groupName := fmt.Sprintf("tfacc%s", acctest.RandString(5))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafRuleGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafRuleGroupConfig(ruleName, groupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafRuleExists("aws_waf_rule.test", &rule),
					testAccCheckAWSWafRuleGroupExists("aws_waf_rule_group.test", &group),
					resource.TestCheckResourceAttr("aws_waf_rule_group.test", "name", groupName),
					resource.TestCheckResourceAttr("aws_waf_rule_group.test", "activated_rule.#", "1"),
					resource.TestCheckResourceAttr("aws_waf_rule_group.test", "metric_name", groupName),
					computeWafActivatedRuleWithRuleId(&rule, "COUNT", 50, &idx),
					testCheckResourceAttrWithIndexesAddr("aws_waf_rule_group.test", "activated_rule.%d.action.0.type", &idx, "COUNT"),
					testCheckResourceAttrWithIndexesAddr("aws_waf_rule_group.test", "activated_rule.%d.priority", &idx, "50"),
					testCheckResourceAttrWithIndexesAddr("aws_waf_rule_group.test", "activated_rule.%d.type", &idx, waf.WafRuleTypeRegular),
				),
			},
		},
	})
}

func TestAccAWSWafRuleGroup_changeNameForceNew(t *testing.T) {
	var before, after waf.RuleGroup

	ruleName := fmt.Sprintf("tfacc%s", acctest.RandString(5))
	groupName := fmt.Sprintf("tfacc%s", acctest.RandString(5))
	newGroupName := fmt.Sprintf("tfacc%s", acctest.RandString(5))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafRuleGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafRuleGroupConfig(ruleName, groupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafRuleGroupExists("aws_waf_rule_group.test", &before),
					resource.TestCheckResourceAttr("aws_waf_rule_group.test", "name", groupName),
					resource.TestCheckResourceAttr("aws_waf_rule_group.test", "activated_rule.#", "1"),
					resource.TestCheckResourceAttr("aws_waf_rule_group.test", "metric_name", groupName),
				),
			},
			{
				Config: testAccAWSWafRuleGroupConfig(ruleName, newGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafRuleGroupExists("aws_waf_rule_group.test", &after),
					resource.TestCheckResourceAttr("aws_waf_rule_group.test", "name", newGroupName),
					resource.TestCheckResourceAttr("aws_waf_rule_group.test", "activated_rule.#", "1"),
					resource.TestCheckResourceAttr("aws_waf_rule_group.test", "metric_name", newGroupName),
				),
			},
		},
	})
}

func TestAccAWSWafRuleGroup_disappears(t *testing.T) {
	var group waf.RuleGroup
	ruleName := fmt.Sprintf("tfacc%s", acctest.RandString(5))
	groupName := fmt.Sprintf("tfacc%s", acctest.RandString(5))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafRuleGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafRuleGroupConfig(ruleName, groupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafRuleGroupExists("aws_waf_rule_group.test", &group),
					testAccCheckAWSWafRuleGroupDisappears(&group),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSWafRuleGroup_changeActivatedRules(t *testing.T) {
	var rule0, rule1, rule2, rule3 waf.Rule
	var groupBefore, groupAfter waf.RuleGroup
	var idx0, idx1, idx2, idx3 int

	groupName := fmt.Sprintf("tfacc%s", acctest.RandString(5))
	ruleName1 := fmt.Sprintf("tfacc%s", acctest.RandString(5))
	ruleName2 := fmt.Sprintf("tfacc%s", acctest.RandString(5))
	ruleName3 := fmt.Sprintf("tfacc%s", acctest.RandString(5))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafRuleGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafRuleGroupConfig(ruleName1, groupName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSWafRuleExists("aws_waf_rule.test", &rule0),
					testAccCheckAWSWafRuleGroupExists("aws_waf_rule_group.test", &groupBefore),
					resource.TestCheckResourceAttr("aws_waf_rule_group.test", "name", groupName),
					resource.TestCheckResourceAttr("aws_waf_rule_group.test", "activated_rule.#", "1"),
					computeWafActivatedRuleWithRuleId(&rule0, "COUNT", 50, &idx0),
					testCheckResourceAttrWithIndexesAddr("aws_waf_rule_group.test", "activated_rule.%d.action.0.type", &idx0, "COUNT"),
					testCheckResourceAttrWithIndexesAddr("aws_waf_rule_group.test", "activated_rule.%d.priority", &idx0, "50"),
					testCheckResourceAttrWithIndexesAddr("aws_waf_rule_group.test", "activated_rule.%d.type", &idx0, waf.WafRuleTypeRegular),
				),
			},
			{
				Config: testAccAWSWafRuleGroupConfig_changeActivatedRules(ruleName1, ruleName2, ruleName3, groupName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("aws_waf_rule_group.test", "name", groupName),
					resource.TestCheckResourceAttr("aws_waf_rule_group.test", "activated_rule.#", "3"),
					testAccCheckAWSWafRuleGroupExists("aws_waf_rule_group.test", &groupAfter),

					testAccCheckAWSWafRuleExists("aws_waf_rule.test", &rule1),
					computeWafActivatedRuleWithRuleId(&rule1, "BLOCK", 10, &idx1),
					testCheckResourceAttrWithIndexesAddr("aws_waf_rule_group.test", "activated_rule.%d.action.0.type", &idx1, "BLOCK"),
					testCheckResourceAttrWithIndexesAddr("aws_waf_rule_group.test", "activated_rule.%d.priority", &idx1, "10"),
					testCheckResourceAttrWithIndexesAddr("aws_waf_rule_group.test", "activated_rule.%d.type", &idx1, waf.WafRuleTypeRegular),

					testAccCheckAWSWafRuleExists("aws_waf_rule.test2", &rule2),
					computeWafActivatedRuleWithRuleId(&rule2, "COUNT", 1, &idx2),
					testCheckResourceAttrWithIndexesAddr("aws_waf_rule_group.test", "activated_rule.%d.action.0.type", &idx2, "COUNT"),
					testCheckResourceAttrWithIndexesAddr("aws_waf_rule_group.test", "activated_rule.%d.priority", &idx2, "1"),
					testCheckResourceAttrWithIndexesAddr("aws_waf_rule_group.test", "activated_rule.%d.type", &idx2, waf.WafRuleTypeRegular),

					testAccCheckAWSWafRuleExists("aws_waf_rule.test3", &rule3),
					computeWafActivatedRuleWithRuleId(&rule3, "BLOCK", 15, &idx3),
					testCheckResourceAttrWithIndexesAddr("aws_waf_rule_group.test", "activated_rule.%d.action.0.type", &idx3, "BLOCK"),
					testCheckResourceAttrWithIndexesAddr("aws_waf_rule_group.test", "activated_rule.%d.priority", &idx3, "15"),
					testCheckResourceAttrWithIndexesAddr("aws_waf_rule_group.test", "activated_rule.%d.type", &idx3, waf.WafRuleTypeRegular),
				),
			},
		},
	})
}

// computeWafActivatedRuleWithRuleId calculates index
// which isn't static because ruleId is generated as part of the test
func computeWafActivatedRuleWithRuleId(rule *waf.Rule, actionType string, priority int, idx *int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ruleResource := resourceAwsWafRuleGroup().Schema["activated_rule"].Elem.(*schema.Resource)

		m := map[string]interface{}{
			"action": []interface{}{
				map[string]interface{}{
					"type": actionType,
				},
			},
			"priority": priority,
			"rule_id":  *rule.RuleId,
			"type":     waf.WafRuleTypeRegular,
		}

		f := schema.HashResource(ruleResource)
		*idx = f(m)

		return nil
	}
}

func TestAccAWSWafRuleGroup_noActivatedRules(t *testing.T) {
	var group waf.RuleGroup
	groupName := fmt.Sprintf("test%s", acctest.RandString(5))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafRuleGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafRuleGroupConfig_noActivatedRules(groupName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSWafRuleGroupExists("aws_waf_rule_group.test", &group),
					resource.TestCheckResourceAttr(
						"aws_waf_rule_group.test", "name", groupName),
					resource.TestCheckResourceAttr(
						"aws_waf_rule_group.test", "activated_rule.#", "0"),
				),
			},
		},
	})
}

func testAccCheckAWSWafRuleGroupDisappears(group *waf.RuleGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).wafconn

		rResp, err := conn.ListActivatedRulesInRuleGroup(&waf.ListActivatedRulesInRuleGroupInput{
			RuleGroupId: group.RuleGroupId,
		})
		if err != nil {
			return fmt.Errorf("error listing activated rules in WAF Rule Group (%s): %s", aws.StringValue(group.RuleGroupId), err)
		}

		wr := newWafRetryer(conn)
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
			return fmt.Errorf("Error Updating WAF Rule Group: %s", err)
		}

		_, err = wr.RetryWithToken(func(token *string) (interface{}, error) {
			opts := &waf.DeleteRuleGroupInput{
				ChangeToken: token,
				RuleGroupId: group.RuleGroupId,
			}
			return conn.DeleteRuleGroup(opts)
		})
		if err != nil {
			return fmt.Errorf("Error Deleting WAF Rule Group: %s", err)
		}
		return nil
	}
}

func testAccCheckAWSWafRuleGroupDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_waf_rule_group" {
			continue
		}

		conn := testAccProvider.Meta().(*AWSClient).wafconn
		resp, err := conn.GetRuleGroup(&waf.GetRuleGroupInput{
			RuleGroupId: aws.String(rs.Primary.ID),
		})

		if err == nil {
			if *resp.RuleGroup.RuleGroupId == rs.Primary.ID {
				return fmt.Errorf("WAF Rule Group %s still exists", rs.Primary.ID)
			}
		}

		if isAWSErr(err, "WAFNonexistentItemException", "") {
			return nil
		}

		return err
	}

	return nil
}

func testAccCheckAWSWafRuleGroupExists(n string, group *waf.RuleGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No WAF Rule Group ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).wafconn
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

		return fmt.Errorf("WAF Rule Group (%s) not found", rs.Primary.ID)
	}
}

func testAccAWSWafRuleGroupConfig(ruleName, groupName string) string {
	return fmt.Sprintf(`
resource "aws_waf_rule" "test" {
  name        = "%[1]s"
  metric_name = "%[1]s"
}

resource "aws_waf_rule_group" "test" {
  name = "%[2]s"
  metric_name = "%[2]s"
  activated_rule {
  	action {
      type = "COUNT"
    }
    priority = 50
    rule_id = "${aws_waf_rule.test.id}"
  }
}`, ruleName, groupName)
}

func testAccAWSWafRuleGroupConfig_changeActivatedRules(ruleName1, ruleName2, ruleName3, groupName string) string {
	return fmt.Sprintf(`
resource "aws_waf_rule" "test" {
  name        = "%[1]s"
  metric_name = "%[1]s"
}

resource "aws_waf_rule" "test2" {
  name        = "%[2]s"
  metric_name = "%[2]s"
}

resource "aws_waf_rule" "test3" {
  name        = "%[3]s"
  metric_name = "%[3]s"
}

resource "aws_waf_rule_group" "test" {
  name = "%[4]s"
  metric_name = "%[4]s"
  activated_rule {
    action {
      type = "BLOCK"
    }
    priority = 10
    rule_id = "${aws_waf_rule.test.id}"
  }
  activated_rule {
  	action {
      type = "COUNT"
    }
    priority = 1
    rule_id = "${aws_waf_rule.test2.id}"
  }
  activated_rule {
  	action {
      type = "BLOCK"
    }
    priority = 15
    rule_id = "${aws_waf_rule.test3.id}"
  }
}`, ruleName1, ruleName2, ruleName3, groupName)
}

func testAccAWSWafRuleGroupConfig_noActivatedRules(groupName string) string {
	return fmt.Sprintf(`
resource "aws_waf_rule_group" "test" {
  name = "%[1]s"
  metric_name = "%[1]s"
}`, groupName)
}
