package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/iot"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSIoTTopicRule_importbasic(t *testing.T) {
	resourceName := "aws_iot_topic_rule.rule"
	rName := acctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSIoTTopicRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIoTTopicRule_basic(rName),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSIoTTopicRule_basic(t *testing.T) {
	rName := acctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSIoTTopicRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIoTTopicRule_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSIoTTopicRuleExists_basic("aws_iot_topic_rule.rule"),
					resource.TestCheckResourceAttr("aws_iot_topic_rule.rule", "name", fmt.Sprintf("test_rule_%s", rName)),
					resource.TestCheckResourceAttr("aws_iot_topic_rule.rule", "description", "Example rule"),
					resource.TestCheckResourceAttr("aws_iot_topic_rule.rule", "enabled", "true"),
					resource.TestCheckResourceAttr("aws_iot_topic_rule.rule", "sql", "SELECT * FROM 'topic/test'"),
					resource.TestCheckResourceAttr("aws_iot_topic_rule.rule", "sql_version", "2015-10-08"),
				),
			},
		},
	})
}

func TestAccAWSIoTTopicRule_cloudwatchalarm(t *testing.T) {
	rName := acctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSIoTTopicRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIoTTopicRule_cloudwatchalarm(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSIoTTopicRuleExists_basic("aws_iot_topic_rule.rule"),
				),
			},
		},
	})
}

func TestAccAWSIoTTopicRule_cloudwatchmetric(t *testing.T) {
	rName := acctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSIoTTopicRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIoTTopicRule_cloudwatchmetric(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSIoTTopicRuleExists_basic("aws_iot_topic_rule.rule"),
				),
			},
		},
	})
}

func TestAccAWSIoTTopicRule_dynamodb(t *testing.T) {
	rName := acctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSIoTTopicRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIoTTopicRule_dynamodb(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSIoTTopicRuleExists_basic("aws_iot_topic_rule.rule"),
				),
			},
			{
				Config: testAccAWSIoTTopicRule_dynamodb_rangekeys(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSIoTTopicRuleExists_basic("aws_iot_topic_rule.rule"),
				),
			},
		},
	})
}

func TestAccAWSIoTTopicRule_elasticsearch(t *testing.T) {
	rName := acctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSIoTTopicRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIoTTopicRule_elasticsearch(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSIoTTopicRuleExists_basic("aws_iot_topic_rule.rule"),
				),
			},
		},
	})
}

func TestAccAWSIoTTopicRule_firehose(t *testing.T) {
	rName := acctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSIoTTopicRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIoTTopicRule_firehose(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSIoTTopicRuleExists_basic("aws_iot_topic_rule.rule"),
				),
			},
		},
	})
}

func TestAccAWSIoTTopicRule_firehose_separator(t *testing.T) {
	rName := acctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSIoTTopicRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIoTTopicRule_firehose_separator(rName, "\n"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSIoTTopicRuleExists_basic("aws_iot_topic_rule.rule"),
				),
			},
			{
				Config: testAccAWSIoTTopicRule_firehose_separator(rName, ","),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSIoTTopicRuleExists_basic("aws_iot_topic_rule.rule"),
				),
			},
		},
	})
}

func TestAccAWSIoTTopicRule_kinesis(t *testing.T) {
	rName := acctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSIoTTopicRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIoTTopicRule_kinesis(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSIoTTopicRuleExists_basic("aws_iot_topic_rule.rule"),
				),
			},
		},
	})
}

func TestAccAWSIoTTopicRule_lambda(t *testing.T) {
	rName := acctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSIoTTopicRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIoTTopicRule_lambda(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSIoTTopicRuleExists_basic("aws_iot_topic_rule.rule"),
				),
			},
		},
	})
}

func TestAccAWSIoTTopicRule_republish(t *testing.T) {
	rName := acctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSIoTTopicRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIoTTopicRule_republish(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSIoTTopicRuleExists_basic("aws_iot_topic_rule.rule"),
				),
			},
		},
	})
}

func TestAccAWSIoTTopicRule_s3(t *testing.T) {
	rName := acctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSIoTTopicRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIoTTopicRule_s3(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSIoTTopicRuleExists_basic("aws_iot_topic_rule.rule"),
				),
			},
		},
	})
}

func TestAccAWSIoTTopicRule_sns(t *testing.T) {
	rName := acctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSIoTTopicRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIoTTopicRule_sns(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSIoTTopicRuleExists_basic("aws_iot_topic_rule.rule"),
				),
			},
		},
	})
}

func TestAccAWSIoTTopicRule_sqs(t *testing.T) {
	rName := acctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSIoTTopicRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIoTTopicRule_sqs(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSIoTTopicRuleExists_basic("aws_iot_topic_rule.rule"),
				),
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

		out, err := conn.ListTopicRules(&iot.ListTopicRulesInput{})

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

func testAccCheckAWSIoTTopicRuleExists_basic(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		return nil
	}
}

const testAccAWSIoTTopicRuleRole = `
resource "aws_iam_role" "iot_role" {
    name = "test_role_%[1]s"
    assume_role_policy = <<EOF
{
    "Version":"2012-10-17",
    "Statement":[{
        "Effect": "Allow",
        "Principal": {
            "Service": "iot.amazonaws.com"
        },
        "Action": "sts:AssumeRole"
    }]
}
EOF
}

resource "aws_iam_policy" "policy" {
    name = "test_policy_%[1]s"
    path = "/"
    description = "My test policy"
    policy = <<EOF
{
    "Version": "2012-10-17",
    "Statement": [{
        "Effect": "Allow",
        "Action": "*",
        "Resource": "*"
    }]
}
EOF
}

resource "aws_iam_policy_attachment" "attach_policy" {
    name = "test_policy_attachment_%[1]s"
    roles = ["${aws_iam_role.iot_role.name}"]
    policy_arn = "${aws_iam_policy.policy.arn}"
}
`

func testAccAWSIoTTopicRule_basic(rName string) string {
	return fmt.Sprintf(testAccAWSIoTTopicRuleRole+`
resource "aws_iot_topic_rule" "rule" {
  name = "test_rule_%[1]s"
  description = "Example rule"
  enabled = true
  sql = "SELECT * FROM 'topic/test'"
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
    role_arn     = "${aws_iam_role.iot_role.arn}"
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
    role_arn         = "${aws_iam_role.iot_role.arn}"
  }
}
`, rName)
}

func testAccAWSIoTTopicRule_dynamodb(rName string) string {
	return fmt.Sprintf(testAccAWSIoTTopicRuleRole+`
resource "aws_iot_topic_rule" "rule" {
  name = "test_rule_%[1]s"
  description = "Example rule"
  enabled = true
  sql = "SELECT * FROM 'topic/test'"
  sql_version = "2015-10-08"

  dynamodb {
    hash_key_field = "hash_key_field"
    hash_key_value = "hash_key_value"
    payload_field = "payload_field"
    role_arn = "${aws_iam_role.iot_role.arn}"
    table_name = "table_name"
  }
}
`, rName)
}

func testAccAWSIoTTopicRule_dynamodb_rangekeys(rName string) string {
	return fmt.Sprintf(testAccAWSIoTTopicRuleRole+`
resource "aws_iot_topic_rule" "rule" {
  name = "test_rule_%[1]s"
  description = "Example rule"
  enabled = true
  sql = "SELECT * FROM 'topic/test'"
  sql_version = "2015-10-08"

  dynamodb {
    hash_key_field = "hash_key_field"
    hash_key_value = "hash_key_value"
    payload_field = "payload_field"
    range_key_field = "range_key_field"
    range_key_value = "range_key_value"
    range_key_type  = "STRING"
    role_arn = "${aws_iam_role.iot_role.arn}"
    table_name = "table_name"
  }
}
`, rName)
}

func testAccAWSIoTTopicRule_elasticsearch(rName string) string {
	return fmt.Sprintf(testAccAWSIoTTopicRuleRole+`
data "aws_region" "current" {
  current = true
}

resource "aws_iot_topic_rule" "rule" {
  name        = "test_rule_%[1]s"
  description = "Example rule"
  enabled     = true
  sql         = "SELECT * FROM 'topic/test'"
  sql_version = "2015-10-08"

  elasticsearch {
    endpoint = "https://domain.${data.aws_region.current.name}.es.amazonaws.com"
    id       = "myIdentifier"
    index    = "myindex"
    type     = "mydocument"
    role_arn = "${aws_iam_role.iot_role.arn}"
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
    role_arn             = "${aws_iam_role.iot_role.arn}"
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
    role_arn             = "${aws_iam_role.iot_role.arn}"
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
    role_arn    = "${aws_iam_role.iot_role.arn}"
  }
}
`, rName)
}

func testAccAWSIoTTopicRule_lambda(rName string) string {
	return fmt.Sprintf(testAccAWSIoTTopicRuleRole+`
resource "aws_iot_topic_rule" "rule" {
  name        = "test_rule_%[1]s"
  description = "Example rule"
  enabled     = true
  sql         = "SELECT * FROM 'topic/test'"
  sql_version = "2015-10-08"

  lambda {
    function_arn = "arn:aws:lambda:us-east-1:123456789012:function:ProcessKinesisRecords"
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
    role_arn = "${aws_iam_role.iot_role.arn}"
    topic    = "mytopic"
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
    role_arn    = "${aws_iam_role.iot_role.arn}"
  }
}
`, rName)
}

func testAccAWSIoTTopicRule_sns(rName string) string {
	return fmt.Sprintf(testAccAWSIoTTopicRuleRole+`
resource "aws_iot_topic_rule" "rule" {
  name        = "test_rule_%[1]s"
  description = "Example rule"
  enabled     = true
  sql         = "SELECT * FROM 'topic/test'"
  sql_version = "2015-10-08"

  sns {
    role_arn = "${aws_iam_role.iot_role.arn}"
    target_arn = "arn:aws:sns:us-east-1:123456789012:my_corporate_topic"
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
    queue_url = "fakedata"
    role_arn  = "${aws_iam_role.iot_role.arn}"
    use_base64 = false
  }
}
`, rName)
}
