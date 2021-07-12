package aws

import (
	"fmt"
	"log"
	"regexp"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/gamelift"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

const testAccGameliftMatchmakingRuleSetPrefix = "tfAccRuleSet-"

func init() {
	resource.AddTestSweepers("aws_gamelift_matchmaking_rule_set", &resource.Sweeper{
		Name: "aws_gamelift_matchmaking_rule_set",
		F:    testSweepGameliftMatchmakingRuleSet,
	})
}

func testSweepGameliftMatchmakingRuleSet(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).gameliftconn

	out, err := conn.DescribeMatchmakingRuleSets(&gamelift.DescribeMatchmakingRuleSetsInput{})

	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping Gamelift Matchmaking Rule Set sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing Gamelift Rule Set: %s", err)
	}

	if len(out.RuleSets) == 0 {
		log.Print("[DEBUG] No Gamelift Matchmaking Rule Set to sweep")
		return nil
	}

	log.Printf("[INFO] Found %d Gamelift Matchmaking Rule Sets", len(out.RuleSets))

	for _, ruleSet := range out.RuleSets {
		log.Printf("[INFO] Deleting Gamelift Matchmaking Rule Set %q", *ruleSet.RuleSetName)
		_, err := conn.DeleteMatchmakingRuleSet(&gamelift.DeleteMatchmakingRuleSetInput{
			Name: aws.String(*ruleSet.RuleSetName),
		})
		if err != nil {
			return fmt.Errorf("error deleting Gamelift Matchmaking Rule Set (%s): %s",
				*ruleSet.RuleSetName, err)
		}
	}

	return nil
}

func TestAccAWSGameliftMatchmakingRuleSet_basic(t *testing.T) {
	var conf gamelift.MatchmakingRuleSet

	resourceName := "aws_gamelift_matchmaking_rule_set.test"
	ruleSetName := testAccGameliftMatchmakingRuleSetPrefix + acctest.RandString(8)
	maxPlayers := 5

	uRuleSetName := ruleSetName + "-updated"
	uMaxPlayers := 10

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSGamelift(t) },
		Providers:    testAccProviders,
		ErrorCheck:   testAccErrorCheck(t, gamelift.EndpointsID),
		CheckDestroy: testAccCheckAWSGameliftMatchmakingRuleSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSGameliftMatchmakingRuleSetBasicConfig(ruleSetName, maxPlayers),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGameliftMatchmakingRuleSetExists(resourceName, &conf),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "gamelift", regexp.MustCompile(`matchmakingruleset/.+`)),
					resource.TestCheckResourceAttr(resourceName, "name", ruleSetName),
					resource.TestCheckResourceAttr(resourceName, "rule_set_body", testAccAWSGameliftMatchmakingRuleSetBody(maxPlayers)+"\n"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				Config: testAccAWSGameliftMatchmakingRuleSetBasicConfig(uRuleSetName, uMaxPlayers),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGameliftMatchmakingRuleSetExists(resourceName, &conf),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "gamelift", regexp.MustCompile(`matchmakingruleset/.+`)),
					resource.TestCheckResourceAttr(resourceName, "name", uRuleSetName),
					resource.TestCheckResourceAttr(resourceName, "rule_set_body", testAccAWSGameliftMatchmakingRuleSetBody(uMaxPlayers)+"\n"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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

func TestAccAWSGameliftMatchmakingRuleSet_tags(t *testing.T) {
	var conf gamelift.MatchmakingRuleSet

	resourceName := "aws_gamelift_matchmaking_rule_set.test"
	ruleSetName := testAccGameliftMatchmakingRuleSetPrefix + acctest.RandString(8)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSGamelift(t) },
		Providers:    testAccProviders,
		ErrorCheck:   testAccErrorCheck(t, gamelift.EndpointsID),
		CheckDestroy: testAccCheckAWSGameliftMatchmakingRuleSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSGameliftMatchmakingRuleSetBasicConfigTags1(ruleSetName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGameliftMatchmakingRuleSetExists(resourceName, &conf),
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
				Config: testAccAWSGameliftMatchmakingRuleSetBasicConfigTags2(ruleSetName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGameliftMatchmakingRuleSetExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSGameliftMatchmakingRuleSetBasicConfigTags1(ruleSetName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGameliftMatchmakingRuleSetExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccAWSGameliftMatchmakingRuleSet_disappears(t *testing.T) {
	var conf gamelift.MatchmakingRuleSet

	resourceName := "aws_gamelift_matchmaking_rule_set.test"
	ruleSetName := testAccGameliftMatchmakingRuleSetPrefix + acctest.RandString(8)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSGamelift(t) },
		Providers:    testAccProviders,
		ErrorCheck:   testAccErrorCheck(t, gamelift.EndpointsID),
		CheckDestroy: testAccCheckAWSGameliftMatchmakingRuleSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSGameliftMatchmakingRuleSetBasicConfig(ruleSetName, 5),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGameliftMatchmakingRuleSetExists(resourceName, &conf),
					testAccCheckAWSGameliftMatchmakingRuleSetDisappears(&conf),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAWSGameliftMatchmakingRuleSetExists(n string, res *gamelift.MatchmakingRuleSet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no Gamelift Matchmaking Rule Set Name is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).gameliftconn

		name := rs.Primary.Attributes["name"]
		out, err := conn.DescribeMatchmakingRuleSets(&gamelift.DescribeMatchmakingRuleSetsInput{
			Names: aws.StringSlice([]string{name}),
		})
		if err != nil {
			return err
		}
		ruleSets := out.RuleSets
		if len(ruleSets) == 0 {
			return fmt.Errorf("GameLift Matchmaking Rule Set %q not found", name)
		}

		*res = *ruleSets[0]

		return nil
	}
}

func testAccCheckAWSGameliftMatchmakingRuleSetDisappears(res *gamelift.MatchmakingRuleSet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).gameliftconn

		input := &gamelift.DeleteMatchmakingRuleSetInput{Name: res.RuleSetName}

		_, err := conn.DeleteMatchmakingRuleSet(input)

		return err
	}
}

func testAccCheckAWSGameliftMatchmakingRuleSetDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).gameliftconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_gamelift_matchmaking_rule_set" {
			continue
		}

		name := rs.Primary.Attributes["name"]

		input := &gamelift.DescribeMatchmakingRuleSetsInput{
			Names: aws.StringSlice([]string{name}),
		}

		// Deletions can take a few seconds
		err := resource.Retry(30*time.Second, func() *resource.RetryError {
			out, err := conn.DescribeMatchmakingRuleSets(input)

			if isAWSErr(err, gamelift.ErrCodeInvalidRequestException, "Failed to find rule set") {
				return nil
			}

			if err != nil {
				return resource.NonRetryableError(err)
			}

			ruleSets := out.RuleSets

			if len(ruleSets) > 0 {
				return resource.RetryableError(fmt.Errorf("GameLift Matchmaking Rule Set still exists"))
			}

			return nil
		})

		if err != nil {
			return err
		}
	}

	return nil
}

func testAccAWSGameliftMatchmakingRuleSetBody(maxPlayers int) string {
	return fmt.Sprintf(`{
	"name": "test",
	"ruleLanguageVersion": "1.0",
	"teams": [{
		"name": "alpha",
		"minPlayers": 1,
		"maxPlayers": %[1]d
	}]
}`, maxPlayers)
}

func testAccAWSGameliftMatchmakingRuleSetBasicConfig(rName string, maxPlayers int) string {
	return fmt.Sprintf(`
resource "aws_gamelift_matchmaking_rule_set" "test" {
  name          = %[1]q
  rule_set_body = <<RULE_SET_BODY
%[2]s
RULE_SET_BODY
}
`, rName, testAccAWSGameliftMatchmakingRuleSetBody(maxPlayers))
}

func testAccAWSGameliftMatchmakingRuleSetBasicConfigTags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_gamelift_matchmaking_rule_set" "test" {
  name          = %[1]q
  rule_set_body = <<RULE_SET_BODY
%[2]s
RULE_SET_BODY
  tags = {
    %[3]q = %[4]q
  }
}
`, rName, testAccAWSGameliftMatchmakingRuleSetBody(1), tagKey1, tagValue1)
}

func testAccAWSGameliftMatchmakingRuleSetBasicConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_gamelift_matchmaking_rule_set" "test" {
  name          = %[1]q
  rule_set_body = <<RULE_SET_BODY
%[2]s
RULE_SET_BODY
  tags = {
    %[3]q = %[4]q
    %[5]q = %[6]q
  }
}
`, rName, testAccAWSGameliftMatchmakingRuleSetBody(1), tagKey1, tagValue1, tagKey2, tagValue2)
}
