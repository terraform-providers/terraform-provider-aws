package aws

import (
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	awspolicy "github.com/jen20/awspolicyequivalence"
)

func init() {
	resource.AddTestSweepers("aws_sns_topic", &resource.Sweeper{
		Name: "aws_sns_topic",
		F:    testSweepSnsTopics,
		Dependencies: []string{
			"aws_autoscaling_group",
			"aws_backup_vault_notifications",
			"aws_budgets_budget",
			"aws_config_delivery_channel",
			"aws_dax_cluster",
			"aws_db_event_subscription",
			"aws_elasticache_cluster",
			"aws_elasticache_replication_group",
			"aws_glacier_vault",
			"aws_iot_topic_rule",
			"aws_neptune_event_subscription",
			"aws_redshift_event_subscription",
			"aws_s3_bucket",
			"aws_ses_configuration_set",
			"aws_ses_domain_identity",
			"aws_ses_email_identity",
			"aws_ses_receipt_rule_set",
			"aws_sns_platform_application",
		},
	})
}

func testSweepSnsTopics(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.(*AWSClient).snsconn
	var sweeperErrs *multierror.Error

	err = conn.ListTopicsPages(&sns.ListTopicsInput{}, func(page *sns.ListTopicsOutput, isLast bool) bool {
		if page == nil {
			return !isLast
		}

		for _, topic := range page.Topics {
			arn := aws.StringValue(topic.TopicArn)

			log.Printf("[INFO] Deleting SNS Topic: %s", arn)
			_, err := conn.DeleteTopic(&sns.DeleteTopicInput{
				TopicArn: aws.String(arn),
			})
			if isAWSErr(err, sns.ErrCodeNotFoundException, "") {
				continue
			}
			if err != nil {
				sweeperErr := fmt.Errorf("error deleting SNS Topic (%s): %w", arn, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}
		}

		return !isLast
	})
	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping SNS Topics sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}
	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error retrieving SNS Topics: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func TestAccAWSSNSTopic_basic(t *testing.T) {
	attributes := make(map[string]string)
	resourceName := "aws_sns_topic.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSSNSTopicDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSNSTopicConfig_withGeneratedName,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSNSTopicExists(resourceName, attributes),
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

func TestAccAWSSNSTopic_name(t *testing.T) {
	attributes := make(map[string]string)
	resourceName := "aws_sns_topic.test"
	rName := acctest.RandString(10)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSSNSTopicDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSNSTopicConfig_withName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSNSTopicExists(resourceName, attributes),
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

func TestAccAWSSNSTopic_namePrefix(t *testing.T) {
	attributes := make(map[string]string)
	resourceName := "aws_sns_topic.test"
	startsWithPrefix := regexp.MustCompile("^terraform-test-topic-")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSSNSTopicDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSNSTopicConfig_withNamePrefix(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSNSTopicExists(resourceName, attributes),
					resource.TestMatchResourceAttr(resourceName, "name", startsWithPrefix),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"name_prefix"},
			},
		},
	})
}

func TestAccAWSSNSTopic_policy(t *testing.T) {
	attributes := make(map[string]string)
	resourceName := "aws_sns_topic.test"
	rName := acctest.RandString(10)
	expectedPolicy := fmt.Sprintf(`{"Statement":[{"Sid":"Stmt1445931846145","Effect":"Allow","Principal":{"AWS":"*"},"Action":"sns:Publish","Resource":"arn:%s:sns:%s::example"}],"Version":"2012-10-17","Id":"Policy1445931846145"}`, testAccGetPartition(), testAccGetRegion())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSSNSTopicDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSNSTopicWithPolicy(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSNSTopicExists(resourceName, attributes),
					testAccCheckAWSNSTopicHasPolicy(resourceName, expectedPolicy),
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

func TestAccAWSSNSTopic_withIAMRole(t *testing.T) {
	attributes := make(map[string]string)
	resourceName := "aws_sns_topic.test"
	rName := acctest.RandString(10)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSSNSTopicDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSNSTopicConfig_withIAMRole(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSNSTopicExists(resourceName, attributes),
				),
			},
		},
	})
}

func TestAccAWSSNSTopic_withFakeIAMRole(t *testing.T) {
	rName := acctest.RandString(10)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_sns_topic.test",
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
	resourceName := "aws_sns_topic.test"
	rName := acctest.RandString(10)
	expectedPolicy := `{"http":{"defaultHealthyRetryPolicy": {"minDelayTarget": 20,"maxDelayTarget": 20,"numMaxDelayRetries": 0,"numRetries": 3,"numNoDelayRetries": 0,"numMinDelayRetries": 0,"backoffFunction": "linear"},"disableSubscriptionOverrides": false}}`

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSSNSTopicDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSNSTopicConfig_withDeliveryPolicy(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSNSTopicExists(resourceName, attributes),
					testAccCheckAWSNSTopicHasDeliveryPolicy(resourceName, expectedPolicy),
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

func TestAccAWSSNSTopic_deliveryStatus(t *testing.T) {
	attributes := make(map[string]string)
	resourceName := "aws_sns_topic.test"
	iamRoleResourceName := "aws_iam_role.example"

	rName := acctest.RandString(10)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSSNSTopicDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSNSTopicConfig_deliveryStatus(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSNSTopicExists(resourceName, attributes),
					resource.TestCheckResourceAttrPair(resourceName, "application_success_feedback_role_arn", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "application_success_feedback_sample_rate", "100"),
					resource.TestCheckResourceAttrPair(resourceName, "application_failure_feedback_role_arn", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "lambda_success_feedback_role_arn", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "lambda_success_feedback_sample_rate", "90"),
					resource.TestCheckResourceAttrPair(resourceName, "lambda_failure_feedback_role_arn", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "http_success_feedback_role_arn", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "http_success_feedback_sample_rate", "80"),
					resource.TestCheckResourceAttrPair(resourceName, "http_failure_feedback_role_arn", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "sqs_success_feedback_role_arn", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "sqs_success_feedback_sample_rate", "70"),
					resource.TestCheckResourceAttrPair(resourceName, "sqs_failure_feedback_role_arn", iamRoleResourceName, "arn"),
				),
			},
		},
	})
}

func TestAccAWSSNSTopic_encryption(t *testing.T) {
	attributes := make(map[string]string)
	resourceName := "aws_sns_topic.test"
	rName := acctest.RandString(10)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSSNSTopicDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSNSTopicConfig_withEncryption(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSNSTopicExists(resourceName, attributes),
					resource.TestCheckResourceAttr(resourceName, "kms_master_key_id", "alias/aws/sns"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSSNSTopicConfig_withName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSNSTopicExists(resourceName, attributes),
					resource.TestCheckResourceAttr(resourceName, "kms_master_key_id", ""),
				),
			},
		},
	})
}

func TestAccAWSSNSTopic_tags(t *testing.T) {
	attributes := make(map[string]string)
	resourceName := "aws_sns_topic.test"
	rName := acctest.RandString(10)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSSNSTopicDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSNSTopicConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSNSTopicExists(resourceName, attributes),
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
				Config: testAccAWSSNSTopicConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSNSTopicExists(resourceName, attributes),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSSNSTopicConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSNSTopicExists(resourceName, attributes),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
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
			return fmt.Errorf("no Queue URL specified")
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
			return fmt.Errorf("Non-equivalent policy error:\n\nexpected: %s\n\n     got: %s",
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
			return fmt.Errorf("no Queue URL specified")
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
			return fmt.Errorf("Non-equivalent delivery policy error:\n\nexpected: %s\n\n     got: %s",
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
		return fmt.Errorf("SNS topic (%s) exists when it should be destroyed", rs.Primary.ID)
	}

	return nil
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
resource "aws_sns_topic" "test" {}
`

func testAccAWSSNSTopicConfig_withName(r string) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "test" {
  name = "terraform-test-topic-%s"
}
`, r)
}

func testAccAWSSNSTopicConfig_withNamePrefix() string {
	return `
resource "aws_sns_topic" "test" {
  name_prefix = "terraform-test-topic-"
}
`
}

func testAccAWSSNSTopicWithPolicy(r string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

data "aws_region" "current" {}

resource "aws_sns_topic" "test" {
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
      "Resource": "arn:${data.aws_partition.current.partition}:sns:${data.aws_region.current.name}::example"
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
data "aws_partition" "current" {}

resource "aws_iam_role" "example" {
  name = "tf_acc_test_%[1]s"
  path = "/test/"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "ec2.${data.aws_partition.current.dns_suffix}"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

data "aws_region" "current" {}

resource "aws_sns_topic" "test" {
  name = "tf-acc-test-with-iam-role-%[1]s"

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
      "Resource": "arn:${data.aws_partition.current.partition}:sns:${data.aws_region.current.name}::example"
    }
  ],
  "Version": "2012-10-17",
  "Id": "Policy1445931846145"
}
EOF
}
`, r)
}

// Test for https://github.com/hashicorp/terraform/issues/14024
func testAccAWSSNSTopicConfig_withDeliveryPolicy(r string) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "test" {
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
data "aws_partition" "current" {}

data "aws_region" "current" {}

resource "aws_sns_topic" "test" {
  name = "tf_acc_test_fake_iam_role_%s"

  policy = <<EOF
{
  "Statement": [
    {
      "Sid": "Stmt1445931846145",
      "Effect": "Allow",
      "Principal": {
        "AWS": "arn:${data.aws_partition.current.partition}:iam::012345678901:role/wooo"
      },
      "Action": "sns:Publish",
      "Resource": "arn:${data.aws_partition.current.partition}:sns:${data.aws_region.current.name}::example"
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
resource "aws_sns_topic" "test" {
  depends_on                               = [aws_iam_role_policy.example]
  name                                     = "sns-delivery-status-topic-%[1]s"
  application_success_feedback_role_arn    = aws_iam_role.example.arn
  application_success_feedback_sample_rate = 100
  application_failure_feedback_role_arn    = aws_iam_role.example.arn
  lambda_success_feedback_role_arn         = aws_iam_role.example.arn
  lambda_success_feedback_sample_rate      = 90
  lambda_failure_feedback_role_arn         = aws_iam_role.example.arn
  http_success_feedback_role_arn           = aws_iam_role.example.arn
  http_success_feedback_sample_rate        = 80
  http_failure_feedback_role_arn           = aws_iam_role.example.arn
  sqs_success_feedback_role_arn            = aws_iam_role.example.arn
  sqs_success_feedback_sample_rate         = 70
  sqs_failure_feedback_role_arn            = aws_iam_role.example.arn
}

data "aws_partition" "current" {}

resource "aws_iam_role" "example" {
  name = "sns-delivery-status-role-%[1]s"
  path = "/"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "sns.${data.aws_partition.current.dns_suffix}"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "example" {
  name = "sns-delivery-status-role-policy-%[1]s"
  role = aws_iam_role.example.id

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
`, r)
}

func testAccAWSSNSTopicConfig_withEncryption(r string) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "test" {
  name              = "terraform-test-topic-%s"
  kms_master_key_id = "alias/aws/sns"
}
`, r)
}

func testAccAWSSNSTopicConfigTags1(r, tag1Key, tag1Value string) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "test" {
  name = "terraform-test-topic-%s"

  tags = {
    %q = %q
  }
}
`, r, tag1Key, tag1Value)
}

func testAccAWSSNSTopicConfigTags2(r, tag1Key, tag1Value, tag2Key, tag2Value string) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "test" {
  name = "terraform-test-topic-%s"

  tags = {
    %q = %q
    %q = %q
  }
}
`, r, tag1Key, tag1Value, tag2Key, tag2Value)
}
