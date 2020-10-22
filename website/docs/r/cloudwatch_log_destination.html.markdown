---
subcategory: "CloudWatch"
layout: "aws"
page_title: "AWS: aws_cloudwatch_log_destination"
description: |-
  Provides a CloudWatch Logs destination.
---

# Resource: aws_cloudwatch_log_destination

Provides a CloudWatch Logs destination resource.

## Example Usage

```hcl
resource "aws_cloudwatch_log_destination" "test_destination" {
  name       = "test_destination"
  role_arn   = aws_iam_role.iam_for_cloudwatch.arn
  target_arn = aws_kinesis_stream.kinesis_for_cloudwatch.arn
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) A name for the log destination
* `role_arn` - (Required) The ARN of an IAM role that grants Amazon CloudWatch Logs permissions to put data into the target
* `target_arn` - (Required) The ARN of the target Amazon Kinesis stream resource for the destination

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The Amazon Resource Name (ARN) specifying the log destination.

## Import

CloudWatch Logs destinations can be imported using the `name`, e.g.

```
$ terraform import aws_cloudwatch_log_destination.test_destination test_destination
```
