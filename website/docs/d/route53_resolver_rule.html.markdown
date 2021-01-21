---
subcategory: "Route53 Resolver"
layout: "aws"
page_title: "AWS: aws_route53_resolver_rule"
description: |-
    Provides details about a specific Route53 Resolver rule
---

# Data Source: aws_route53_resolver_rule

`aws_route53_resolver_rule` provides details about a specific Route53 Resolver rule.

## Example Usage

The following example shows how to get a Route53 Resolver rule based on its associated domain name and rule type.

```hcl
data "aws_route53_resolver_rule" "example" {
  domain_name = "subdomain.example.com"
  rule_type   = "SYSTEM"
}
```

## Argument Reference

The arguments of this data source act as filters for querying the available resolver rules in the current region.
The given filters must match exactly one resolver rule whose data will be exported as attributes.

* `domain_name` - (Optional) The domain name the desired resolver rule forwards DNS queries for. Conflicts with `resolver_rule_id`.
* `name` - (Optional) The friendly name of the desired resolver rule. Conflicts with `resolver_rule_id`.
* `resolver_endpoint_id` (Optional) The ID of the outbound resolver endpoint of the desired resolver rule. Conflicts with `resolver_rule_id`.
* `resolver_rule_id` (Optional) The ID of the desired resolver rule. Conflicts with `domain_name`, `name`, `resolver_endpoint_id` and `rule_type`.
* `rule_type` - (Optional) The rule type of the desired resolver rule. Valid values are `FORWARD`, `SYSTEM` and `RECURSIVE`. Conflicts with `resolver_rule_id`.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the resolver rule.
* `arn` - The ARN (Amazon Resource Name) for the resolver rule.
* `owner_id` - When a rule is shared with another AWS account, the account ID of the account that the rule is shared with.
* `share_status` - Whether the rules is shared and, if so, whether the current account is sharing the rule with another account, or another account is sharing the rule with the current account.
Values are `NOT_SHARED`, `SHARED_BY_ME` or `SHARED_WITH_ME`
* `tags` - A map of tags assigned to the resolver rule.
