---
subcategory: "Step Function (SFN)"
layout: "aws"
page_title: "AWS: aws_sfn_activity"
description: |-
  Use this data source to get information about a Step Functions Activity.
---

# Data Source: aws_sfn_activity

Provides a Step Functions Activity data source

## Example Usage

```hcl
data "aws_sfn_activity" "sfn_activity" {
  name = "my-activity"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Optional) The name that identifies the activity.
* `arn` - (Optional) The Amazon Resource Name (ARN) that identifies the activity.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The Amazon Resource Name (ARN) that identifies the activity.
* `creation_date` - The date the activity was created.
