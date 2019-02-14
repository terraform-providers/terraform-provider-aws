package aws

import (
	"fmt"
	"regexp"
	"testing"

	events "github.com/aws/aws-sdk-go/service/cloudwatchevents"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSCloudWatchEventTarget_basic(t *testing.T) {
	var target events.Target
	rName1 := acctest.RandString(5)
	rName2 := acctest.RandString(5)
	ruleName := fmt.Sprintf("tf-acc-cw-event-rule-basic-%s", rName1)
	snsTopicName1 := fmt.Sprintf("tf-acc-%s", rName1)
	snsTopicName2 := fmt.Sprintf("tf-acc-%s", rName2)
	targetID1 := fmt.Sprintf("tf-acc-cw-target-%s", rName1)
	targetID2 := fmt.Sprintf("tf-acc-cw-target-%s", rName2)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudWatchEventTargetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudWatchEventTargetConfig(ruleName, snsTopicName1, targetID1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchEventTargetExists("aws_cloudwatch_event_target.moobar", &target),
					resource.TestCheckResourceAttr("aws_cloudwatch_event_target.moobar", "rule", ruleName),
					resource.TestCheckResourceAttr("aws_cloudwatch_event_target.moobar", "target_id", targetID1),
					resource.TestMatchResourceAttr("aws_cloudwatch_event_target.moobar", "arn",
						regexp.MustCompile(fmt.Sprintf(":%s$", snsTopicName1))),
				),
			},
			{
				Config: testAccAWSCloudWatchEventTargetConfig(ruleName, snsTopicName2, targetID2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchEventTargetExists("aws_cloudwatch_event_target.moobar", &target),
					resource.TestCheckResourceAttr("aws_cloudwatch_event_target.moobar", "rule", ruleName),
					resource.TestCheckResourceAttr("aws_cloudwatch_event_target.moobar", "target_id", targetID2),
					resource.TestMatchResourceAttr("aws_cloudwatch_event_target.moobar", "arn",
						regexp.MustCompile(fmt.Sprintf(":%s$", snsTopicName2))),
				),
			},
		},
	})
}

func TestAccAWSCloudWatchEventTarget_missingTargetId(t *testing.T) {
	var target events.Target
	rName := acctest.RandString(5)
	ruleName := fmt.Sprintf("tf-acc-cw-event-rule-missing-target-id-%s", rName)
	snsTopicName := fmt.Sprintf("tf-acc-%s", rName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudWatchEventTargetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudWatchEventTargetConfigMissingTargetId(ruleName, snsTopicName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchEventTargetExists("aws_cloudwatch_event_target.moobar", &target),
					resource.TestCheckResourceAttr("aws_cloudwatch_event_target.moobar", "rule", ruleName),
					resource.TestMatchResourceAttr("aws_cloudwatch_event_target.moobar", "arn",
						regexp.MustCompile(fmt.Sprintf(":%s$", snsTopicName))),
				),
			},
		},
	})
}

func TestAccAWSCloudWatchEventTarget_full(t *testing.T) {
	var target events.Target
	rName := acctest.RandString(5)
	ruleName := fmt.Sprintf("tf-acc-cw-event-rule-full-%s", rName)
	ssmDocumentName := acctest.RandomWithPrefix("tf_ssm_Document")
	targetID := fmt.Sprintf("tf-acc-cw-target-full-%s", rName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudWatchEventTargetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudWatchEventTargetConfig_full(ruleName, targetID, ssmDocumentName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchEventTargetExists("aws_cloudwatch_event_target.foobar", &target),
					resource.TestCheckResourceAttr("aws_cloudwatch_event_target.foobar", "rule", ruleName),
					resource.TestCheckResourceAttr("aws_cloudwatch_event_target.foobar", "target_id", targetID),
					resource.TestMatchResourceAttr("aws_cloudwatch_event_target.foobar", "arn",
						regexp.MustCompile("^arn:aws:kinesis:.*:stream/tf_ssm_Document")),
					resource.TestCheckResourceAttr("aws_cloudwatch_event_target.foobar", "input", "{ \"source\": [\"aws.cloudtrail\"] }\n"),
					resource.TestCheckResourceAttr("aws_cloudwatch_event_target.foobar", "input_path", ""),
				),
			},
		},
	})
}

func TestAccAWSCloudWatchEventTarget_ssmDocument(t *testing.T) {
	var target events.Target
	rName := acctest.RandomWithPrefix("tf_ssm_Document")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudWatchEventTargetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudWatchEventTargetConfigSsmDocument(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchEventTargetExists("aws_cloudwatch_event_target.test", &target),
				),
			},
		},
	})
}

func TestAccAWSCloudWatchEventTarget_ecs(t *testing.T) {
	var target events.Target
	rName := acctest.RandomWithPrefix("tf_ecs_target")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudWatchEventTargetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudWatchEventTargetConfigEcs(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchEventTargetExists("aws_cloudwatch_event_target.test", &target),
				),
			},
		},
	})
}

func TestAccAWSCloudWatchEventTarget_batch(t *testing.T) {
	var target events.Target
	rName := acctest.RandomWithPrefix("tf_batch_target")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudWatchEventTargetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudWatchEventTargetConfigBatch(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchEventTargetExists("aws_cloudwatch_event_target.test", &target),
				),
			},
		},
	})
}

func TestAccAWSCloudWatchEventTarget_kinesis(t *testing.T) {
	var target events.Target
	rName := acctest.RandomWithPrefix("tf_kinesis_target")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudWatchEventTargetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudWatchEventTargetConfigKinesis(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchEventTargetExists("aws_cloudwatch_event_target.test", &target),
				),
			},
		},
	})
}

func TestAccAWSCloudWatchEventTarget_sqs(t *testing.T) {
	var target events.Target
	rName := acctest.RandomWithPrefix("tf_sqs_target")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudWatchEventTargetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudWatchEventTargetConfigSqs(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchEventTargetExists("aws_cloudwatch_event_target.test", &target),
				),
			},
		},
	})
}

func TestAccAWSCloudWatchEventTarget_input_transformer(t *testing.T) {
	var target events.Target
	rName := acctest.RandomWithPrefix("tf_input_transformer")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudWatchEventTargetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudWatchEventTargetConfigInputTransformer(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchEventTargetExists("aws_cloudwatch_event_target.test", &target),
				),
			},
		},
	})
}

func testAccCheckCloudWatchEventTargetExists(n string, rule *events.Target) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := testAccProvider.Meta().(*AWSClient).cloudwatcheventsconn
		t, err := findEventTargetById(rs.Primary.Attributes["target_id"],
			rs.Primary.Attributes["rule"], nil, conn)
		if err != nil {
			return fmt.Errorf("Event Target not found: %s", err)
		}

		*rule = *t

		return nil
	}
}

func testAccCheckAWSCloudWatchEventTargetDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).cloudwatcheventsconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_cloudwatch_event_target" {
			continue
		}

		t, err := findEventTargetById(rs.Primary.Attributes["target_id"],
			rs.Primary.Attributes["rule"], nil, conn)
		if err == nil {
			return fmt.Errorf("CloudWatch Event Target %q still exists: %s",
				rs.Primary.ID, t)
		}
	}

	return nil
}

func testAccAWSCloudWatchEventTargetConfig(ruleName, snsTopicName, targetID string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_event_rule" "foo" {
	name = "%s"
	schedule_expression = "rate(1 hour)"
}

resource "aws_cloudwatch_event_target" "moobar" {
	rule = "${aws_cloudwatch_event_rule.foo.name}"
	target_id = "%s"
	arn = "${aws_sns_topic.moon.arn}"
}

resource "aws_sns_topic" "moon" {
	name = "%s"
}
`, ruleName, targetID, snsTopicName)
}

func testAccAWSCloudWatchEventTargetConfigMissingTargetId(ruleName, snsTopicName string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_event_rule" "foo" {
	name = "%s"
	schedule_expression = "rate(1 hour)"
}

resource "aws_cloudwatch_event_target" "moobar" {
	rule = "${aws_cloudwatch_event_rule.foo.name}"
	arn = "${aws_sns_topic.moon.arn}"
}

resource "aws_sns_topic" "moon" {
	name = "%s"
}
`, ruleName, snsTopicName)
}

func testAccAWSCloudWatchEventTargetConfig_full(ruleName, targetName, rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_event_rule" "foo" {
    name = "%s"
    schedule_expression = "rate(1 hour)"
    role_arn = "${aws_iam_role.role.arn}"
}

resource "aws_iam_role" "role" {
	name = "%s"
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

resource "aws_iam_role_policy" "test_policy" {
    name = "%s_policy"
    role = "${aws_iam_role.role.id}"
    policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "kinesis:PutRecord",
        "kinesis:PutRecords"
      ],
      "Resource": [
        "*"
      ],
      "Effect": "Allow"
    }
  ]
}
EOF
}

resource "aws_cloudwatch_event_target" "foobar" {
	rule = "${aws_cloudwatch_event_rule.foo.name}"
	target_id = "%s"
	input = <<INPUT
{ "source": ["aws.cloudtrail"] }
INPUT
	arn = "${aws_kinesis_stream.test_stream.arn}"
}

resource "aws_kinesis_stream" "test_stream" {
    name = "%s_kinesis_test"
    shard_count = 1
}`, ruleName, rName, rName, targetName, rName)
}

func testAccAWSCloudWatchEventTargetConfigSsmDocument(rName string) string {
	return fmt.Sprintf(`
resource "aws_ssm_document" "foo" {
  name = "%s"
  document_type = "Command"

  content = <<DOC
    {
      "schemaVersion": "1.2",
      "description": "Check ip configuration of a Linux instance.",
      "parameters": {

      },
      "runtimeConfig": {
        "aws:runShellScript": {
          "properties": [
            {
              "id": "0.aws:runShellScript",
              "runCommand": ["ifconfig"]
            }
          ]
        }
      }
    }
DOC
}

resource "aws_cloudwatch_event_rule" "console" {
  name        = "%s"
  description = "another_test"

  event_pattern = <<PATTERN
{
  "source": [
    "aws.autoscaling"
  ]
}
PATTERN
}

resource "aws_cloudwatch_event_target" "test" {

  arn = "${aws_ssm_document.foo.arn}"
  rule = "${aws_cloudwatch_event_rule.console.id}"
  role_arn = "${aws_iam_role.test_role.arn}"

  run_command_targets {
    key = "tag:Name"
    values = ["acceptance_test"]
  }
}

resource "aws_iam_role" "test_role" {
  name = "%s"

  assume_role_policy = <<EOF
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
EOF
}

resource "aws_iam_role_policy" "test_policy" {
  name = "%s"
  role = "${aws_iam_role.test_role.id}"

  policy = <<EOF
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Action": "ssm:*",
            "Effect": "Allow",
            "Resource": [
                "*"
            ]
        }
    ]
}
EOF
}`, rName, rName, rName, rName)
}

func testAccAWSCloudWatchEventTargetConfigEcs(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_event_rule" "schedule" {
  name        = "%s"
  description = "schedule_ecs_test"

	schedule_expression = "rate(5 minutes)"
}

resource "aws_vpc" "vpc" {
  cidr_block = "10.1.0.0/16"
}

resource "aws_subnet" "subnet" {
  vpc_id = "${aws_vpc.vpc.id}"
  cidr_block = "10.1.1.0/24"
}

resource "aws_cloudwatch_event_target" "test" {
	arn = "${aws_ecs_cluster.test.id}"
  rule = "${aws_cloudwatch_event_rule.schedule.id}"
  role_arn = "${aws_iam_role.test_role.arn}"

  ecs_target {
    task_count = 1
    task_definition_arn = "${aws_ecs_task_definition.task.arn}"
    launch_type = "FARGATE"
    network_configuration {
      subnets = ["${aws_subnet.subnet.id}"]
    }
  }
}

resource "aws_iam_role" "test_role" {
  name = "%s"

  assume_role_policy = <<EOF
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
EOF
}

resource "aws_iam_role_policy" "test_policy" {
  name = "%s"
  role = "${aws_iam_role.test_role.id}"

  policy = <<EOF
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "ecs:RunTask"
            ],
            "Resource": [
                "*"
            ]
        }
    ]
}
EOF
}

resource "aws_ecs_cluster" "test" {
  name = "%s"
}

resource "aws_ecs_task_definition" "task" {
  family                = "%s"
  cpu = 256
  memory = 512
  requires_compatibilities = ["FARGATE"]
  network_mode = "awsvpc"

  container_definitions = <<EOF
[
  {
    "name": "first",
    "image": "service-first",
    "cpu": 10,
    "memory": 512,
    "essential": true
  }
]
EOF
}`, rName, rName, rName, rName, rName)
}

func testAccAWSCloudWatchEventTargetConfigBatch(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_event_rule" "cloudwatch_event_rule" {
  name        = "%[1]s"
  description = "schedule_batch_test"
  schedule_expression = "rate(5 minutes)"
}

resource "aws_cloudwatch_event_target" "test" {
  arn = "${aws_batch_job_queue.batch_job_queue.arn}"
  rule = "${aws_cloudwatch_event_rule.cloudwatch_event_rule.id}"
  role_arn = "${aws_iam_role.event_iam_role.arn}"

  batch_target {
    job_definition = "${aws_batch_job_definition.batch_job_definition.arn}"
    job_name = "%[1]s"
  }

  depends_on = [
    "aws_batch_job_queue.batch_job_queue",
    "aws_batch_job_definition.batch_job_definition",
    "aws_iam_role.event_iam_role",
  ]
}

resource "aws_iam_role" "event_iam_role" {
  name = "event_%[1]s"
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Effect": "Allow",
      "Principal": {
        "Service": "events.amazonaws.com"
      }
    }
  ]
}
EOF
}

resource "aws_iam_role" "ecs_iam_role" {
  name = "ecs_%[1]s"
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Effect": "Allow",
      "Principal": {
        "Service": "ec2.amazonaws.com"
      }
    }
  ]
}
EOF
}

resource "aws_iam_role_policy_attachment" "ecs_policy_attachment" {
  role       = "${aws_iam_role.ecs_iam_role.name}"
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonEC2ContainerServiceforEC2Role"
}

resource "aws_iam_instance_profile" "iam_instance_profile" {
  name  = "ecs_%[1]s"
  role = "${aws_iam_role.ecs_iam_role.name}"
}

resource "aws_iam_role" "batch_iam_role" {
  name = "batch_%[1]s"
  assume_role_policy = <<EOF
{
    "Version": "2012-10-17",
    "Statement": [
    {
        "Action": "sts:AssumeRole",
        "Effect": "Allow",
        "Principal": {
          "Service": "batch.amazonaws.com"
        }
    }
    ]
}
EOF
}

resource "aws_iam_role_policy_attachment" "batch_policy_attachment" {
  role       = "${aws_iam_role.batch_iam_role.name}"
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSBatchServiceRole"
}

resource "aws_security_group" "security_group" {
  name = "%[1]s"
}

resource "aws_vpc" "vpc" {
  cidr_block = "10.1.0.0/16"
}

resource "aws_subnet" "subnet" {
  vpc_id = "${aws_vpc.vpc.id}"
  cidr_block = "10.1.1.0/24"
}

resource "aws_batch_compute_environment" "batch_compute_environment" {
  compute_environment_name = "%[1]s"
  compute_resources {
    instance_role = "${aws_iam_instance_profile.iam_instance_profile.arn}"
    instance_type = [
      "c4.large",
    ]
    max_vcpus = 16
    min_vcpus = 0
    security_group_ids = [
      "${aws_security_group.security_group.id}"
    ]
    subnets = [
      "${aws_subnet.subnet.id}"
    ]
    type = "EC2"
  }
  service_role = "${aws_iam_role.batch_iam_role.arn}"
  type = "MANAGED"
  depends_on = ["aws_iam_role_policy_attachment.batch_policy_attachment"]
}

resource "aws_batch_job_queue" "batch_job_queue" {
  name = "%[1]s"
  state = "ENABLED"
  priority = 1
  compute_environments = ["${aws_batch_compute_environment.batch_compute_environment.arn}"]
}

resource "aws_batch_job_definition" "batch_job_definition" {
	name = "%[1]s"
	type = "container"
	container_properties = <<CONTAINER_PROPERTIES
{
	"command": ["ls", "-la"],
	"image": "busybox",
	"memory": 512,
	"vcpus": 1,
	"volumes": [ ],
	"environment": [ ],
	"mountPoints": [ ],
    "ulimits": [ ]
}
CONTAINER_PROPERTIES
}
`, rName)
}

func testAccAWSCloudWatchEventTargetConfigKinesis(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_event_rule" "cloudwatch_event_rule" {
  name        = "%[1]s"
  description = "schedule_batch_test"
  schedule_expression = "rate(5 minutes)"
}

resource "aws_cloudwatch_event_target" "test" {
  arn = "${aws_kinesis_stream.kinesis_stream.arn}"
  rule = "${aws_cloudwatch_event_rule.cloudwatch_event_rule.id}"
  role_arn = "${aws_iam_role.iam_role.arn}"

  kinesis_target {
    partition_key_path = "$.detail"
  }
}

resource "aws_iam_role" "iam_role" {
  name = "event_%[1]s"
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Effect": "Allow",
      "Principal": {
        "Service": "events.amazonaws.com"
      }
    }
  ]
}
EOF
}

resource "aws_kinesis_stream" "kinesis_stream" {
  name = "%[1]s"
  shard_count = 1
}
`, rName)
}

func testAccAWSCloudWatchEventTargetConfigSqs(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_event_rule" "cloudwatch_event_rule" {
  name        = "%[1]s"
  description = "schedule_batch_test"
  schedule_expression = "rate(5 minutes)"
}

resource "aws_cloudwatch_event_target" "test" {
  arn = "${aws_sqs_queue.sqs_queue.arn}"
  rule = "${aws_cloudwatch_event_rule.cloudwatch_event_rule.id}"

  sqs_target {
    message_group_id = "event_group"
  }
}

resource "aws_sqs_queue" "sqs_queue" {
  name       = "%[1]s.fifo"
  fifo_queue = true
}`, rName)
}

func testAccAWSCloudWatchEventTargetConfigInputTransformer(rName string) string {
	return fmt.Sprintf(`

	resource "aws_iam_role" "iam_for_lambda" {
  name = "tf_acc_input_transformer"
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "lambda.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_lambda_function" "lambda" {
	function_name = "tf_acc_input_transformer"
	filename = "test-fixtures/lambdatest.zip"
  source_code_hash = "${base64sha256(file("test-fixtures/lambdatest.zip"))}"
  role = "${aws_iam_role.iam_for_lambda.arn}"
  handler = "exports.example"
	runtime = "nodejs8.10"
}

resource "aws_cloudwatch_event_rule" "schedule" {
  name        = "%s"
  description = "test_input_transformer"

	schedule_expression = "rate(5 minutes)"
}

resource "aws_cloudwatch_event_target" "test" {
	arn = "${aws_lambda_function.lambda.arn}"
  rule = "${aws_cloudwatch_event_rule.schedule.id}"

  input_transformer {
    input_paths {
      time = "$.time"
    }

    input_template = <<EOF
{
  "detail-type": "Scheduled Event",
  "source": "aws.events",
  "time": <time>,
  "region": "eu-west-1",
  "detail": {}
}
EOF
  }
}`, rName)
}
