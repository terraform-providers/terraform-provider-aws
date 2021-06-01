---
subcategory: "Route53 Resolver"
layout: "aws"
page_title: "AWS: aws_route53_resolver_rule_association"
description: |-
  Provides a Route53 Resolver rule association.
---

# Resource: aws_route53_resolver_rule_association

Provides a Route53 Resolver rule association.

## Example Usage

```hcl
resource "aws_route53_resolver_rule_association" "example" {
  resolver_rule_id = aws_route53_resolver_rule.sys.id
  vpc_id           = aws_vpc.foo.id
}
```

## Argument Reference

The following arguments are supported:

* `resolver_rule_id` - (Required) The ID of the resolver rule that you want to associate with the VPC.
* `vpc_id` - (Required) The ID of the VPC that you want to associate the resolver rule with.
* `name` - (Optional) A name for the association that you're creating between a resolver rule and a VPC.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the resolver rule association.

## Import

Route53 Resolver rule associations can be imported using the `id`, e.g.

```
$ terraform import aws_route53_resolver_rule_association.example rslvr-rrassoc-97242eaf88example
```
