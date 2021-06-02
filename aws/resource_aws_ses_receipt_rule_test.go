package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSSESReceiptRule_basic(t *testing.T) {
	var rule ses.ReceiptRule

	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ses_receipt_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPreCheckAWSSES(t)
			testAccPreCheckSESReceiptRule(t)
		},
		ErrorCheck:   testAccErrorCheck(t, ses.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSESReceiptRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSESReceiptRuleBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsSESReceiptRuleExists(resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "rule_set_name", rName),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "ses", fmt.Sprintf("receipt-rule-set/%s:receipt-rule/%s", rName, rName)),
					resource.TestCheckResourceAttr(resourceName, "add_header_action.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "bounce_action.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "lambda_action.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "s3_action.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "sns_action.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "stop_action.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "workmail_action.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "recipients.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "recipients.*", "test@example.com"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "scan_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "tls_policy", "Require"),
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
	var rule ses.ReceiptRule

	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ses_receipt_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPreCheckAWSSES(t)
			testAccPreCheckSESReceiptRule(t)
		},
		ErrorCheck:   testAccErrorCheck(t, ses.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSESReceiptRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSESReceiptRuleS3ActionConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsSESReceiptRuleExists(resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "s3_action.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "s3_action.*.bucket_name", "aws_s3_bucket.test", "id"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "s3_action.*", map[string]string{
						"position": "1",
					}),
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

func TestAccAWSSESReceiptRule_snsAction(t *testing.T) {
	var rule ses.ReceiptRule

	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ses_receipt_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPreCheckAWSSES(t)
			testAccPreCheckSESReceiptRule(t)
		},
		ErrorCheck:   testAccErrorCheck(t, ses.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSESReceiptRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSESReceiptRuleSNSActionConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsSESReceiptRuleExists(resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "sns_action.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "sns_action.*.topic_arn", "aws_sns_topic.test", "arn"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "sns_action.*", map[string]string{
						"encoding": "UTF-8",
						"position": "1",
					}),
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

func TestAccAWSSESReceiptRule_snsActionEncoding(t *testing.T) {
	var rule ses.ReceiptRule

	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ses_receipt_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPreCheckAWSSES(t)
			testAccPreCheckSESReceiptRule(t)
		},
		ErrorCheck:   testAccErrorCheck(t, ses.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSESReceiptRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSESReceiptRuleSNSActionEncodingConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsSESReceiptRuleExists(resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "sns_action.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "sns_action.*.topic_arn", "aws_sns_topic.test", "arn"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "sns_action.*", map[string]string{
						"encoding": "Base64",
						"position": "1",
					}),
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

func TestAccAWSSESReceiptRule_lambdaAction(t *testing.T) {
	var rule ses.ReceiptRule

	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ses_receipt_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPreCheckAWSSES(t)
			testAccPreCheckSESReceiptRule(t)
		},
		ErrorCheck:   testAccErrorCheck(t, ses.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSESReceiptRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSESReceiptRuleLambdaActionConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsSESReceiptRuleExists(resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "lambda_action.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "lambda_action.*.function_arn", "aws_lambda_function.test", "arn"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "lambda_action.*", map[string]string{
						"invocation_type": "Event",
						"position":        "1",
					}),
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

func TestAccAWSSESReceiptRule_stopAction(t *testing.T) {
	var rule ses.ReceiptRule

	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ses_receipt_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPreCheckAWSSES(t)
			testAccPreCheckSESReceiptRule(t)
		},
		ErrorCheck:   testAccErrorCheck(t, ses.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSESReceiptRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSESReceiptRuleStopActionConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsSESReceiptRuleExists(resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "stop_action.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "stop_action.*", map[string]string{
						"scope": "RuleSet",
					}),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "stop_action.*.topic_arn", "aws_sns_topic.test", "arn"),
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
	var rule ses.ReceiptRule

	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ses_receipt_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPreCheckAWSSES(t)
			testAccPreCheckSESReceiptRule(t)
		},
		ErrorCheck:   testAccErrorCheck(t, ses.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSESReceiptRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSESReceiptRuleOrderConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsSESReceiptRuleExists(resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "name", "second"),
					resource.TestCheckResourceAttrPair(resourceName, "after", "aws_ses_receipt_rule.test1", "name"),
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
	var rule ses.ReceiptRule

	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ses_receipt_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPreCheckAWSSES(t)
			testAccPreCheckSESReceiptRule(t)
		},
		ErrorCheck:   testAccErrorCheck(t, ses.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSESReceiptRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSESReceiptRuleActionsConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsSESReceiptRuleExists(resourceName, &rule),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "add_header_action.*", map[string]string{
						"header_name":  "Added-By",
						"header_value": "Terraform",
						"position":     "2",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "add_header_action.*", map[string]string{
						"header_name":  "Another-Header",
						"header_value": "First",
						"position":     "1",
					}),
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
	var rule ses.ReceiptRule

	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ses_receipt_rule.test"

	ruleSetResourceName := "aws_ses_receipt_rule_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPreCheckAWSSES(t)
			testAccPreCheckSESReceiptRule(t)
		},
		ErrorCheck:   testAccErrorCheck(t, ses.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSESReceiptRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSESReceiptRuleBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsSESReceiptRuleExists(resourceName, &rule),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsSesReceiptRuleSet(), ruleSetResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccAWSSESReceiptRuleBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsSESReceiptRuleExists(resourceName, &rule),
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

func testAccCheckAwsSESReceiptRuleExists(n string, rule *ses.ReceiptRule) resource.TestCheckFunc {
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

		resp, err := conn.DescribeReceiptRule(params)
		if err != nil {
			return err
		}

		*rule = *resp.Rule

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

func testAccAWSSESReceiptRuleSNSActionConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_ses_receipt_rule_set" "test" {
  rule_set_name = %[1]q
}

resource "aws_sns_topic" "test" {
  name = %[1]q
}

resource "aws_ses_receipt_rule" "test" {
  name          = %[1]q
  rule_set_name = aws_ses_receipt_rule_set.test.rule_set_name
  recipients    = ["test@example.com"]
  enabled       = true
  scan_enabled  = true
  tls_policy    = "Require"

  sns_action {
    topic_arn = aws_sns_topic.test.arn
    position  = 1
  }
}
`, rName)
}

func testAccAWSSESReceiptRuleSNSActionEncodingConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_ses_receipt_rule_set" "test" {
  rule_set_name = %[1]q
}

resource "aws_sns_topic" "test" {
  name = %[1]q
}

resource "aws_ses_receipt_rule" "test" {
  name          = %[1]q
  rule_set_name = aws_ses_receipt_rule_set.test.rule_set_name
  recipients    = ["test@example.com"]
  enabled       = true
  scan_enabled  = true
  tls_policy    = "Require"

  sns_action {
    topic_arn = aws_sns_topic.test.arn
    encoding  = "Base64"
    position  = 1
  }
}
`, rName)
}

func testAccAWSSESReceiptRuleLambdaActionConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_ses_receipt_rule_set" "test" {
  rule_set_name = %[1]q
}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = jsonencode({
    "Version" : "2012-10-17",
    "Statement" : [
      {
        "Action" : "sts:AssumeRole",
        "Principal" : {
          "Service" : "lambda.amazonaws.com"
        },
        "Effect" : "Allow",
        "Sid" : ""
      }
    ]
  })
}

resource "aws_lambda_function" "test" {
  filename         = "test-fixtures/lambdatest.zip"
  source_code_hash = filebase64sha256("test-fixtures/lambdatest.zip")
  function_name    = %[1]q
  role             = aws_iam_role.test.arn
  handler          = "exports.example"
  runtime          = "nodejs12.x"
}

resource "aws_lambda_permission" "test" {
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.test.arn
  principal     = "ses.amazonaws.com"
}

resource "aws_ses_receipt_rule" "test" {
  name          = %[1]q
  rule_set_name = aws_ses_receipt_rule_set.test.rule_set_name
  recipients    = ["test@example.com"]
  enabled       = true
  scan_enabled  = true
  tls_policy    = "Require"

  lambda_action {
    function_arn = aws_lambda_function.test.arn
    position     = 1
  }

  depends_on = [aws_lambda_permission.test]
}
`, rName)
}

func testAccAWSSESReceiptRuleStopActionConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_ses_receipt_rule_set" "test" {
  rule_set_name = %[1]q
}

resource "aws_sns_topic" "test" {
  name = %[1]q
}

resource "aws_ses_receipt_rule" "test" {
  name          = %[1]q
  rule_set_name = aws_ses_receipt_rule_set.test.rule_set_name
  recipients    = ["test@example.com"]
  enabled       = true
  scan_enabled  = true
  tls_policy    = "Require"

  stop_action {
    topic_arn = aws_sns_topic.test.arn
    scope     = "RuleSet"
    position  = 1
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
