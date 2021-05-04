---
subcategory: "CloudWatch"
layout: "aws"
page_title: "AWS: aws_cloudwatch_event_rules"
description: |-
  Get Cloudwatch Event Rules.
---

# Data Source: aws_cloudwatch_event_rules

Use this data source to get the names of Cloudwatch Event Rules

## Example Usage

```hcl
data "aws_cloudwatch_event_rules" "example" {
  name_prefix = "MyEventRule"
}
```

## Argument Reference

The following arguments are supported:

* `name_prefix` - (Required) The prefix matching the rule name.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `rule_names` - A list of rule names.
