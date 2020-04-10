package aws

import (
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	events "github.com/aws/aws-sdk-go/service/cloudwatchevents"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func init() {
	resource.AddTestSweepers("aws_cloudwatch_event_rule", &resource.Sweeper{
		Name: "aws_cloudwatch_event_rule",
		F:    testSweepCloudWatchEventRules,
		Dependencies: []string{
			"aws_cloudwatch_event_target",
		},
	})
}

func testSweepCloudWatchEventRules(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("Error getting client: %s", err)
	}
	conn := client.(*AWSClient).cloudwatcheventsconn

	input := &events.ListRulesInput{}

	for {
		output, err := conn.ListRules(input)
		if err != nil {
			if testSweepSkipSweepError(err) {
				log.Printf("[WARN] Skipping CloudWatch Event Rule sweep for %s: %s", region, err)
				return nil
			}
			return fmt.Errorf("Error retrieving CloudWatch Event Rules: %s", err)
		}

		if len(output.Rules) == 0 {
			log.Print("[DEBUG] No CloudWatch Event Rules to sweep")
			return nil
		}

		for _, rule := range output.Rules {
			name := aws.StringValue(rule.Name)

			log.Printf("[INFO] Deleting CloudWatch Event Rule %s", name)
			_, err := conn.DeleteRule(&events.DeleteRuleInput{
				Name:  aws.String(name),
				Force: aws.Bool(true),
			})
			if err != nil {
				return fmt.Errorf("Error deleting CloudWatch Event Rule %s: %s", name, err)
			}
		}

		if output.NextToken == nil {
			break
		}
		input.NextToken = output.NextToken
	}

	return nil
}

func TestAccAWSCloudWatchEventRule_basic(t *testing.T) {
	var rule events.DescribeRuleOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	rName2 := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_cloudwatch_event_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudWatchEventRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudWatchEventRuleConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchEventRuleExists(resourceName, &rule),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "events", regexp.MustCompile(`rule/.+`)),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "schedule_expression", "rate(1 hour)"),
					resource.TestCheckResourceAttr(resourceName, "role_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "is_enabled", "true"),
					testAccCheckCloudWatchEventRuleEnabled(resourceName, "ENABLED"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"is_enabled"}, //this has a default value
			},
			{
				Config: testAccAWSCloudWatchEventRuleConfig(rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchEventRuleExists(resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "name", rName2),
					resource.TestCheckResourceAttr(resourceName, "schedule_expression", "rate(1 hour)"),
					resource.TestCheckResourceAttr(resourceName, "role_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "is_enabled", "true"),
					testAccCheckCloudWatchEventRuleEnabled(resourceName, "ENABLED"),
				),
			},
		},
	})
}

func TestAccAWSCloudWatchEventRule_role(t *testing.T) {
	var rule events.DescribeRuleOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_cloudwatch_event_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudWatchEventRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudWatchEventRuleConfigRole(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchEventRuleExists(resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					testAccMatchResourceAttrGlobalARN(resourceName, "role_arn", "iam", regexp.MustCompile(`role/.+`)),
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

func TestAccAWSCloudWatchEventRule_description(t *testing.T) {
	var rule events.DescribeRuleOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_cloudwatch_event_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudWatchEventRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudWatchEventRuleConfigDescription(rName, "description1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchEventRuleExists(resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "description", "description1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSCloudWatchEventRuleConfigDescription(rName, "description2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchEventRuleExists(resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "description", "description2"),
				),
			},
		},
	})
}

func TestAccAWSCloudWatchEventRule_pattern(t *testing.T) {
	var rule events.DescribeRuleOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_cloudwatch_event_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudWatchEventRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudWatchEventRuleConfigPattern(rName, "{\"source\":[\"aws.ec2\"]}"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchEventRuleExists(resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "event_pattern", "{\"source\":[\"aws.ec2\"]}"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSCloudWatchEventRuleConfigPattern(rName, "{\"source\":[\"aws.lambda\"]}"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchEventRuleExists(resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "event_pattern", "{\"source\":[\"aws.lambda\"]}"),
				),
			},
		},
	})
}

func TestAccAWSCloudWatchEventRule_prefix(t *testing.T) {
	var rule events.DescribeRuleOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	startsWithPrefix := regexp.MustCompile(rName)
	resourceName := "aws_cloudwatch_event_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudWatchEventRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudWatchEventRuleConfigPrefix(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchEventRuleExists(resourceName, &rule),
					resource.TestMatchResourceAttr(resourceName, "name", startsWithPrefix),
				),
			},
		},
	})
}

func TestAccAWSCloudWatchEventRule_tags(t *testing.T) {
	var rule events.DescribeRuleOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_cloudwatch_event_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudWatchEventRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudWatchEventRuleConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchEventRuleExists(resourceName, &rule),
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
				Config: testAccAWSCloudWatchEventRuleConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchEventRuleExists(resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSCloudWatchEventRuleConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchEventRuleExists(resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccAWSCloudWatchEventRule_IsEnabled(t *testing.T) {
	var rule events.DescribeRuleOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_cloudwatch_event_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudWatchEventRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudWatchEventRuleConfigIsEnabled(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchEventRuleExists(resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "is_enabled", "false"),
					testAccCheckCloudWatchEventRuleEnabled(resourceName, "DISABLED"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSCloudWatchEventRuleConfigIsEnabled(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchEventRuleExists(resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "is_enabled", "true"),
					testAccCheckCloudWatchEventRuleEnabled(resourceName, "ENABLED"),
				),
			},
			{
				Config: testAccAWSCloudWatchEventRuleConfigIsEnabled(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchEventRuleExists(resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "is_enabled", "false"),
					testAccCheckCloudWatchEventRuleEnabled(resourceName, "DISABLED"),
				),
			},
		},
	})
}

func testAccCheckCloudWatchEventRuleExists(n string, rule *events.DescribeRuleOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := testAccProvider.Meta().(*AWSClient).cloudwatcheventsconn
		params := events.DescribeRuleInput{
			Name: aws.String(rs.Primary.ID),
		}
		resp, err := conn.DescribeRule(&params)
		if err != nil {
			return err
		}
		if resp == nil {
			return fmt.Errorf("Rule not found")
		}

		*rule = *resp

		return nil
	}
}

func testAccCheckCloudWatchEventRuleEnabled(n string, desired string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := testAccProvider.Meta().(*AWSClient).cloudwatcheventsconn
		params := events.DescribeRuleInput{
			Name: aws.String(rs.Primary.ID),
		}
		resp, err := conn.DescribeRule(&params)

		if err != nil {
			return err
		}
		if *resp.State != desired {
			return fmt.Errorf("Expected state %q, given %q", desired, *resp.State)
		}

		return nil
	}
}

func testAccCheckAWSCloudWatchEventRuleDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).cloudwatcheventsconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_cloudwatch_event_rule" {
			continue
		}

		params := events.DescribeRuleInput{
			Name: aws.String(rs.Primary.ID),
		}

		resp, err := conn.DescribeRule(&params)

		if err == nil {
			return fmt.Errorf("CloudWatch Event Rule %q still exists: %s",
				rs.Primary.ID, resp)
		}
	}

	return nil
}

func testAccAWSCloudWatchEventRuleConfig(name string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_event_rule" "test" {
	name = "%s"
	schedule_expression = "rate(1 hour)"
}
`, name)
}

func testAccAWSCloudWatchEventRuleConfigPattern(name, pattern string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_event_rule" "test" {
	name = "%s"
	schedule_expression = "rate(1 hour)"
	event_pattern = <<PATTERN
	%s
	PATTERN
}
`, name, pattern)
}

func testAccAWSCloudWatchEventRuleConfigDescription(name, description string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_event_rule" "test" {
	name = %[1]q
	description = %[2]q
	schedule_expression = "rate(1 hour)"
}
`, name, description)
}

func testAccAWSCloudWatchEventRuleConfigIsEnabled(name string, enabled bool) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_event_rule" "test" {
	name = "%s"
	schedule_expression = "rate(1 hour)"
	is_enabled = %t
}
`, name, enabled)
}

func testAccAWSCloudWatchEventRuleConfigPrefix(name string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_event_rule" "test" {
	name_prefix = "%s"
	schedule_expression = "rate(5 minutes)"
}
`, name)
}

func testAccAWSCloudWatchEventRuleConfigTags1(name, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_event_rule" "test" {
	name = %[1]q
	schedule_expression = "rate(1 hour)"

	tags = {
	  %[2]q = %[3]q
	}
}
`, name, tagKey1, tagValue1)
}

func testAccAWSCloudWatchEventRuleConfigTags2(name, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_event_rule" "test" {
    name = %[1]q
    schedule_expression = "rate(1 hour)"

	tags = {
	  %[2]q = %[3]q
	  %[4]q = %[5]q
	}
}
`, name, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccAWSCloudWatchEventRuleConfigRole(name string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "events.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
POLICY
}

resource "aws_cloudwatch_event_rule" "test" {
	name = %[1]q
	schedule_expression = "rate(1 hour)"
	role_arn = "${aws_iam_role.test.arn}"
}
`, name)
}
