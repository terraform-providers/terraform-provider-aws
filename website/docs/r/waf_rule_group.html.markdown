---
layout: "aws"
page_title: "AWS: waf_rule_group"
sidebar_current: "docs-aws-resource-waf-rule-group"
description: |-
  Provides a AWS WAF rule group resource.
---

# aws_waf_rule_group

Provides a WAF Rule Group Resource

## Example Usage

```hcl
resource "aws_waf_rule" "example" {
  name        = "example"
  metric_name = "example"
}

resource "aws_waf_rule_group" "example" {
  name        = "example"
  metric_name = "example"

  activated_rule {
    action {
      type = "COUNT"
    }

    priority = 50
    rule_id  = "${aws_waf_rule.example.id}"
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) A friendly name of the rule group
* `metric_name` - (Required) A friendly name for the metrics from the rule group
* `activated_rule` - (Optional) A list of activated rules, see below

## Nested Blocks

### `activated_rule`

#### Arguments

* `action` - (Required) Specifies the action that CloudFront or AWS WAF takes when a web request matches the conditions in the rule.
  * `type` - (Required) e.g. `BLOCK`, `ALLOW`, or `COUNT`
* `priority` - (Required) Specifies the order in which the rules are evaluated. Rules with a lower value are evaluated before rules with a higher value.
* `rule_id` - (Required) The ID of a [rule](/docs/providers/aws/r/waf_rule.html)
* `type` - (Optional) The rule type, either [`REGULAR`](/docs/providers/aws/r/waf_rule.html), [`RATE_BASED`]((/docs/providers/aws/r/waf_rate_based_rule.html), or `GROUP`. Defaults to `REGULAR`.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the WAF rule group.
