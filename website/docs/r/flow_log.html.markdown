---
layout: "aws"
page_title: "AWS: aws_flow_log"
sidebar_current: "docs-aws-resource-flow-log"
description: |-
  Provides a VPC/Subnet/ENI Flow Log
---

# aws_flow_log

Provides a VPC/Subnet/ENI Flow Log to capture IP traffic for a specific network
interface, subnet, or VPC. Logs are sent to a CloudWatch Log Group or a S3 Bucket.

## Example Usage

### CloudWatch Logging

```hcl
resource "aws_flow_log" "example" {
  iam_role_arn    = "${aws_iam_role.example.arn}"
  log_destination = "${aws_cloudwatch_log_group.example.arn}"
  traffic_type    = "ALL"
  vpc_id          = "${aws_vpc.example.id}"
}

resource "aws_cloudwatch_log_group" "example" {
  name = "example"
}

resource "aws_iam_role" "test_role" {
  name = "example"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "vpc-flow-logs.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "example" {
  name = "example"
  role = "${aws_iam_role.example.id}"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "logs:CreateLogGroup",
        "logs:CreateLogStream",
        "logs:PutLogEvents",
        "logs:DescribeLogGroups",
        "logs:DescribeLogStreams"
      ],
      "Effect": "Allow",
      "Resource": "*"
    }
  ]
}
EOF
}
```

### S3 Logging

```hcl
resource "aws_flow_log" "example" {
  log_destination      = "${aws_s3_bucket.example.arn}"
  log_destination_type = "s3"
  traffic_type         = "ALL"
  vpc_id               = "${aws_vpc.example.id}"
}

resource "aws_s3_bucket" "example" {
  name = "example"
}
```

## Argument Reference

~> **NOTE:** One of `eni_id`, `subnet_id`, or `vpc_id` must be specified.

The following arguments are supported:

* `traffic_type` - (Required) The type of traffic to capture. Valid values: `ACCEPT`,`REJECT`, `ALL`.
* `eni_id` - (Optional) Elastic Network Interface ID to attach to
* `iam_role_arn` - (Optional) The ARN for the IAM role that's used to post flow logs to a CloudWatch Logs log group
* `log_destination_type` - (Optional) The type of the logging destination. Valid values: `cloud-watch-logs`, `s3`. Default: `cloud-watch-logs`.
* `log_destination` - (Optional) The ARN of the logging destination.
* `log_group_name` - (Optional) *Deprecated:* Use `log_destination` instead. The name of the CloudWatch log group.
* `subnet_id` - (Optional) Subnet ID to attach to
* `vpc_id` - (Optional) VPC ID to attach to

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The Flow Log ID

## Import

Flow Logs can be imported using the `id`, e.g.

```
$ terraform import aws_flow_log.test_flow_log fl-1a2b3c4d
```
