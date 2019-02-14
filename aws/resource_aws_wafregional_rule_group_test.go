package aws

import (
	"fmt"
	"log"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/waf"
	"github.com/aws/aws-sdk-go/service/wafregional"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
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

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafRegionalRuleGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafRegionalRuleGroupConfig(ruleName, groupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafRegionalRuleExists("aws_wafregional_rule.test", &rule),
					testAccCheckAWSWafRegionalRuleGroupExists("aws_wafregional_rule_group.test", &group),
					resource.TestCheckResourceAttr("aws_wafregional_rule_group.test", "name", groupName),
					resource.TestCheckResourceAttr("aws_wafregional_rule_group.test", "activated_rule.#", "1"),
					resource.TestCheckResourceAttr("aws_wafregional_rule_group.test", "metric_name", groupName),
					computeWafActivatedRuleWithRuleId(&rule, "COUNT", 50, &idx),
					testCheckResourceAttrWithIndexesAddr("aws_wafregional_rule_group.test", "activated_rule.%d.action.0.type", &idx, "COUNT"),
					testCheckResourceAttrWithIndexesAddr("aws_wafregional_rule_group.test", "activated_rule.%d.priority", &idx, "50"),
					testCheckResourceAttrWithIndexesAddr("aws_wafregional_rule_group.test", "activated_rule.%d.type", &idx, waf.WafRuleTypeRegular),
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

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafRegionalRuleGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafRegionalRuleGroupConfig(ruleName, groupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafRegionalRuleGroupExists("aws_wafregional_rule_group.test", &before),
					resource.TestCheckResourceAttr("aws_wafregional_rule_group.test", "name", groupName),
					resource.TestCheckResourceAttr("aws_wafregional_rule_group.test", "activated_rule.#", "1"),
					resource.TestCheckResourceAttr("aws_wafregional_rule_group.test", "metric_name", groupName),
				),
			},
			{
				Config: testAccAWSWafRegionalRuleGroupConfig(ruleName, newGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafRegionalRuleGroupExists("aws_wafregional_rule_group.test", &after),
					resource.TestCheckResourceAttr("aws_wafregional_rule_group.test", "name", newGroupName),
					resource.TestCheckResourceAttr("aws_wafregional_rule_group.test", "activated_rule.#", "1"),
					resource.TestCheckResourceAttr("aws_wafregional_rule_group.test", "metric_name", newGroupName),
				),
			},
		},
	})
}

func TestAccAWSWafRegionalRuleGroup_disappears(t *testing.T) {
	var group waf.RuleGroup
	ruleName := fmt.Sprintf("tfacc%s", acctest.RandString(5))
	groupName := fmt.Sprintf("tfacc%s", acctest.RandString(5))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafRegionalRuleGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafRegionalRuleGroupConfig(ruleName, groupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafRegionalRuleGroupExists("aws_wafregional_rule_group.test", &group),
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

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafRegionalRuleGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafRegionalRuleGroupConfig(ruleName1, groupName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSWafRegionalRuleExists("aws_wafregional_rule.test", &rule0),
					testAccCheckAWSWafRegionalRuleGroupExists("aws_wafregional_rule_group.test", &groupBefore),
					resource.TestCheckResourceAttr("aws_wafregional_rule_group.test", "name", groupName),
					resource.TestCheckResourceAttr("aws_wafregional_rule_group.test", "activated_rule.#", "1"),
					computeWafActivatedRuleWithRuleId(&rule0, "COUNT", 50, &idx0),
					testCheckResourceAttrWithIndexesAddr("aws_wafregional_rule_group.test", "activated_rule.%d.action.0.type", &idx0, "COUNT"),
					testCheckResourceAttrWithIndexesAddr("aws_wafregional_rule_group.test", "activated_rule.%d.priority", &idx0, "50"),
					testCheckResourceAttrWithIndexesAddr("aws_wafregional_rule_group.test", "activated_rule.%d.type", &idx0, waf.WafRuleTypeRegular),
				),
			},
			{
				Config: testAccAWSWafRegionalRuleGroupConfig_changeActivatedRules(ruleName1, ruleName2, ruleName3, groupName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("aws_wafregional_rule_group.test", "name", groupName),
					resource.TestCheckResourceAttr("aws_wafregional_rule_group.test", "activated_rule.#", "3"),
					testAccCheckAWSWafRegionalRuleGroupExists("aws_wafregional_rule_group.test", &groupAfter),

					testAccCheckAWSWafRegionalRuleExists("aws_wafregional_rule.test", &rule1),
					computeWafActivatedRuleWithRuleId(&rule1, "BLOCK", 10, &idx1),
					testCheckResourceAttrWithIndexesAddr("aws_wafregional_rule_group.test", "activated_rule.%d.action.0.type", &idx1, "BLOCK"),
					testCheckResourceAttrWithIndexesAddr("aws_wafregional_rule_group.test", "activated_rule.%d.priority", &idx1, "10"),
					testCheckResourceAttrWithIndexesAddr("aws_wafregional_rule_group.test", "activated_rule.%d.type", &idx1, waf.WafRuleTypeRegular),

					testAccCheckAWSWafRegionalRuleExists("aws_wafregional_rule.test2", &rule2),
					computeWafActivatedRuleWithRuleId(&rule2, "COUNT", 1, &idx2),
					testCheckResourceAttrWithIndexesAddr("aws_wafregional_rule_group.test", "activated_rule.%d.action.0.type", &idx2, "COUNT"),
					testCheckResourceAttrWithIndexesAddr("aws_wafregional_rule_group.test", "activated_rule.%d.priority", &idx2, "1"),
					testCheckResourceAttrWithIndexesAddr("aws_wafregional_rule_group.test", "activated_rule.%d.type", &idx2, waf.WafRuleTypeRegular),

					testAccCheckAWSWafRegionalRuleExists("aws_wafregional_rule.test3", &rule3),
					computeWafActivatedRuleWithRuleId(&rule3, "BLOCK", 15, &idx3),
					testCheckResourceAttrWithIndexesAddr("aws_wafregional_rule_group.test", "activated_rule.%d.action.0.type", &idx3, "BLOCK"),
					testCheckResourceAttrWithIndexesAddr("aws_wafregional_rule_group.test", "activated_rule.%d.priority", &idx3, "15"),
					testCheckResourceAttrWithIndexesAddr("aws_wafregional_rule_group.test", "activated_rule.%d.type", &idx3, waf.WafRuleTypeRegular),
				),
			},
		},
	})
}

func TestAccAWSWafRegionalRuleGroup_noActivatedRules(t *testing.T) {
	var group waf.RuleGroup
	groupName := fmt.Sprintf("tfacc%s", acctest.RandString(5))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafRegionalRuleGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafRegionalRuleGroupConfig_noActivatedRules(groupName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSWafRegionalRuleGroupExists("aws_wafregional_rule_group.test", &group),
					resource.TestCheckResourceAttr(
						"aws_wafregional_rule_group.test", "name", groupName),
					resource.TestCheckResourceAttr(
						"aws_wafregional_rule_group.test", "activated_rule.#", "0"),
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
  name = "%[2]s"
  metric_name = "%[2]s"
  activated_rule {
  	action {
      type = "COUNT"
    }
    priority = 50
    rule_id = "${aws_wafregional_rule.test.id}"
  }
}`, ruleName, groupName)
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
  name = "%[4]s"
  metric_name = "%[4]s"
  activated_rule {
    action {
      type = "BLOCK"
    }
    priority = 10
    rule_id = "${aws_wafregional_rule.test.id}"
  }
  activated_rule {
  	action {
      type = "COUNT"
    }
    priority = 1
    rule_id = "${aws_wafregional_rule.test2.id}"
  }
  activated_rule {
  	action {
      type = "BLOCK"
    }
    priority = 15
    rule_id = "${aws_wafregional_rule.test3.id}"
  }
}`, ruleName1, ruleName2, ruleName3, groupName)
}

func testAccAWSWafRegionalRuleGroupConfig_noActivatedRules(groupName string) string {
	return fmt.Sprintf(`
resource "aws_wafregional_rule_group" "test" {
  name = "%[1]s"
  metric_name = "%[1]s"
}`, groupName)
}
