---
subcategory: "Directory Service"
layout: "aws"
page_title: "AWS: aws_directory_service_log_subscription"
description: |-
  Provides a Log subscription for AWS Directory Service that pushes logs to cloudwatch.
---

# Resource: aws_directory_service_log_subscription

Provides a Log subscription for AWS Directory Service that pushes logs to cloudwatch.

## Example Usage

```hcl
resource "aws_cloudwatch_log_group" "example" {
  name              = "/aws/directoryservice/${aws_directory_service_directory.example.id}"
  retention_in_days = 14
}

data "aws_iam_policy_document" "ad-log-policy" {
  statement {
    actions = [
      "logs:CreateLogStream",
      "logs:PutLogEvents",
    ]

    principals {
      identifiers = ["ds.amazonaws.com"]
      type        = "Service"
    }

    resources = ["${aws_cloudwatch_log_group.example.arn}:*"]

    effect = "Allow"
  }
}

resource "aws_cloudwatch_log_resource_policy" "ad-log-policy" {
  policy_document = data.aws_iam_policy_document.ad-log-policy.json
  policy_name     = "ad-log-policy"
}

resource "aws_directory_service_log_subscription" "example" {
  directory_id   = aws_directory_service_directory.example.id
  log_group_name = aws_cloudwatch_log_group.example.name
}
```

## Argument Reference

The following arguments are supported:

* `directory_id` - (Required) The id of directory.
* `log_group_name` - (Required) Name of the cloudwatch log group to which the logs should be published. The log group should be already created and the directory service principal should be provided with required permission to create stream and publish logs. Changing this value would delete the current subscription and create a new one. A directory can only have one log subscription at a time.

## Import

Directory Service Log Subscriptions can be imported using the directory id, e.g.

```
$ terraform import aws_directory_service_log_subscription.msad d-1234567890
```
