---
layout: "aws"
page_title: "AWS: aws_cloudwatch_event_target"
sidebar_current: "docs-aws-resource-cloudwatch-event-target"
description: |-
  Provides a CloudWatch Event Target resource.
---

# aws_cloudwatch_event_target

Provides a CloudWatch Event Target resource.

## Example Usage

```hcl
resource "aws_cloudwatch_event_target" "yada" {
  target_id = "Yada"
  rule      = "${aws_cloudwatch_event_rule.console.name}"
  arn       = "${aws_kinesis_stream.test_stream.arn}"

  run_command_targets {
    key    = "tag:Name"
    values = ["FooBar"]
  }

  run_command_targets {
    key    = "InstanceIds"
    values = ["i-162058cd308bffec2"]
  }
}

resource "aws_cloudwatch_event_rule" "console" {
  name        = "capture-ec2-scaling-events"
  description = "Capture all EC2 scaling events"

  event_pattern = <<PATTERN
{
  "source": [
    "aws.autoscaling"
  ],
  "detail-type": [
    "EC2 Instance Launch Successful",
    "EC2 Instance Terminate Successful",
    "EC2 Instance Launch Unsuccessful",
    "EC2 Instance Terminate Unsuccessful"
  ]
}
PATTERN
}

resource "aws_kinesis_stream" "test_stream" {
  name        = "terraform-kinesis-test"
  shard_count = 1
}
```

## Example SSM Document Usage

```
data "aws_iam_policy_document" "ssm_lifecycle_trust" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["events.amazonaws.com"]
    }
  }
}

data "aws_iam_policy_document" "ssm_lifecycle" {
  statement {
    effect    = "Allow"
    actions   = ["ssm:SendCommand"]
    resources = ["arn:aws:ec2:eu-west-1:1234567890:instance/*"]

    condition {
      test     = "StringEquals"
      variable = "ec2:ResourceTag/Terminate"
      values   = ["*"]
    }
  }

  statement {
    effect    = "Allow"
    actions   = ["ssm:SendCommand"]
    resources = ["${aws_ssm_document.stop_instance.arn}"]
  }
}

resource "aws_iam_role" "ssm_lifecycle" {
  name               = "SSMLifecycle"
  assume_role_policy = "${data.aws_iam_policy_document.ssm_lifecycle_trust.json}"
}

resource "aws_iam_policy" "ssm_lifecycle" {
  name   = "SSMLifecycle"
  policy = "${data.aws_iam_policy_document.ssm_lifecycle.json}"
}

resource "aws_ssm_document" "stop_instance" {
  name          = "stop_instance"
  document_type = "Command"

  content = <<DOC
  {
    "schemaVersion": "1.2",
    "description": "Stop an instance",
    "parameters": {

    },
    "runtimeConfig": {
      "aws:runShellScript": {
        "properties": [
          {
            "id": "0.aws:runShellScript",
            "runCommand": ["halt"]
          }
        ]
      }
    }
  }
DOC
}

resource "aws_cloudwatch_event_rule" "stop_instances" {
  name                = "StopInstance"
  description         = "Stop instances nightly"
  schedule_expression = "cron(0 0 * * ? *)"
}

resource "aws_cloudwatch_event_target" "stop_instances" {
  target_id = "StopInstance"
  arn       = "${aws_ssm_document.stop_instance.arn}"
  rule      = "${aws_cloudwatch_event_rule.stop_instances.name}"
  role_arn  = "${aws_iam_role.ssm_lifecycle.arn}"

  run_command_targets {
    key    = "tag:Terminate"
    values = ["midnight"]
  }
}

```

## Example RunCommand Usage

```

resource "aws_cloudwatch_event_rule" "stop_instances" {
  name                = "StopInstance"
  description         = "Stop instances nightly"
  schedule_expression = "cron(0 0 * * ? *)"
}

resource "aws_cloudwatch_event_target" "stop_instances" {
  target_id = "StopInstance"
  arn       = "arn:aws:ssm:${var.aws_region}::document/AWS-RunShellScript"
  input     = "{\"commands\":[\"halt\"]}"
  rule      = "${aws_cloudwatch_event_rule.stop_instances.name}"
  role_arn  = "${aws_iam_role.ssm_lifecycle.arn}"

  run_command_targets {
    key    = "tag:Terminate"
    values = ["midnight"]
  }
}

```

## Example ECS Run Task with Role and Task Override Usage

```
resource "aws_iam_role" "ecs_events" {
  name = "ecs_events"
  assume_role_policy = <<DOC
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "events.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
DOC
}

resource "aws_iam_role_policy" "ecs_events_run_task_with_any_role" {
  name = "ecs_events_run_task_with_any_role"
  role = "${aws_iam_role.ecs_events.id}"
  policy = <<DOC
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": "iam:PassRole",
            "Resource": "*"
        },
        {
            "Effect": "Allow",
            "Action": "ecs:RunTask",
            "Resource": "${replace(aws_ecs_task_definition.task_name.arn, "/:\\d+$/", ":*")}"
        }
    ]
}
DOC
}

resource "aws_cloudwatch_event_target" "ecs_scheduled_task" {
  target_id = "run-scheduled-task-every-hour"
  arn       = "${aws_ecs_cluster.cluster_name.arn}"
  rule      = "${aws_cloudwatch_event_rule.every_hour.name}"
  role_arn  = "${aws_iam_role.ecs_events.arn}"

  ecs_target = {
    task_count = 1
    task_definition_arn = "${aws_ecs_task_definition.task_name.arn}"
  }

  input = <<DOC
{
  "containerOverrides": [
    {
      "name": "name-of-container-to-override",
      "command": ["bin/console", "scheduled-task"]
    }
  ]
}
DOC
}
```

## Argument Reference

-> **Note:** `input` and `input_path` are mutually exclusive options.

-> **Note:** In order to be able to have your AWS Lambda function or
   SNS topic invoked by a CloudWatch Events rule, you must setup the right permissions
   using [`aws_lambda_permission`](https://www.terraform.io/docs/providers/aws/r/lambda_permission.html)
   or [`aws_sns_topic.policy`](https://www.terraform.io/docs/providers/aws/r/sns_topic.html#policy).
   More info [here](https://docs.aws.amazon.com/AmazonCloudWatch/latest/events/resource-based-policies-cwe.html).

The following arguments are supported:

* `rule` - (Required) The name of the rule you want to add targets to.
* `target_id` - (Optional) The unique target assignment ID.  If missing, will generate a random, unique id.
* `arn` - (Required) The Amazon Resource Name (ARN) associated of the target.
* `input` - (Optional) Valid JSON text passed to the target.
* `input_path` - (Optional) The value of the [JSONPath](http://goessner.net/articles/JsonPath/)
	that is used for extracting part of the matched event when passing it to the target.
* `role_arn` - (Optional) The Amazon Resource Name (ARN) of the IAM role to be used for this target when the rule is triggered. Required if `ecs_target` is used.
* `run_command_targets` - (Optional) Parameters used when you are using the rule to invoke Amazon EC2 Run Command. Documented below. A maximum of 5 are allowed.
* `ecs_target` - (Optional) Parameters used when you are using the rule to invoke Amazon ECS Task. Documented below. A maximum of 1 are allowed.
* `batch_target` - (Optional) Parameters used when you are using the rule to invoke an Amazon Batch Job. Documented below. A maximum of 1 are allowed.
* `kinesis_target` - (Optional) Parameters used when you are using the rule to invoke an Amazon Kinesis Stream. Documented below. A maximum of 1 are allowed.
* `sqs_target` - (Optional) Parameters used when you are using the rule to invoke an Amazon SQS Queue. Documented below. A maximum of 1 are allowed.
* `input_transformer` - (Optional) Parameters used when you are providing a custom input to a target based on certain event data.

`run_command_targets` support the following:

* `key` - (Required) Can be either `tag:tag-key` or `InstanceIds`.
* `values` - (Required) If Key is `tag:tag-key`, Values is a list of tag values. If Key is `InstanceIds`, Values is a list of Amazon EC2 instance IDs.

`ecs_target` support the following:

* `group` - (Optional) Specifies an ECS task group for the task. The maximum length is 255 characters.
* `launch_type` - (Optional) Specifies the launch type on which your task is running. The launch type that you specify here must match one of the launch type (compatibilities) of the target task. Valid values are EC2 or FARGATE.
* `network_configuration` - (Optional) Use this if the ECS task uses the awsvpc network mode. This specifies the VPC subnets and security groups associated with the task, and whether a public IP address is to be used. Required if launch_type is FARGATE because the awsvpc mode is required for Fargate tasks.
* `platform_version` - (Optional) Specifies the platform version for the task. Specify only the numeric portion of the platform version, such as 1.1.0. This is used only if LaunchType is FARGATE. For more information about valid platform versions, see [AWS Fargate Platform Versions](http://docs.aws.amazon.com/AmazonECS/latest/developerguide/platform_versions.html).
* `task_count` - (Optional) The number of tasks to create based on the TaskDefinition. The default is 1.
* `task_definition_arn` - (Required) The ARN of the task definition to use if the event target is an Amazon ECS cluster.

`network_configuration` support the following:

* `subnets` - (Required) The subnets associated with the task or service.
* `security_groups` - (Optional) The security groups associated with the task or service. If you do not specify a security group, the default security group for the VPC is used.
* `assign_public_ip` - (Optional) Assign a public IP address to the ENI (Fargate launch type only). Valid values are `true` or `false`. Default `false`.

For more information, see [Task Networking](https://docs.aws.amazon.com/AmazonECS/latest/developerguide/task-networking.html)

`batch_target` support the following:

* `job_definition` - (Required) The ARN or name of the job definition to use if the event target is an AWS Batch job. This job definition must already exist.
* `job_name` - (Required) The name to use for this execution of the job, if the target is an AWS Batch job.
* `array_size` - (Optional) The size of the array, if this is an array batch job. Valid values are integers between 2 and 10,000.
* `job_attempts` - (Optional) The number of times to attempt to retry, if the job fails. Valid values are 1 to 10.

`kinesis_target` support the following:

* `partition_key_path` - (Optional) The JSON path to be extracted from the event and used as the partition key.

`sqs_target` support the following:

* `message_group_id` - (Optional) The FIFO message group ID to use as the target.

`input_transformer` support the following:

* `input_paths` - (Optional) Key value pairs specified in the form of JSONPath (for example, time = $.time)
* `input_template` - (Required) Structure containing the template body.
