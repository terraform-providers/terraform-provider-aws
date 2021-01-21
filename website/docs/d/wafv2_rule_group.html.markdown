---
subcategory: "WAFv2"
layout: "aws"
page_title: "AWS: aws_wafv2_rule_group"
description: |-
  Retrieves the summary of a WAFv2 Rule Group.
---

# Data Source: aws_wafv2_rule_group

Retrieves the summary of a WAFv2 Rule Group.

## Example Usage

```hcl
data "aws_wafv2_rule_group" "example" {
  name  = "some-rule-group"
  scope = "REGIONAL"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the WAFv2 Rule Group.
* `scope` - (Required) Specifies whether this is for an AWS CloudFront distribution or for a regional application. Valid values are `CLOUDFRONT` or `REGIONAL`. To work with CloudFront, you must also specify the region `us-east-1` (N. Virginia) on the AWS provider.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The Amazon Resource Name (ARN) of the entity.
* `description` - The description of the rule group that helps with identification.
* `id` - The unique identifier of the rule group.
