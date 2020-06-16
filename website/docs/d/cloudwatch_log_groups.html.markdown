---
subcategory: "CloudWatch"
layout: "aws"
page_title: "AWS: aws_cloudwatch_log_groups"
description: |-
  Get information on a Cloudwatch Log Groups.
---

# Data Source: aws_cloudwatch_log_groups

Use this data source to get information about an AWS Cloudwatch Log Group

## Example Usage

```hcl
data "aws_cloudwatch_log_groups" "example" {
  prefix = "tf-logs-"
}
```

## Argument Reference

The following arguments are supported:

* `prefix` - (Required) The prefix of the Cloudwatch log group name

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `names` - A list of the Cloud Watch Log Groups Names in the current region.
* `arns` - A list of the Cloud Watch Log Groups Arns in the current region.
* `creation_times` - A list of the Cloud Watch Log Groups creation time, expressed as the number of milliseconds after Jan 1, 1970 00:00:00 UTC.
* `retention_in_days` - A list of the Cloud Watch Log Groups retention in day.
* `metric_filter_counts` - A list of the Cloud Watch Log Groups metric filter count.
* `kms_key_ids` - A list of the Cloud Watch Log Groups KMS Key ID.
