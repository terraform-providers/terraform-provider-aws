---
subcategory: "WAF Regional"
layout: "aws"
page_title: "AWS: aws_wafregional_rule"
description: |-
  Provides an AWS WAF Regional rule resource for use with ALB.
---

# Resource: aws_wafregional_rule

Provides an WAF Regional Rule Resource for use with Application Load Balancer.

## Example Usage

```hcl
resource "aws_wafregional_ipset" "ipset" {
  name = "tfIPSet"

  ip_set_descriptor {
    type  = "IPV4"
    value = "192.0.7.0/24"
  }
}

resource "aws_wafregional_rule" "wafrule" {
  name        = "tfWAFRule"
  metric_name = "tfWAFRule"

  predicate {
    type    = "IPMatch"
    data_id = aws_wafregional_ipset.ipset.id
    negated = false
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name or description of the rule.
* `metric_name` - (Required) The name or description for the Amazon CloudWatch metric of this rule.
* `predicate` - (Optional) The objects to include in a rule (documented below).
* `tags` - (Optional) Key-value map of resource tags

## Nested Fields

### `predicate`

See the [WAF Documentation](https://docs.aws.amazon.com/waf/latest/APIReference/API_Predicate.html) for more information.

#### Arguments

* `type` - (Required) The type of predicate in a rule. Valid values: `ByteMatch`, `GeoMatch`, `IPMatch`, `RegexMatch`, `SizeConstraint`, `SqlInjectionMatch`, or `XssMatch`
* `data_id` - (Required) The unique identifier of a predicate, such as the ID of a `ByteMatchSet` or `IPSet`.
* `negated` - (Required) Whether to use the settings or the negated settings that you specified in the objects.

## Remarks

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the WAF Regional Rule.
* `arn` - The ARN of the WAF Regional Rule.

## Import

WAF Regional Rule can be imported using the id, e.g.

```
$ terraform import aws_wafregional_rule.wafrule a1b2c3d4-d5f6-7777-8888-9999aaaabbbbcccc
```
