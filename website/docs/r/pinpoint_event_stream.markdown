---
layout: "aws"
page_title: "AWS: aws_pinpoint_event_stream"
sidebar_current: "docs-aws-resource-pinpoint-event-stream"
description: |-
  Provides a Pinpoint Event Stream resource.
---

# aws_pinpoint_event_stream

Provides a Pinpoint Event Stream resource.

## Example Usage

```hcl
resource "aws_pinpoint_event_stream" "stream" {
  application_id         = "${aws_pinpoint_app.app.application_id}"
  destination_stream_arn = "${aws_kinesis_stream.test_stream.arn}"
  role_arn               = "${aws_iam_role.test_role.arn}"
}

resource "aws_pinpoint_app" "app" {}

resource "aws_kinesis_stream" "test_stream" {
  name        = "pinpoint-kinesis-test"
  shard_count = 1
}

resource "aws_iam_role" "test_role" {
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "pinpoint.us-east-1.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "test_role_policy" {
  name = "test_policy"
  role = "${aws_iam_role.test_role.id}"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": {
    "Action": [
      "kinesis:PutRecords",
      "kinesis:DescribeStream"
    ],
    "Effect": "Allow",
    "Resource": [
      "arn:aws:kinesis:us-east-1:*:*/*"
    ]
  }
}
EOF
}
```


## Argument Reference

The following arguments are supported:

* `application_id` - (Required) The application ID.
* `destination_stream_arn` - (Required) The Amazon Resource Name (ARN) of the Amazon Kinesis stream or Firehose delivery stream to which you want to publish events.
* `role_arn` - (Required) The IAM role that authorizes Amazon Pinpoint to publish events to the stream in your account.

## Import

Pinpoint Event Stream can be imported using the `application-id`, e.g.

```
$ terraform import aws_pinpoint_event_stream.stream application-id
```
