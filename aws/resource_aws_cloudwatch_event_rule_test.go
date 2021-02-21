package aws

import (
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	events "github.com/aws/aws-sdk-go/service/cloudwatchevents"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/naming"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/cloudwatchevents/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/cloudwatchevents/lister"
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
		return fmt.Errorf("Error getting client: %w", err)
	}
	conn := client.(*AWSClient).cloudwatcheventsconn

	var sweeperErrs *multierror.Error
	var count int

	rulesInput := &events.ListRulesInput{}

	err = lister.ListRulesPages(conn, rulesInput, func(rulesPage *events.ListRulesOutput, lastRulesPage bool) bool {
		if rulesPage == nil {
			return !lastRulesPage
		}

		for _, rule := range rulesPage.Rules {
			count++
			name := aws.StringValue(rule.Name)

			log.Printf("[INFO] Deleting CloudWatch Events rule (%s)", name)
			_, err := conn.DeleteRule(&events.DeleteRuleInput{
				Name:  aws.String(name),
				Force: aws.Bool(true), // Required for AWS-managed rules, ignored otherwise
			})
			if err != nil {
				sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error deleting CloudWatch Events rule (%s): %w", name, err))
				continue
			}
		}

		return !lastRulesPage
	})

	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping CloudWatch Events rule sweeper for %q: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing CloudWatch Events rules: %w", err))
	}

	log.Printf("[INFO] Deleted %d CloudWatch Events rules", count)

	return sweeperErrs.ErrorOrNil()
}

func TestAccAWSCloudWatchEventRule_basic(t *testing.T) {
	var v1, v2, v3 events.DescribeRuleOutput
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
					testAccCheckCloudWatchEventRuleExists(resourceName, &v1),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "events", regexp.MustCompile(fmt.Sprintf(`rule/%s$`, rName))),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "schedule_expression", "rate(1 hour)"),
					resource.TestCheckResourceAttr(resourceName, "event_bus_name", "default"),
					resource.TestCheckNoResourceAttr(resourceName, "event_pattern"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "role_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "is_enabled", "true"),
					testAccCheckCloudWatchEventRuleEnabled(resourceName, "ENABLED"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccAWSCloudWatchEventRuleNoBusNameImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSCloudWatchEventRuleConfig(rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchEventRuleExists(resourceName, &v2),
					testAccCheckCloudWatchEventRuleRecreated(&v1, &v2),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "events", regexp.MustCompile(fmt.Sprintf(`rule/%s$`, rName2))),
					resource.TestCheckResourceAttr(resourceName, "name", rName2),
					resource.TestCheckResourceAttr(resourceName, "event_bus_name", "default"),
					resource.TestCheckResourceAttr(resourceName, "schedule_expression", "rate(1 hour)"),
					resource.TestCheckResourceAttr(resourceName, "role_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "is_enabled", "true"),
					testAccCheckCloudWatchEventRuleEnabled(resourceName, "ENABLED"),
				),
			},
			{
				Config: testAccAWSCloudWatchEventRuleConfigDefaultEventBusName(rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchEventRuleExists(resourceName, &v3),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "events", regexp.MustCompile(fmt.Sprintf(`rule/%s$`, rName2))),
					testAccCheckCloudWatchEventRuleNotRecreated(&v2, &v3),
					resource.TestCheckResourceAttr(resourceName, "name", rName2),
					resource.TestCheckResourceAttr(resourceName, "event_bus_name", "default"),
				),
			},
		},
	})
}

func TestAccAWSCloudWatchEventRule_EventBusName(t *testing.T) {
	var v1, v2, v3 events.DescribeRuleOutput
	rName := acctest.RandomWithPrefix("tf-acc-test-rule")
	rName2 := acctest.RandomWithPrefix("tf-acc-test-rule")
	busName := acctest.RandomWithPrefix("tf-acc-test-bus")
	busName2 := acctest.RandomWithPrefix("tf-acc-test-bus")
	resourceName := "aws_cloudwatch_event_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudWatchEventRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudWatchEventRuleConfigEventBusName(rName, busName, "description 1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchEventRuleExists(resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "event_bus_name", busName),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "events", regexp.MustCompile(fmt.Sprintf(`rule/%s/%s$`, busName, rName))),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSCloudWatchEventRuleConfigEventBusName(rName, busName, "description 2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchEventRuleExists(resourceName, &v2),
					testAccCheckCloudWatchEventRuleNotRecreated(&v1, &v2),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "event_bus_name", busName),
				),
			},
			{
				Config: testAccAWSCloudWatchEventRuleConfigEventBusName(rName2, busName2, "description 2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchEventRuleExists(resourceName, &v3),
					testAccCheckCloudWatchEventRuleRecreated(&v2, &v3),
					resource.TestCheckResourceAttr(resourceName, "name", rName2),
					resource.TestCheckResourceAttr(resourceName, "event_bus_name", busName2),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "events", regexp.MustCompile(fmt.Sprintf(`rule/%s/%s$`, busName2, rName2))),
				),
			},
		},
	})
}

func TestAccAWSCloudWatchEventRule_role(t *testing.T) {
	var v events.DescribeRuleOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resourceName := "aws_cloudwatch_event_rule.test"
	iamRoleResourceName := "aws_iam_role.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudWatchEventRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudWatchEventRuleConfigRole(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchEventRuleExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrPair(resourceName, "role_arn", iamRoleResourceName, "arn"),
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
	var v1, v2 events.DescribeRuleOutput
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
					testAccCheckCloudWatchEventRuleExists(resourceName, &v1),
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
					testAccCheckCloudWatchEventRuleExists(resourceName, &v2),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "description", "description2"),
				),
			},
		},
	})
}

func TestAccAWSCloudWatchEventRule_pattern(t *testing.T) {
	var v1, v2 events.DescribeRuleOutput
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
					testAccCheckCloudWatchEventRuleExists(resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "schedule_expression", ""),
					testAccCheckResourceAttrEquivalentJSON(resourceName, "event_pattern", "{\"source\":[\"aws.ec2\"]}"),
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
					testAccCheckCloudWatchEventRuleExists(resourceName, &v2),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					testAccCheckResourceAttrEquivalentJSON(resourceName, "event_pattern", "{\"source\":[\"aws.lambda\"]}"),
				),
			},
		},
	})
}

func TestAccAWSCloudWatchEventRule_ScheduleAndPattern(t *testing.T) {
	var v events.DescribeRuleOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_cloudwatch_event_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudWatchEventRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudWatchEventRuleConfigScheduleAndPattern(rName, "{\"source\":[\"aws.ec2\"]}"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchEventRuleExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "schedule_expression", "rate(1 hour)"),
					testAccCheckResourceAttrEquivalentJSON(resourceName, "event_pattern", "{\"source\":[\"aws.ec2\"]}"),
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

func TestAccAWSCloudWatchEventRule_NamePrefix(t *testing.T) {
	var v events.DescribeRuleOutput
	rName := "tf-acc-test-prefix-"
	resourceName := "aws_cloudwatch_event_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudWatchEventRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudWatchEventRuleConfigNamePrefix(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchEventRuleExists(resourceName, &v),
					naming.TestCheckResourceAttrNameFromPrefix(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", rName),
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

func TestAccAWSCloudWatchEventRule_Name_Generated(t *testing.T) {
	var v events.DescribeRuleOutput
	resourceName := "aws_cloudwatch_event_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudWatchEventRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudWatchEventRuleConfigNameGenerated,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchEventRuleExists(resourceName, &v),
					naming.TestCheckResourceAttrNameGenerated(resourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", "terraform-"),
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

func TestAccAWSCloudWatchEventRule_tags(t *testing.T) {
	var v1, v2, v3 events.DescribeRuleOutput
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
					testAccCheckCloudWatchEventRuleExists(resourceName, &v1),
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
					testAccCheckCloudWatchEventRuleExists(resourceName, &v2),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSCloudWatchEventRuleConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchEventRuleExists(resourceName, &v3),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSCloudWatchEventRuleConfigTags0(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchEventRuleExists(resourceName, &v3),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
		},
	})
}

func TestAccAWSCloudWatchEventRule_IsEnabled(t *testing.T) {
	var v1, v2, v3 events.DescribeRuleOutput
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
					testAccCheckCloudWatchEventRuleExists(resourceName, &v1),
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
					testAccCheckCloudWatchEventRuleExists(resourceName, &v2),
					resource.TestCheckResourceAttr(resourceName, "is_enabled", "true"),
					testAccCheckCloudWatchEventRuleEnabled(resourceName, "ENABLED"),
				),
			},
			{
				Config: testAccAWSCloudWatchEventRuleConfigIsEnabled(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchEventRuleExists(resourceName, &v3),
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

		resp, err := finder.RuleByID(conn, rs.Primary.ID)
		if err != nil {
			return err
		}
		if resp == nil {
			return fmt.Errorf("CloudWatch Events rule (%s) not found", rs.Primary.ID)
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

		resp, err := finder.RuleByID(conn, rs.Primary.ID)
		if err != nil {
			return err
		}
		if aws.StringValue(resp.State) != desired {
			return fmt.Errorf("Expected state %q, given %q", desired, aws.StringValue(resp.State))
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

		resp, err := finder.RuleByID(conn, rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("CloudWatch Events Rule (%s) still exists: %s", rs.Primary.ID, resp)
		}
	}

	return nil
}

func testAccCheckCloudWatchEventRuleRecreated(i, j *events.DescribeRuleOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(i.Arn) == aws.StringValue(j.Arn) {
			return fmt.Errorf("CloudWatch Events rule not recreated, but expected it to be")
		}
		return nil
	}
}

func testAccCheckCloudWatchEventRuleNotRecreated(i, j *events.DescribeRuleOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(i.Arn) != aws.StringValue(j.Arn) {
			return fmt.Errorf("CloudWatch Events rule recreated, but expected it to not be")
		}
		return nil
	}
}

func testAccAWSCloudWatchEventRuleNoBusNameImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return rs.Primary.Attributes["name"], nil
	}
}

func testAccAWSCloudWatchEventRuleConfig(name string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_event_rule" "test" {
  name                = "%s"
  schedule_expression = "rate(1 hour)"
}
`, name)
}

func testAccAWSCloudWatchEventRuleConfigDefaultEventBusName(name string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_event_rule" "test" {
  name                = %[1]q
  schedule_expression = "rate(1 hour)"
  event_bus_name      = "default"
}
`, name)
}

func testAccAWSCloudWatchEventRuleConfigEventBusName(name, eventBusName, description string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_event_rule" "test" {
  name           = %[1]q
  event_bus_name = aws_cloudwatch_event_bus.test.name
  description    = %[2]q
  event_pattern  = <<PATTERN
{
	"source": [
		"aws.ec2"
	]
}
PATTERN
}

resource "aws_cloudwatch_event_bus" "test" {
  name = %[3]q
}
`, name, description, eventBusName)
}

func testAccAWSCloudWatchEventRuleConfigPattern(name, pattern string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_event_rule" "test" {
  name          = "%s"
  event_pattern = <<PATTERN
	%s
PATTERN
}
`, name, pattern)
}

func testAccAWSCloudWatchEventRuleConfigScheduleAndPattern(name, pattern string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_event_rule" "test" {
  name                = "%s"
  schedule_expression = "rate(1 hour)"
  event_pattern       = <<PATTERN
	%s
PATTERN
}
`, name, pattern)
}

func testAccAWSCloudWatchEventRuleConfigDescription(name, description string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_event_rule" "test" {
  name                = %[1]q
  description         = %[2]q
  schedule_expression = "rate(1 hour)"
}
`, name, description)
}

func testAccAWSCloudWatchEventRuleConfigIsEnabled(name string, enabled bool) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_event_rule" "test" {
  name                = "%s"
  schedule_expression = "rate(1 hour)"
  is_enabled          = %t
}
`, name, enabled)
}

func testAccAWSCloudWatchEventRuleConfigNamePrefix(name string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_event_rule" "test" {
  name_prefix         = "%s"
  schedule_expression = "rate(5 minutes)"
}
`, name)
}

const testAccAWSCloudWatchEventRuleConfigNameGenerated = `
resource "aws_cloudwatch_event_rule" "test" {
  schedule_expression = "rate(5 minutes)"
}
`

func testAccAWSCloudWatchEventRuleConfigTags1(name, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_event_rule" "test" {
  name                = %[1]q
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
  name                = %[1]q
  schedule_expression = "rate(1 hour)"

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, name, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccAWSCloudWatchEventRuleConfigTags0(name string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_event_rule" "test" {
  name                = %[1]q
  schedule_expression = "rate(1 hour)"
}
`, name)
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
  name                = %[1]q
  schedule_expression = "rate(1 hour)"
  role_arn            = aws_iam_role.test.arn
}
`, name)
}
