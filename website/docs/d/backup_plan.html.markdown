---
subcategory: "Backup"
layout: "aws"
page_title: "AWS: aws_backup_plan"
description: |-
  Provides details about an AWS Backup plan.
---

# Data Source: aws_backup_plan

Use this data source to get information on an existing backup plan.

## Example Usage

```hcl
data "aws_backup_plan" "example" {
  plan_id = "tf_example_backup_plan_id"
}
```

## Argument Reference

The following arguments are supported:

* `plan_id` - (Required) The backup plan ID.

## Attributes Reference

In addition to the arguments above, the following attributes are exported:

* `arn` - The ARN of the backup plan.
* `name` - The display name of a backup plan.
* `tags` - Metadata that you can assign to help organize the plans you create.
* `version` - Unique, randomly generated, Unicode, UTF-8 encoded string that serves as the version ID of the backup plan.
