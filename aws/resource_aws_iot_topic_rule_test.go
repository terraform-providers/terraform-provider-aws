package aws

import (
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iot"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	resource.AddTestSweepers("aws_iot_topic_rule", &resource.Sweeper{
		Name: "aws_iot_topic_rule",
		F:    testSweepIotTopicRules,
	})
}

func testSweepIotTopicRules(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.(*AWSClient).iotconn
	input := &iot.ListTopicRulesInput{}
	var sweeperErrs *multierror.Error

	for {
		output, err := conn.ListTopicRules(input)
		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping IoT Topic Rules sweep for %s: %s", region, err)
			return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
		}
		if err != nil {
			sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error retrieving IoT Topic Rules: %w", err))
			return sweeperErrs
		}

		for _, rule := range output.Rules {
			name := aws.StringValue(rule.RuleName)

			log.Printf("[INFO] Deleting IoT Topic Rule: %s", name)
			_, err := conn.DeleteTopicRule(&iot.DeleteTopicRuleInput{
				RuleName: aws.String(name),
			})
			if isAWSErr(err, iot.ErrCodeUnauthorizedException, "") {
				continue
			}
			if err != nil {
				sweeperErr := fmt.Errorf("error deleting IoT Topic Rule (%s): %w", name, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}
		}

		if aws.StringValue(output.NextToken) == "" {
			break
		}
		input.NextToken = output.NextToken
	}

	return sweeperErrs.ErrorOrNil()
}

func TestAccAWSIoTTopicRule_basic(t *testing.T) {
	rName := acctest.RandString(5)
	resourceName := "aws_iot_topic_rule.rule"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSIoTTopicRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIoTTopicRule_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSIoTTopicRuleExists("aws_iot_topic_rule.rule"),
					resource.TestCheckResourceAttr("aws_iot_topic_rule.rule", "name", fmt.Sprintf("test_rule_%s", rName)),
					resource.TestCheckResourceAttr("aws_iot_topic_rule.rule", "description", "Example rule"),
					resource.TestCheckResourceAttr("aws_iot_topic_rule.rule", "enabled", "true"),
					resource.TestCheckResourceAttr("aws_iot_topic_rule.rule", "sql", "SELECT * FROM 'topic/test'"),
					resource.TestCheckResourceAttr("aws_iot_topic_rule.rule", "sql_version", "2015-10-08"),
					resource.TestCheckResourceAttr("aws_iot_topic_rule.rule", "tags.%", "0"),
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

func TestAccAWSIoTTopicRule_cloudwatchalarm(t *testing.T) {
	rName := acctest.RandString(5)
	resourceName := "aws_iot_topic_rule.rule"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSIoTTopicRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIoTTopicRule_cloudwatchalarm(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSIoTTopicRuleExists("aws_iot_topic_rule.rule"),
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

func TestAccAWSIoTTopicRule_cloudwatchmetric(t *testing.T) {
	rName := acctest.RandString(5)
	resourceName := "aws_iot_topic_rule.rule"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSIoTTopicRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIoTTopicRule_cloudwatchmetric(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSIoTTopicRuleExists("aws_iot_topic_rule.rule"),
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

func TestAccAWSIoTTopicRule_dynamodb(t *testing.T) {
	rName := acctest.RandString(5)
	resourceName := "aws_iot_topic_rule.rule"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSIoTTopicRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIoTTopicRule_dynamodb(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSIoTTopicRuleExists("aws_iot_topic_rule.rule"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSIoTTopicRule_dynamodb_rangekeys(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSIoTTopicRuleExists("aws_iot_topic_rule.rule"),
				),
			},
		},
	})
}

func TestAccAWSIoTTopicRule_dynamoDbv2(t *testing.T) {
	rName := acctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSIoTTopicRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIoTTopicRule_dynamoDbv2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSIoTTopicRuleExists("aws_iot_topic_rule.rule"),
				),
			},
		},
	})
}

func TestAccAWSIoTTopicRule_elasticsearch(t *testing.T) {
	rName := acctest.RandString(5)
	resourceName := "aws_iot_topic_rule.rule"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSIoTTopicRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIoTTopicRule_elasticsearch(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSIoTTopicRuleExists("aws_iot_topic_rule.rule"),
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

func TestAccAWSIoTTopicRule_firehose(t *testing.T) {
	rName := acctest.RandString(5)
	resourceName := "aws_iot_topic_rule.rule"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSIoTTopicRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIoTTopicRule_firehose(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSIoTTopicRuleExists("aws_iot_topic_rule.rule"),
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

func TestAccAWSIoTTopicRule_firehose_separator(t *testing.T) {
	rName := acctest.RandString(5)
	resourceName := "aws_iot_topic_rule.rule"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSIoTTopicRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIoTTopicRule_firehose_separator(rName, "\n"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSIoTTopicRuleExists("aws_iot_topic_rule.rule"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSIoTTopicRule_firehose_separator(rName, ","),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSIoTTopicRuleExists("aws_iot_topic_rule.rule"),
				),
			},
		},
	})
}

func TestAccAWSIoTTopicRule_kinesis(t *testing.T) {
	rName := acctest.RandString(5)
	resourceName := "aws_iot_topic_rule.rule"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSIoTTopicRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIoTTopicRule_kinesis(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSIoTTopicRuleExists("aws_iot_topic_rule.rule"),
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

func TestAccAWSIoTTopicRule_lambda(t *testing.T) {
	rName := acctest.RandString(5)
	resourceName := "aws_iot_topic_rule.rule"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSIoTTopicRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIoTTopicRule_lambda(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSIoTTopicRuleExists("aws_iot_topic_rule.rule"),
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

func TestAccAWSIoTTopicRule_republish(t *testing.T) {
	rName := acctest.RandString(5)
	resourceName := "aws_iot_topic_rule.rule"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSIoTTopicRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIoTTopicRule_republish(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSIoTTopicRuleExists("aws_iot_topic_rule.rule"),
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

func TestAccAWSIoTTopicRule_republish_with_qos(t *testing.T) {
	rName := acctest.RandString(5)
	resourceName := "aws_iot_topic_rule.rule"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSIoTTopicRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIoTTopicRule_republish_with_qos(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSIoTTopicRuleExists("aws_iot_topic_rule.rule"),
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

func TestAccAWSIoTTopicRule_s3(t *testing.T) {
	rName := acctest.RandString(5)
	resourceName := "aws_iot_topic_rule.rule"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSIoTTopicRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIoTTopicRule_s3(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSIoTTopicRuleExists("aws_iot_topic_rule.rule"),
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

func TestAccAWSIoTTopicRule_sns(t *testing.T) {
	rName := acctest.RandString(5)
	resourceName := "aws_iot_topic_rule.rule"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSIoTTopicRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIoTTopicRule_sns(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSIoTTopicRuleExists("aws_iot_topic_rule.rule"),
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

func TestAccAWSIoTTopicRule_sqs(t *testing.T) {
	rName := acctest.RandString(5)
	resourceName := "aws_iot_topic_rule.rule"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSIoTTopicRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIoTTopicRule_sqs(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSIoTTopicRuleExists("aws_iot_topic_rule.rule"),
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

func TestAccAWSIoTTopicRule_step_functions(t *testing.T) {
	rName := acctest.RandString(5)
	resourceName := "aws_iot_topic_rule.rule"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSIoTTopicRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIoTTopicRule_step_functions(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSIoTTopicRuleExists("aws_iot_topic_rule.rule"),
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

func TestAccAWSIoTTopicRule_iot_analytics(t *testing.T) {
	rName := acctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSIoTTopicRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIoTTopicRule_iot_analytics(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSIoTTopicRuleExists("aws_iot_topic_rule.rule"),
				),
			},
		},
	})
}

func TestAccAWSIoTTopicRule_iot_events(t *testing.T) {
	rName := acctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSIoTTopicRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIoTTopicRule_iot_events(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSIoTTopicRuleExists("aws_iot_topic_rule.rule"),
				),
			},
		},
	})
}

func TestAccAWSIoTTopicRule_Tags(t *testing.T) {
	rName := acctest.RandString(5)
	resourceName := "aws_iot_topic_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSIoTTopicRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIoTTopicRuleTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSIoTTopicRuleExists(resourceName),
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
				Config: testAccAWSIoTTopicRuleTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSIoTTopicRuleExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSIoTTopicRuleTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSIoTTopicRuleExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccAWSIoTTopicRule_errorAction(t *testing.T) {
	rName := acctest.RandString(5)
	resourceName := "aws_iot_topic_rule.rule"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSIoTTopicRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIoTTopicRule_errorAction(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSIoTTopicRuleExists("aws_iot_topic_rule.rule"),
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

func testAccCheckAWSIoTTopicRuleDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).iotconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_iot_topic_rule" {
			continue
		}

		input := &iot.ListTopicRulesInput{}

		out, err := conn.ListTopicRules(input)

		if err != nil {
			return err
		}

		for _, r := range out.Rules {
			if *r.RuleName == rs.Primary.ID {
				return fmt.Errorf("IoT topic rule still exists:\n%s", r)
			}
		}

	}

	return nil
}

func testAccCheckAWSIoTTopicRuleExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		conn := testAccProvider.Meta().(*AWSClient).iotconn
		input := &iot.ListTopicRulesInput{}

		output, err := conn.ListTopicRules(input)

		if err != nil {
			return err
		}

		for _, rule := range output.Rules {
			if aws.StringValue(rule.RuleName) == rs.Primary.ID {
				return nil
			}
		}

		return fmt.Errorf("IoT Topic Rule (%s) not found", rs.Primary.ID)
	}
}

const testAccAWSIoTTopicRuleRole = `
data "aws_partition" "current" {}

resource "aws_iam_role" "iot_role" {
  name = "test_role_%[1]s"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "iot.${data.aws_partition.current.dns_suffix}"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF

}

resource "aws_iam_policy" "policy" {
  name        = "test_policy_%[1]s"
  path        = "/"
  description = "My test policy"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": "*",
      "Resource": "*"
    }
  ]
}
EOF

}

resource "aws_iam_policy_attachment" "attach_policy" {
  name       = "test_policy_attachment_%[1]s"
  roles      = [aws_iam_role.iot_role.name]
  policy_arn = aws_iam_policy.policy.arn
}
`

func testAccAWSIoTTopicRule_basic(rName string) string {
	return fmt.Sprintf(testAccAWSIoTTopicRuleRole+`
resource "aws_iot_topic_rule" "rule" {
  name        = "test_rule_%[1]s"
  description = "Example rule"
  enabled     = true
  sql         = "SELECT * FROM 'topic/test'"
  sql_version = "2015-10-08"
}
`, rName)
}

func testAccAWSIoTTopicRule_cloudwatchalarm(rName string) string {
	return fmt.Sprintf(testAccAWSIoTTopicRuleRole+`
resource "aws_iot_topic_rule" "rule" {
  name        = "test_rule_%[1]s"
  description = "Example rule"
  enabled     = true
  sql         = "SELECT * FROM 'topic/test'"
  sql_version = "2015-10-08"

  cloudwatch_alarm {
    alarm_name   = "myalarm"
    role_arn     = aws_iam_role.iot_role.arn
    state_reason = "test"
    state_value  = "OK"
  }
}
`, rName)
}

func testAccAWSIoTTopicRule_cloudwatchmetric(rName string) string {
	return fmt.Sprintf(testAccAWSIoTTopicRuleRole+`
resource "aws_iot_topic_rule" "rule" {
  name        = "test_rule_%[1]s"
  description = "Example rule"
  enabled     = true
  sql         = "SELECT * FROM 'topic/test'"
  sql_version = "2015-10-08"

  cloudwatch_metric {
    metric_name      = "FakeData"
    metric_namespace = "FakeData"
    metric_value     = "FakeData"
    metric_unit      = "FakeData"
    role_arn         = aws_iam_role.iot_role.arn
  }
}
`, rName)
}

func testAccAWSIoTTopicRule_dynamoDbv2(rName string) string {
	return fmt.Sprintf(testAccAWSIoTTopicRuleRole+`
resource "aws_iot_topic_rule" "rule" {
  name        = "test_rule_%[1]s"
  description = "Example rule"
  enabled     = true
  sql         = "SELECT field as column_name FROM 'topic/test'"
  sql_version = "2015-10-08"

  dynamodbv2 {
    put_item {
      table_name = "test"
    }

    role_arn = aws_iam_role.iot_role.arn
  }
}
`, rName)
}

func testAccAWSIoTTopicRule_dynamodb(rName string) string {
	return fmt.Sprintf(testAccAWSIoTTopicRuleRole+`
resource "aws_iot_topic_rule" "rule" {
  name        = "test_rule_%[1]s"
  description = "Example rule"
  enabled     = true
  sql         = "SELECT * FROM 'topic/test'"
  sql_version = "2015-10-08"

  dynamodb {
    hash_key_field = "hash_key_field"
    hash_key_value = "hash_key_value"
    payload_field  = "payload_field"
    role_arn       = aws_iam_role.iot_role.arn
    table_name     = "table_name"
  }
}
`, rName)
}

func testAccAWSIoTTopicRule_dynamodb_rangekeys(rName string) string {
	return fmt.Sprintf(testAccAWSIoTTopicRuleRole+`
resource "aws_iot_topic_rule" "rule" {
  name        = "test_rule_%[1]s"
  description = "Example rule"
  enabled     = true
  sql         = "SELECT * FROM 'topic/test'"
  sql_version = "2015-10-08"

  dynamodb {
    hash_key_field  = "hash_key_field"
    hash_key_value  = "hash_key_value"
    payload_field   = "payload_field"
    range_key_field = "range_key_field"
    range_key_value = "range_key_value"
    range_key_type  = "STRING"
    role_arn        = aws_iam_role.iot_role.arn
    table_name      = "table_name"
    operation       = "INSERT"
  }
}
`, rName)
}

func testAccAWSIoTTopicRule_elasticsearch(rName string) string {
	return fmt.Sprintf(testAccAWSIoTTopicRuleRole+`
data "aws_region" "current" {}

resource "aws_iot_topic_rule" "rule" {
  name        = "test_rule_%[1]s"
  description = "Example rule"
  enabled     = true
  sql         = "SELECT * FROM 'topic/test'"
  sql_version = "2015-10-08"

  elasticsearch {
    endpoint = "https://domain.${data.aws_region.current.name}.es.${data.aws_partition.current.dns_suffix}"
    id       = "myIdentifier"
    index    = "myindex"
    type     = "mydocument"
    role_arn = aws_iam_role.iot_role.arn
  }
}
`, rName)
}

func testAccAWSIoTTopicRule_firehose(rName string) string {
	return fmt.Sprintf(testAccAWSIoTTopicRuleRole+`
resource "aws_iot_topic_rule" "rule" {
  name        = "test_rule_%[1]s"
  description = "Example rule"
  enabled     = true
  sql         = "SELECT * FROM 'topic/test'"
  sql_version = "2015-10-08"

  firehose {
    delivery_stream_name = "mystream"
    role_arn             = aws_iam_role.iot_role.arn
  }
}
`, rName)
}

func testAccAWSIoTTopicRule_firehose_separator(rName, separator string) string {
	return fmt.Sprintf(testAccAWSIoTTopicRuleRole+`
resource "aws_iot_topic_rule" "rule" {
  name        = "test_rule_%[1]s"
  description = "Example rule"
  enabled     = true
  sql         = "SELECT * FROM 'topic/test'"
  sql_version = "2015-10-08"

  firehose {
    delivery_stream_name = "mystream"
    role_arn             = aws_iam_role.iot_role.arn
    separator            = %q
  }
}
`, rName, separator)
}

func testAccAWSIoTTopicRule_kinesis(rName string) string {
	return fmt.Sprintf(testAccAWSIoTTopicRuleRole+`
resource "aws_iot_topic_rule" "rule" {
  name        = "test_rule_%[1]s"
  description = "Example rule"
  enabled     = true
  sql         = "SELECT * FROM 'topic/test'"
  sql_version = "2015-10-08"

  kinesis {
    stream_name = "mystream"
    role_arn    = aws_iam_role.iot_role.arn
  }
}
`, rName)
}

func testAccAWSIoTTopicRule_lambda(rName string) string {
	return fmt.Sprintf(testAccAWSIoTTopicRuleRole+`
data "aws_region" "current" {}

resource "aws_iot_topic_rule" "rule" {
  name        = "test_rule_%[1]s"
  description = "Example rule"
  enabled     = true
  sql         = "SELECT * FROM 'topic/test'"
  sql_version = "2015-10-08"

  lambda {
    function_arn = "arn:${data.aws_partition.current.partition}:lambda:${data.aws_region.current.name}:123456789012:function:ProcessKinesisRecords"
  }
}
`, rName)
}

func testAccAWSIoTTopicRule_republish(rName string) string {
	return fmt.Sprintf(testAccAWSIoTTopicRuleRole+`
resource "aws_iot_topic_rule" "rule" {
  name        = "test_rule_%[1]s"
  description = "Example rule"
  enabled     = true
  sql         = "SELECT * FROM 'topic/test'"
  sql_version = "2015-10-08"

  republish {
    role_arn = aws_iam_role.iot_role.arn
    topic    = "mytopic"
  }
}
`, rName)
}

func testAccAWSIoTTopicRule_republish_with_qos(rName string) string {
	return fmt.Sprintf(testAccAWSIoTTopicRuleRole+`
resource "aws_iot_topic_rule" "rule" {
  name        = "test_rule_%[1]s"
  description = "Example rule"
  enabled     = true
  sql         = "SELECT * FROM 'topic/test'"
  sql_version = "2015-10-08"

  republish {
    role_arn = aws_iam_role.iot_role.arn
    topic    = "mytopic"
    qos      = 1
  }
}
`, rName)
}

func testAccAWSIoTTopicRule_s3(rName string) string {
	return fmt.Sprintf(testAccAWSIoTTopicRuleRole+`
resource "aws_iot_topic_rule" "rule" {
  name        = "test_rule_%[1]s"
  description = "Example rule"
  enabled     = true
  sql         = "SELECT * FROM 'topic/test'"
  sql_version = "2015-10-08"

  s3 {
    bucket_name = "mybucket"
    key         = "mykey"
    role_arn    = aws_iam_role.iot_role.arn
  }
}
`, rName)
}

func testAccAWSIoTTopicRule_sns(rName string) string {
	return fmt.Sprintf(testAccAWSIoTTopicRuleRole+`
data "aws_region" "current" {}

resource "aws_iot_topic_rule" "rule" {
  name        = "test_rule_%[1]s"
  description = "Example rule"
  enabled     = true
  sql         = "SELECT * FROM 'topic/test'"
  sql_version = "2015-10-08"

  sns {
    role_arn   = aws_iam_role.iot_role.arn
    target_arn = "arn:${data.aws_partition.current.partition}:sns:${data.aws_region.current.name}:123456789012:my_corporate_topic"
  }
}
`, rName)
}

func testAccAWSIoTTopicRule_sqs(rName string) string {
	return fmt.Sprintf(testAccAWSIoTTopicRuleRole+`
resource "aws_iot_topic_rule" "rule" {
  name        = "test_rule_%[1]s"
  description = "Example rule"
  enabled     = true
  sql         = "SELECT * FROM 'topic/test'"
  sql_version = "2015-10-08"

  sqs {
    queue_url  = "fakedata"
    role_arn   = aws_iam_role.iot_role.arn
    use_base64 = false
  }
}
`, rName)
}

func testAccAWSIoTTopicRule_step_functions(rName string) string {
	return fmt.Sprintf(testAccAWSIoTTopicRuleRole+`
resource "aws_iot_topic_rule" "rule" {
  name        = "test_rule_%[1]s"
  description = "Example rule"
  enabled     = true
  sql         = "SELECT * FROM 'topic/test'"
  sql_version = "2015-10-08"

  step_functions {
    execution_name_prefix = "myprefix"
    state_machine_name    = "mystatemachine"
    role_arn              = aws_iam_role.iot_role.arn
  }
}
`, rName)
}

func testAccAWSIoTTopicRule_iot_analytics(rName string) string {
	return fmt.Sprintf(testAccAWSIoTTopicRuleRole+`
resource "aws_iot_topic_rule" "rule" {
  name        = "test_rule_%[1]s"
  description = "Example rule"
  enabled     = true
  sql         = "SELECT * FROM 'topic/test'"
  sql_version = "2015-10-08"

  iot_analytics {
    channel_name = "fakedata"
    role_arn     = aws_iam_role.iot_role.arn
  }
}
`, rName)
}

func testAccAWSIoTTopicRule_iot_events(rName string) string {
	return fmt.Sprintf(testAccAWSIoTTopicRuleRole+`
resource "aws_iot_topic_rule" "rule" {
  name        = "test_rule_%[1]s"
  description = "Example rule"
  enabled     = true
  sql         = "SELECT * FROM 'topic/test'"
  sql_version = "2015-10-08"

  iot_events {
    input_name = "fake_input_name"
    role_arn   = aws_iam_role.iot_role.arn
    message_id = "fake_message_id"
  }
}
`, rName)
}

func testAccAWSIoTTopicRuleTags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(testAccAWSIoTTopicRuleRole+`
resource "aws_iot_topic_rule" "test" {
  name        = "test_rule_%[1]s"
  enabled     = true
  sql         = "SELECT * FROM 'topic/test'"
  sql_version = "2015-10-08"

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccAWSIoTTopicRuleTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(testAccAWSIoTTopicRuleRole+`
resource "aws_iot_topic_rule" "test" {
  name        = "test_rule_%[1]s"
  enabled     = true
  sql         = "SELECT * FROM 'topic/test'"
  sql_version = "2015-10-08"

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccAWSIoTTopicRule_errorAction(rName string) string {
	return fmt.Sprintf(testAccAWSIoTTopicRuleRole+`
resource "aws_iot_topic_rule" "rule" {
  name        = "test_rule_%[1]s"
  description = "Example rule"
  enabled     = true
  sql         = "SELECT * FROM 'topic/test'"
  sql_version = "2015-10-08"

  kinesis {
    stream_name = "mystream"
    role_arn    = aws_iam_role.iot_role.arn
  }

  error_action {
    kinesis {
      stream_name = "mystream"
      role_arn    = aws_iam_role.iot_role.arn
    }
  }
}
`, rName)
}
