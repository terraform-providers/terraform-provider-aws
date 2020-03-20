---
subcategory: "Route53 Resolver"
layout: "aws"
page_title: "AWS: aws_route53_resolver_rules"
description: |-
    Provides details about a set of Route53 Resolver rules
---

# Data Source: aws_route53_resolver_rules

`aws_route53_resolver_rules` provides details about a set of Route53 Resolver rules.

## Example Usage

The following example shows how to get Route53 Resolver rules based on tags.

```hcl
data "aws_route53_resolver_rules" "example" {
  tags = {
    Environment = "dev"
  }
}
```

## Argument Reference

The arguments of this data source act as filters for querying the available resolver rules in the current region.

* `owner_id` (Optional) When the desired resolver rules are shared with another AWS account, the account ID of the account that the rules are shared with.
* `resolver_endpoint_id` (Optional) The ID of the outbound resolver endpoint for the desired resolver rules.
* `rule_type` (Optional) The rule type of the desired resolver rules. Valid values are `FORWARD`, `SYSTEM` and `RECURSIVE`.
* `share_status` (Optional) Whether the desired resolver rules are shared and, if so, whether the current account is sharing the rules with another account, or another account is sharing the rules with the current account.
Values are `NOT_SHARED`, `SHARED_BY_ME` or `SHARED_WITH_ME`

## Attributes Reference

The following attributes are exported:

* `resolver_rule_ids` - The IDs of the matched resolver rules.
