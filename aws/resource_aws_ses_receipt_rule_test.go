package aws

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSSESReceiptRule_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ses_receipt_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPreCheckAWSSES(t)
			testAccPreCheckSESReceiptRule(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSESReceiptRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSESReceiptRuleBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsSESReceiptRuleExists(resourceName),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "ses", fmt.Sprintf("receipt-rule-set/%s:receipt-rule/%s", rName, rName)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccAwsSesReceiptRuleImportStateIdFunc(resourceName),
			},
		},
	})
}

func TestAccAWSSESReceiptRule_s3Action(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ses_receipt_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPreCheckAWSSES(t)
			testAccPreCheckSESReceiptRule(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSESReceiptRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSESReceiptRuleS3ActionConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsSESReceiptRuleExists(resourceName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccAwsSesReceiptRuleImportStateIdFunc(resourceName),
			},
		},
	})
}

func TestAccAWSSESReceiptRule_order(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ses_receipt_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPreCheckAWSSES(t)
			testAccPreCheckSESReceiptRule(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSESReceiptRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSESReceiptRuleOrderConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsSESReceiptRuleOrder(resourceName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccAwsSesReceiptRuleImportStateIdFunc(resourceName),
			},
		},
	})
}

func TestAccAWSSESReceiptRule_actions(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ses_receipt_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPreCheckAWSSES(t)
			testAccPreCheckSESReceiptRule(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSESReceiptRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSESReceiptRuleActionsConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsSESReceiptRuleActions(resourceName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccAwsSesReceiptRuleImportStateIdFunc(resourceName),
			},
		},
	})
}

func TestAccAWSSESReceiptRule_disappears(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ses_receipt_rule.test"

	ruleSetResourceName := "aws_ses_receipt_rule_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPreCheckAWSSES(t)
			testAccPreCheckSESReceiptRule(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSESReceiptRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSESReceiptRuleBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsSESReceiptRuleExists(resourceName),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsSesReceiptRuleSet(), ruleSetResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccAWSSESReceiptRuleBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsSESReceiptRuleExists(resourceName),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsSesReceiptRule(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckSESReceiptRuleDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).sesconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ses_receipt_rule" {
			continue
		}

		params := &ses.DescribeReceiptRuleInput{
			RuleName:    aws.String(rs.Primary.Attributes["name"]),
			RuleSetName: aws.String(rs.Primary.Attributes["rule_set_name"]),
		}

		_, err := conn.DescribeReceiptRule(params)
		if err == nil {
			return fmt.Errorf("Receipt rule %s still exists. Failing!", rs.Primary.ID)
		}

		// Verify the error is what we want
		_, ok := err.(awserr.Error)
		if !ok {
			return err
		}

	}

	return nil

}

func testAccCheckAwsSESReceiptRuleExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("SES Receipt Rule not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("SES Receipt Rule name not set")
		}

		conn := testAccProvider.Meta().(*AWSClient).sesconn

		params := &ses.DescribeReceiptRuleInput{
			RuleName:    aws.String(rs.Primary.Attributes["name"]),
			RuleSetName: aws.String(rs.Primary.Attributes["rule_set_name"]),
		}

		response, err := conn.DescribeReceiptRule(params)
		if err != nil {
			return err
		}

		if !aws.BoolValue(response.Rule.Enabled) {
			return fmt.Errorf("Enabled (%v) was not set to true", *response.Rule.Enabled)
		}

		if !reflect.DeepEqual(response.Rule.Recipients, []*string{aws.String("test@example.com")}) {
			return fmt.Errorf("Recipients (%v) was not set to [test@example.com]", response.Rule.Recipients)
		}

		if !aws.BoolValue(response.Rule.ScanEnabled) {
			return fmt.Errorf("ScanEnabled (%v) was not set to true", *response.Rule.ScanEnabled)
		}

		if aws.StringValue(response.Rule.TlsPolicy) != ses.TlsPolicyRequire {
			return fmt.Errorf("TLS Policy (%s) was not set to Require", *response.Rule.TlsPolicy)
		}

		return nil
	}
}

func testAccAwsSesReceiptRuleImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not Found: %s", resourceName)
		}

		return fmt.Sprintf("%s:%s", rs.Primary.Attributes["rule_set_name"], rs.Primary.Attributes["name"]), nil
	}
}

func testAccCheckAwsSESReceiptRuleOrder(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("SES Receipt Rule not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("SES Receipt Rule name not set")
		}

		conn := testAccProvider.Meta().(*AWSClient).sesconn

		params := &ses.DescribeReceiptRuleSetInput{
			RuleSetName: aws.String(rs.Primary.Attributes["rule_set_name"]),
		}

		response, err := conn.DescribeReceiptRuleSet(params)
		if err != nil {
			return err
		}

		if len(response.Rules) != 2 {
			return fmt.Errorf("Number of rules (%d) was not equal to 2", len(response.Rules))
		} else if aws.StringValue(response.Rules[0].Name) != "first" ||
			aws.StringValue(response.Rules[1].Name) != "second" {
			return fmt.Errorf("Order of rules (%v) was incorrect", response.Rules)
		}

		return nil
	}
}

func testAccCheckAwsSESReceiptRuleActions(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("SES Receipt Rule not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("SES Receipt Rule name not set")
		}

		conn := testAccProvider.Meta().(*AWSClient).sesconn

		params := &ses.DescribeReceiptRuleInput{
			RuleName:    aws.String(rs.Primary.Attributes["name"]),
			RuleSetName: aws.String(rs.Primary.Attributes["rule_set_name"]),
		}

		response, err := conn.DescribeReceiptRule(params)
		if err != nil {
			return err
		}

		actions := response.Rule.Actions

		if len(actions) != 3 {
			return fmt.Errorf("Number of rules (%d) was not equal to 3", len(actions))
		}

		addHeaderAction := actions[0].AddHeaderAction
		if aws.StringValue(addHeaderAction.HeaderName) != "Another-Header" {
			return fmt.Errorf("Header Name (%s) was not equal to Another-Header", *addHeaderAction.HeaderName)
		}

		if aws.StringValue(addHeaderAction.HeaderValue) != "First" {
			return fmt.Errorf("Header Value (%s) was not equal to First", *addHeaderAction.HeaderValue)
		}

		secondAddHeaderAction := actions[1].AddHeaderAction
		if aws.StringValue(secondAddHeaderAction.HeaderName) != "Added-By" {
			return fmt.Errorf("Header Name (%s) was not equal to Added-By", *secondAddHeaderAction.HeaderName)
		}

		if aws.StringValue(secondAddHeaderAction.HeaderValue) != "Terraform" {
			return fmt.Errorf("Header Value (%s) was not equal to Terraform", *secondAddHeaderAction.HeaderValue)
		}

		stopAction := actions[2].StopAction
		if aws.StringValue(stopAction.Scope) != ses.StopScopeRuleSet {
			return fmt.Errorf("Scope (%s) was not equal to RuleSet", *stopAction.Scope)
		}

		return nil
	}
}

func testAccPreCheckSESReceiptRule(t *testing.T) {
	conn := testAccProvider.Meta().(*AWSClient).sesconn

	input := &ses.DescribeReceiptRuleInput{
		RuleName:    aws.String("MyRule"),
		RuleSetName: aws.String("MyRuleSet"),
	}

	_, err := conn.DescribeReceiptRule(input)

	if testAccPreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if isAWSErr(err, "RuleSetDoesNotExist", "") {
		return
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccAWSSESReceiptRuleBasicConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_ses_receipt_rule_set" "test" {
  rule_set_name = %[1]q
}

resource "aws_ses_receipt_rule" "test" {
  name          = %[1]q
  rule_set_name = aws_ses_receipt_rule_set.test.rule_set_name
  recipients    = ["test@example.com"]
  enabled       = true
  scan_enabled  = true
  tls_policy    = "Require"
}
`, rName)
}

func testAccAWSSESReceiptRuleS3ActionConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_ses_receipt_rule_set" "test" {
  rule_set_name = %[1]q
}

resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  acl           = "public-read-write"
  force_destroy = "true"
}

resource "aws_ses_receipt_rule" "test" {
  name          = %[1]q
  rule_set_name = aws_ses_receipt_rule_set.test.rule_set_name
  recipients    = ["test@example.com"]
  enabled       = true
  scan_enabled  = true
  tls_policy    = "Require"

  s3_action {
    bucket_name = aws_s3_bucket.test.id
    position    = 1
  }
}
`, rName)
}

func testAccAWSSESReceiptRuleOrderConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_ses_receipt_rule_set" "test" {
  rule_set_name = %[1]q
}

resource "aws_ses_receipt_rule" "test" {
  name          = "second"
  rule_set_name = aws_ses_receipt_rule_set.test.rule_set_name
  after         = aws_ses_receipt_rule.test1.name
}

resource "aws_ses_receipt_rule" "test1" {
  name          = "first"
  rule_set_name = aws_ses_receipt_rule_set.test.rule_set_name
}
`, rName)
}

func testAccAWSSESReceiptRuleActionsConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_ses_receipt_rule_set" "test" {
  rule_set_name = %[1]q
}

resource "aws_ses_receipt_rule" "test" {
  name          = %[1]q
  rule_set_name = aws_ses_receipt_rule_set.test.rule_set_name

  add_header_action {
    header_name  = "Added-By"
    header_value = "Terraform"
    position     = 2
  }

  add_header_action {
    header_name  = "Another-Header"
    header_value = "First"
    position     = 1
  }

  stop_action {
    scope    = "RuleSet"
    position = 3
  }
}
`, rName)
}
