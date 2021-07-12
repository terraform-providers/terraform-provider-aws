---
subcategory: "CloudWatch"
layout: "aws"
page_title: "AWS: aws_cloudwatch_log_groups"
description: |-
  Get list of Cloudwatch Log Groups.
---

# Data Source: aws_cloudwatch_log_groups

Use this data source to get a list of AWS Cloudwatch Log Groups

## Example Usage

```hcl
data "aws_cloudwatch_log_groups" "example" {
  log_group_prefix = "/MyImportantLogs"
}
```

## Argument Reference

The following arguments are supported:

* `log_group_name_prefix` - (Required) The group prefix of the Cloudwatch log groups to list

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arns` - The ARNs of the Cloudwatch log groups
* `log_group_names` - The names of the Cloudwatch log groups
