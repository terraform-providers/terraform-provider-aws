---
layout: "aws"
page_title: "AWS: aws_cloudwatch_log_group"
sidebar_current: "docs-aws-cloudwatch-log-group"
description: |-
  Get information on a Cloudwatch Log Group.
---

# Data Source: aws_cloudwatch_log_group

Use this data source to get information about an AWS Cloudwatch Log Group

## Example Usage

```hcl
data "aws_cloudwatch_log_group" "example" {
  name = "MyImportantLogs"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the Cloudwatch log group

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The ARN of the Cloudwatch log group
* `creation_time` - The creation time of the log group, expressed as the number of milliseconds after Jan 1, 1970 00:00:00 UTC.
