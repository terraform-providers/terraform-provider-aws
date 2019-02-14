package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sns"
	multierror "github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/jen20/awspolicyequivalence"
)

func TestAccAWSSNSTopic_importBasic(t *testing.T) {
	resourceName := "aws_sns_topic.test_topic"
	rName := acctest.RandString(10)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSNSTopicDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSNSTopicConfig_withName(rName),
			},

			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSSNSTopic_basic(t *testing.T) {
	attributes := make(map[string]string)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_sns_topic.test_topic",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSSNSTopicDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSNSTopicConfig_withGeneratedName,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSNSTopicExists("aws_sns_topic.test_topic", attributes),
				),
			},
		},
	})
}

func TestAccAWSSNSTopic_name(t *testing.T) {
	attributes := make(map[string]string)

	rName := acctest.RandString(10)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_sns_topic.test_topic",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSSNSTopicDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSNSTopicConfig_withName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSNSTopicExists("aws_sns_topic.test_topic", attributes),
				),
			},
		},
	})
}

func TestAccAWSSNSTopic_namePrefix(t *testing.T) {
	attributes := make(map[string]string)

	startsWithPrefix := regexp.MustCompile("^terraform-test-topic-")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_sns_topic.test_topic",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSSNSTopicDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSNSTopicConfig_withNamePrefix(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSNSTopicExists("aws_sns_topic.test_topic", attributes),
					resource.TestMatchResourceAttr("aws_sns_topic.test_topic", "name", startsWithPrefix),
				),
			},
		},
	})
}

func TestAccAWSSNSTopic_policy(t *testing.T) {
	attributes := make(map[string]string)

	rName := acctest.RandString(10)
	expectedPolicy := `{"Statement":[{"Sid":"Stmt1445931846145","Effect":"Allow","Principal":{"AWS":"*"},"Action":"sns:Publish","Resource":"arn:aws:sns:us-west-2::example"}],"Version":"2012-10-17","Id":"Policy1445931846145"}`
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_sns_topic.test_topic",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSSNSTopicDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSNSTopicWithPolicy(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSNSTopicExists("aws_sns_topic.test_topic", attributes),
					testAccCheckAWSNSTopicHasPolicy("aws_sns_topic.test_topic", expectedPolicy),
				),
			},
		},
	})
}

func TestAccAWSSNSTopic_withIAMRole(t *testing.T) {
	attributes := make(map[string]string)

	rName := acctest.RandString(10)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_sns_topic.test_topic",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSSNSTopicDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSNSTopicConfig_withIAMRole(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSNSTopicExists("aws_sns_topic.test_topic", attributes),
				),
			},
		},
	})
}

func TestAccAWSSNSTopic_withFakeIAMRole(t *testing.T) {
	rName := acctest.RandString(10)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_sns_topic.test_topic",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSSNSTopicDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccAWSSNSTopicConfig_withFakeIAMRole(rName),
				ExpectError: regexp.MustCompile(`PrincipalNotFound`),
			},
		},
	})
}

func TestAccAWSSNSTopic_withDeliveryPolicy(t *testing.T) {
	attributes := make(map[string]string)

	rName := acctest.RandString(10)
	expectedPolicy := `{"http":{"defaultHealthyRetryPolicy": {"minDelayTarget": 20,"maxDelayTarget": 20,"numMaxDelayRetries": 0,"numRetries": 3,"numNoDelayRetries": 0,"numMinDelayRetries": 0,"backoffFunction": "linear"},"disableSubscriptionOverrides": false}}`
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_sns_topic.test_topic",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSSNSTopicDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSNSTopicConfig_withDeliveryPolicy(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSNSTopicExists("aws_sns_topic.test_topic", attributes),
					testAccCheckAWSNSTopicHasDeliveryPolicy("aws_sns_topic.test_topic", expectedPolicy),
				),
			},
		},
	})
}

func TestAccAWSSNSTopic_deliveryStatus(t *testing.T) {
	attributes := make(map[string]string)

	rName := acctest.RandString(10)
	arnRegex := regexp.MustCompile("^arn:aws:iam::[0-9]{12}:role/sns-delivery-status-role-")
	expectedAttributes := map[string]*regexp.Regexp{
		"ApplicationFailureFeedbackRoleArn":    arnRegex,
		"ApplicationSuccessFeedbackRoleArn":    arnRegex,
		"ApplicationSuccessFeedbackSampleRate": regexp.MustCompile(`^100$`),
		"HTTPFailureFeedbackRoleArn":           arnRegex,
		"HTTPSuccessFeedbackRoleArn":           arnRegex,
		"HTTPSuccessFeedbackSampleRate":        regexp.MustCompile(`^80$`),
		"LambdaFailureFeedbackRoleArn":         arnRegex,
		"LambdaSuccessFeedbackRoleArn":         arnRegex,
		"LambdaSuccessFeedbackSampleRate":      regexp.MustCompile(`^90$`),
		"SQSFailureFeedbackRoleArn":            arnRegex,
		"SQSSuccessFeedbackRoleArn":            arnRegex,
		"SQSSuccessFeedbackSampleRate":         regexp.MustCompile(`^70$`),
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_sns_topic.test_topic",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSSNSTopicDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSNSTopicConfig_deliveryStatus(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSNSTopicExists("aws_sns_topic.test_topic", attributes),
					testAccCheckAWSSNSTopicAttributes(attributes, expectedAttributes),
					resource.TestMatchResourceAttr("aws_sns_topic.test_topic", "application_success_feedback_role_arn", arnRegex),
					resource.TestCheckResourceAttr("aws_sns_topic.test_topic", "application_success_feedback_sample_rate", "100"),
					resource.TestMatchResourceAttr("aws_sns_topic.test_topic", "application_failure_feedback_role_arn", arnRegex),
					resource.TestMatchResourceAttr("aws_sns_topic.test_topic", "lambda_success_feedback_role_arn", arnRegex),
					resource.TestCheckResourceAttr("aws_sns_topic.test_topic", "lambda_success_feedback_sample_rate", "90"),
					resource.TestMatchResourceAttr("aws_sns_topic.test_topic", "lambda_failure_feedback_role_arn", arnRegex),
					resource.TestMatchResourceAttr("aws_sns_topic.test_topic", "http_success_feedback_role_arn", arnRegex),
					resource.TestCheckResourceAttr("aws_sns_topic.test_topic", "http_success_feedback_sample_rate", "80"),
					resource.TestMatchResourceAttr("aws_sns_topic.test_topic", "http_failure_feedback_role_arn", arnRegex),
					resource.TestMatchResourceAttr("aws_sns_topic.test_topic", "sqs_success_feedback_role_arn", arnRegex),
					resource.TestCheckResourceAttr("aws_sns_topic.test_topic", "sqs_success_feedback_sample_rate", "70"),
					resource.TestMatchResourceAttr("aws_sns_topic.test_topic", "sqs_failure_feedback_role_arn", arnRegex),
				),
			},
		},
	})
}

func TestAccAWSSNSTopic_encryption(t *testing.T) {
	attributes := make(map[string]string)

	rName := acctest.RandString(10)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_sns_topic.test_topic",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSSNSTopicDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSNSTopicConfig_withEncryption(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSNSTopicExists("aws_sns_topic.test_topic", attributes),
					resource.TestCheckResourceAttr("aws_sns_topic.test_topic", "kms_master_key_id", "alias/aws/sns"),
				),
			},
			{
				Config: testAccAWSSNSTopicConfig_withName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSNSTopicExists("aws_sns_topic.test_topic", attributes),
					resource.TestCheckResourceAttr("aws_sns_topic.test_topic", "kms_master_key_id", ""),
				),
			},
		},
	})
}

func testAccCheckAWSNSTopicHasPolicy(n string, expectedPolicyText string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Queue URL specified!")
		}

		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No SNS topic with that ARN exists")
		}

		conn := testAccProvider.Meta().(*AWSClient).snsconn

		params := &sns.GetTopicAttributesInput{
			TopicArn: aws.String(rs.Primary.ID),
		}
		resp, err := conn.GetTopicAttributes(params)
		if err != nil {
			return err
		}

		var actualPolicyText string
		for k, v := range resp.Attributes {
			if k == "Policy" {
				actualPolicyText = *v
				break
			}
		}

		equivalent, err := awspolicy.PoliciesAreEquivalent(actualPolicyText, expectedPolicyText)
		if err != nil {
			return fmt.Errorf("Error testing policy equivalence: %s", err)
		}
		if !equivalent {
			return fmt.Errorf("Non-equivalent policy error:\n\nexpected: %s\n\n     got: %s\n",
				expectedPolicyText, actualPolicyText)
		}

		return nil
	}
}

func testAccCheckAWSNSTopicHasDeliveryPolicy(n string, expectedPolicyText string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Queue URL specified!")
		}

		conn := testAccProvider.Meta().(*AWSClient).snsconn

		params := &sns.GetTopicAttributesInput{
			TopicArn: aws.String(rs.Primary.ID),
		}
		resp, err := conn.GetTopicAttributes(params)
		if err != nil {
			return err
		}

		var actualPolicyText string
		for k, v := range resp.Attributes {
			if k == "DeliveryPolicy" {
				actualPolicyText = *v
				break
			}
		}

		equivalent := suppressEquivalentJsonDiffs("", actualPolicyText, expectedPolicyText, nil)

		if !equivalent {
			return fmt.Errorf("Non-equivalent delivery policy error:\n\nexpected: %s\n\n     got: %s\n",
				expectedPolicyText, actualPolicyText)
		}

		return nil
	}
}

func testAccCheckAWSSNSTopicDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).snsconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_sns_topic" {
			continue
		}

		// Check if the topic exists by fetching its attributes
		params := &sns.GetTopicAttributesInput{
			TopicArn: aws.String(rs.Primary.ID),
		}
		_, err := conn.GetTopicAttributes(params)
		if err != nil {
			if isAWSErr(err, sns.ErrCodeNotFoundException, "") {
				return nil
			}
			return err
		}
		return fmt.Errorf("Topic exists when it should be destroyed!")
	}

	return nil
}

func testAccCheckAWSSNSTopicAttributes(attributes map[string]string, expectedAttributes map[string]*regexp.Regexp) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		var errors error
		for k, expectedR := range expectedAttributes {
			if v, ok := attributes[k]; !ok || !expectedR.MatchString(v) {
				err := fmt.Errorf("expected SNS topic attribute %q to match %q, received: %q", k, expectedR.String(), v)
				errors = multierror.Append(errors, err)
			}
		}
		return errors
	}
}

func testAccCheckAWSSNSTopicExists(n string, attributes map[string]string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No SNS topic with that ARN exists")
		}

		conn := testAccProvider.Meta().(*AWSClient).snsconn

		params := &sns.GetTopicAttributesInput{
			TopicArn: aws.String(rs.Primary.ID),
		}
		out, err := conn.GetTopicAttributes(params)

		if err != nil {
			return err
		}

		for k, v := range out.Attributes {
			attributes[k] = *v
		}

		return nil
	}
}

const testAccAWSSNSTopicConfig_withGeneratedName = `
resource "aws_sns_topic" "test_topic" {}
`

func testAccAWSSNSTopicConfig_withName(r string) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "test_topic" {
    name = "terraform-test-topic-%s"
}
`, r)
}

func testAccAWSSNSTopicConfig_withNamePrefix() string {
	return `
resource "aws_sns_topic" "test_topic" {
    name_prefix = "terraform-test-topic-"
}
`
}

func testAccAWSSNSTopicWithPolicy(r string) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "test_topic" {
  name = "example-%s"
  policy = <<EOF
{
  "Statement": [
    {
      "Sid": "Stmt1445931846145",
      "Effect": "Allow",
      "Principal": {
        "AWS": "*"
       },
      "Action": "sns:Publish",
      "Resource": "arn:aws:sns:us-west-2::example"
    }
  ],
  "Version": "2012-10-17",
  "Id": "Policy1445931846145"
}
EOF
}
`, r)
}

// Test for https://github.com/hashicorp/terraform/issues/3660
func testAccAWSSNSTopicConfig_withIAMRole(r string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "example" {
  name = "tf_acc_test_%s"
  path = "/test/"
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "ec2.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_sns_topic" "test_topic" {
  name = "tf-acc-test-with-iam-role-%s"
  policy = <<EOF
{
  "Statement": [
    {
      "Sid": "Stmt1445931846145",
      "Effect": "Allow",
      "Principal": {
        "AWS": "${aws_iam_role.example.arn}"
			},
      "Action": "sns:Publish",
      "Resource": "arn:aws:sns:us-west-2::example"
    }
  ],
  "Version": "2012-10-17",
  "Id": "Policy1445931846145"
}
EOF
}
`, r, r)
}

// Test for https://github.com/hashicorp/terraform/issues/14024
func testAccAWSSNSTopicConfig_withDeliveryPolicy(r string) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "test_topic" {
  name = "tf_acc_test_delivery_policy_%s"
  delivery_policy = <<EOF
{
  "http": {
    "defaultHealthyRetryPolicy": {
      "minDelayTarget": 20,
      "maxDelayTarget": 20,
      "numRetries": 3,
      "numMaxDelayRetries": 0,
      "numNoDelayRetries": 0,
      "numMinDelayRetries": 0,
      "backoffFunction": "linear"
    },
    "disableSubscriptionOverrides": false
  }
}
EOF
}
`, r)
}

// Test for https://github.com/hashicorp/terraform/issues/3660
func testAccAWSSNSTopicConfig_withFakeIAMRole(r string) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "test_topic" {
  name = "tf_acc_test_fake_iam_role_%s"
  policy = <<EOF
{
  "Statement": [
    {
      "Sid": "Stmt1445931846145",
      "Effect": "Allow",
      "Principal": {
        "AWS": "arn:aws:iam::012345678901:role/wooo"
			},
      "Action": "sns:Publish",
      "Resource": "arn:aws:sns:us-west-2::example"
    }
  ],
  "Version": "2012-10-17",
  "Id": "Policy1445931846145"
}
EOF
}
`, r)
}

func testAccAWSSNSTopicConfig_deliveryStatus(r string) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "test_topic" {
  depends_on                               = ["aws_iam_role_policy.example"]
  name                                     = "sns-delivery-status-topic-%s"
  application_success_feedback_role_arn    = "${aws_iam_role.example.arn}"
  application_success_feedback_sample_rate = 100
  application_failure_feedback_role_arn    = "${aws_iam_role.example.arn}"
  lambda_success_feedback_role_arn         = "${aws_iam_role.example.arn}"
  lambda_success_feedback_sample_rate      = 90
  lambda_failure_feedback_role_arn         = "${aws_iam_role.example.arn}"
  http_success_feedback_role_arn           = "${aws_iam_role.example.arn}"
  http_success_feedback_sample_rate        = 80
  http_failure_feedback_role_arn           = "${aws_iam_role.example.arn}"
  sqs_success_feedback_role_arn            = "${aws_iam_role.example.arn}"
  sqs_success_feedback_sample_rate         = 70
  sqs_failure_feedback_role_arn            = "${aws_iam_role.example.arn}"
}

resource "aws_iam_role" "example" {
  name = "sns-delivery-status-role-%s"
  path = "/"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "sns.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "example" {
  name = "sns-delivery-status-role-policy-%s"
  role = "${aws_iam_role.example.id}"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "logs:CreateLogGroup",
        "logs:CreateLogStream",
        "logs:PutLogEvents",
        "logs:PutMetricFilter",
        "logs:PutRetentionPolicy"
      ],
      "Resource": [
        "*"
      ]
    }
  ]
}
EOF
}
`, r, r, r)
}

func testAccAWSSNSTopicConfig_withEncryption(r string) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "test_topic" {
	name = "terraform-test-topic-%s"
	kms_master_key_id = "alias/aws/sns"
}
`, r)
}
