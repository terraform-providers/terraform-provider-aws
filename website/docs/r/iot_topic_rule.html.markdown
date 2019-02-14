---
layout: "aws"
page_title: "AWS: aws_iot_topic_rule"
sidebar_current: "docs-aws-resource-iot-topic-rule"
description: |-
    Creates and manages an AWS IoT topic rule
---

# aws_iot_topic_rule

## Example Usage

```
resource "aws_iot_topic_rule" "rule" {
  name = "MyRule"
  description = "Example rule"
  enabled = true
  sql = "SELECT * FROM 'topic/test'"
  sql_version = "2015-10-08"

  sns {
    message_format = "RAW"
    role_arn = "${aws_iam_role.role.arn}"
    target_arn = "${aws_sns_topic.mytopic.arn}"
  }
}

resource "aws_sns_topic" "mytopic" {
  name = "mytopic"
}

resource "aws_iam_role" "role" {
    name = "myrole"
    assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "iot.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "iam_policy_for_lambda" {
    name = "mypolicy"
    role = "${aws_iam_role.role.id}"
    policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
        "Effect": "Allow",
        "Action": [
            "sns:Publish"
        ],
        "Resource": "${aws_sns_topic.mytopic.arn}"
    }
  ]
}
EOF
}
```

## Argument Reference

* `name` - (Required) The name of the rule.
* `description` - (Optional) The description of the rule.
* `enabled` - (Required) Specifies whether the rule is enabled.
* `sql` - (Required) The SQL statement used to query the topic. For more information, see AWS IoT SQL Reference (http://docs.aws.amazon.com/iot/latest/developerguide/iot-rules.html#aws-iot-sql-reference) in the AWS IoT Developer Guide.
* `sql_version` - (Required) The version of the SQL rules engine to use when evaluating the rule.

The `cloudwatch_alarm` object takes the following arguments:

* `alarm_name` - (Required) The CloudWatch alarm name.
* `role_arn` - (Required) The IAM role ARN that allows access to the CloudWatch alarm.
* `state_reason` - (Required) The reason for the alarm change.
* `state_value` - (Required) The value of the alarm state. Acceptable values are: OK, ALARM, INSUFFICIENT_DATA.

The `cloudwatch_metric` object takes the following arguments:

* `metric_name` - (Required) The CloudWatch metric name.
* `metric_namespace` - (Required) The CloudWatch metric namespace name.
* `metric_timestamp` - (Optional) An optional Unix timestamp (http://docs.aws.amazon.com/AmazonCloudWatch/latest/DeveloperGuide/cloudwatch_concepts.html#about_timestamp).
* `metric_unit` - (Required) The metric unit (supported units can be found here: http://docs.aws.amazon.com/AmazonCloudWatch/latest/DeveloperGuide/cloudwatch_concepts.html#Unit)
* `metric_value` - (Required) The CloudWatch metric value.
* `role_arn` - (Required) The IAM role ARN that allows access to the CloudWatch metric.

The `dynamodb` object takes the following arguments:

* `hash_key_field` - (Required) The hash key name.
* `hash_key_type` - (Optional) The hash key type. Valid values are "STRING" or "NUMBER".
* `hash_key_value` - (Required) The hash key value.
* `payload_field` - (Optional) The action payload.
* `range_key_field` - (Optional) The range key name.
* `range_key_type` - (Optional) The range key type. Valid values are "STRING" or "NUMBER".
* `range_key_value` - (Optional) The range key value.
* `role_arn` - (Required) The ARN of the IAM role that grants access to the DynamoDB table.
* `table_name` - (Required) The name of the DynamoDB table.

The `elasticsearch` object takes the following arguments:

* `endpoint` - (Required) The endpoint of your Elasticsearch domain.
* `id` - (Required) The unique identifier for the document you are storing.
* `index` - (Required) The Elasticsearch index where you want to store your data.
* `role_arn` - (Required) The IAM role ARN that has access to Elasticsearch.
* `type` - (Required) The type of document you are storing.

The `firehose` object takes the following arguments:

* `delivery_stream_name` - (Required) The delivery stream name.
* `role_arn` - (Required) The IAM role ARN that grants access to the Amazon Kinesis Firehose stream.
* `separator` - (Optional) A character separator that is used to separate records written to the Firehose stream. Valid values are: '\n' (newline), '\t' (tab), '\r\n' (Windows newline), ',' (comma).

The `kinesis` object takes the following arguments:

* `partition_key` - (Optional) The partition key.
* `role_arn` - (Required) The ARN of the IAM role that grants access to the Amazon Kinesis stream.
* `stream_name` - (Required) The name of the Amazon Kinesis stream.

The `lambda` object takes the following arguments:

* `function_arn` - (Required) The ARN of the Lambda function.

The `republish` object takes the following arguments:

* `role_arn` - (Required) The ARN of the IAM role that grants access.
* `topic` - (Required) The name of the MQTT topic the message should be republished to.

The `s3` object takes the following arguments:

* `bucket_name` - (Required) The Amazon S3 bucket name.
* `key` - (Required) The object key.
* `role_arn` - (Required) The ARN of the IAM role that grants access.

The `sns` object takes the following arguments:

* `message_format` - (Required) The message format of the message to publish. Accepted values are "JSON" and "RAW".
* `role_arn` - (Required) The ARN of the IAM role that grants access.
* `target_arn` - (Required) The ARN of the SNS topic.

The `sqs` object takes the following arguments:

* `queue_url` - (Required) The URL of the Amazon SQS queue.
* `role_arn` - (Required) The ARN of the IAM role that grants access.
* `use_base64` - (Required) Specifies whether to use Base64 encoding.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The name of the topic rule
* `arn` - The ARN of the topic rule

## Import

IoT Topic Rules can be imported using the `name`, e.g.

```
$ terraform import aws_iot_topic_rule.rule <name>
```
