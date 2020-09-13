---
subcategory: "WAF Regional"
layout: "aws"
page_title: "AWS: aws_wafregional_web_acl"
description: |-
  Provides a AWS WAF Regional web access control group (ACL) resource for use with ALB.
---

# Resource: aws_wafregional_web_acl

Provides a WAF Regional Web ACL Resource for use with Application Load Balancer.

## Example Usage

### Regular Rule

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
    data_id = aws_wafregional_ipset.ipset.id
    negated = false
    type    = "IPMatch"
  }
}

resource "aws_wafregional_web_acl" "wafacl" {
  name        = "tfWebACL"
  metric_name = "tfWebACL"

  default_action {
    type = "ALLOW"
  }

  rule {
    action {
      type = "BLOCK"
    }

    priority = 1
    rule_id  = aws_wafregional_rule.wafrule.id
    type     = "REGULAR"
  }
}
```

### Group Rule

```hcl
resource "aws_wafregional_web_acl" "example" {
  name        = "example"
  metric_name = "example"

  default_action {
    type = "ALLOW"
  }

  rule {
    priority = 1
    rule_id  = aws_wafregional_rule_group.example.id
    type     = "GROUP"

    override_action {
      type = "NONE"
    }
  }
}
```

### Logging

~> *NOTE:* The Kinesis Firehose Delivery Stream name must begin with `aws-waf-logs-`. See the [AWS WAF Developer Guide](https://docs.aws.amazon.com/waf/latest/developerguide/logging.html) for more information about enabling WAF logging.

```hcl
resource "aws_wafregional_web_acl" "example" {
  # ... other configuration ...

  logging_configuration {
    log_destination = aws_kinesis_firehose_delivery_stream.example.arn

    redacted_fields {
      field_to_match {
        type = "URI"
      }

      field_to_match {
        data = "referer"
        type = "HEADER"
      }
    }
  }
}
```

## Argument Reference

The following arguments are supported:

* `default_action` - (Required) The action that you want AWS WAF Regional to take when a request doesn't match the criteria in any of the rules that are associated with the web ACL.
* `metric_name` - (Required) The name or description for the Amazon CloudWatch metric of this web ACL.
* `name` - (Required) The name or description of the web ACL.
* `logging_configuration` - (Optional) Configuration block to enable WAF logging. Detailed below.
* `rule` - (Optional) Set of configuration blocks containing rules for the web ACL. Detailed below.
* `tags` - (Optional) Key-value map of resource tags

### `default_action` Configuration Block

* `type` - (Required) Specifies how you want AWS WAF Regional to respond to requests that match the settings in a rule. e.g. `ALLOW`, `BLOCK` or `COUNT`

### `logging_configuration` Configuration Block

* `log_destination` - (Required) Amazon Resource Name (ARN) of Kinesis Firehose Delivery Stream
* `redacted_fields` - (Optional) Configuration block containing parts of the request that you want redacted from the logs. Detailed below.

#### `redacted_fields` Configuration Block

* `field_to_match` - (Required) Set of configuration blocks for fields to redact. Detailed below.

##### `field_to_match` Configuration Block

-> Additional information about this configuration can be found in the [AWS WAF Regional API Reference](https://docs.aws.amazon.com/waf/latest/APIReference/API_regional_FieldToMatch.html).

* `data` - (Optional) When the value of `type` is `HEADER`, enter the name of the header that you want the WAF to search, for example, `User-Agent` or `Referer`. If the value of `type` is any other value, omit `data`.
* `type` - (Required) The part of the web request that you want AWS WAF to search for a specified string. e.g. `HEADER` or `METHOD`

### `rule` Configuration Block

-> Additional information about this configuration can be found in the [AWS WAF Regional API Reference](https://docs.aws.amazon.com/waf/latest/APIReference/API_regional_ActivatedRule.html).

* `priority` - (Required) Specifies the order in which the rules in a WebACL are evaluated.
  Rules with a lower value are evaluated before rules with a higher value.
* `rule_id` - (Required) ID of the associated WAF (Regional) rule (e.g. [`aws_wafregional_rule`](/docs/providers/aws/r/wafregional_rule.html)). WAF (Global) rules cannot be used.
* `action` - (Optional) Configuration block of the action that CloudFront or AWS WAF takes when a web request matches the conditions in the rule.  Not used if `type` is `GROUP`. Detailed below.
* `override_action` - (Optional) Configuration block of the override the action that a group requests CloudFront or AWS WAF takes when a web request matches the conditions in the rule.  Only used if `type` is `GROUP`. Detailed below.
* `type` - (Optional) The rule type, either `REGULAR`, as defined by [Rule](http://docs.aws.amazon.com/waf/latest/APIReference/API_Rule.html), `RATE_BASED`, as defined by [RateBasedRule](http://docs.aws.amazon.com/waf/latest/APIReference/API_RateBasedRule.html), or `GROUP`, as defined by [RuleGroup](https://docs.aws.amazon.com/waf/latest/APIReference/API_RuleGroup.html). The default is REGULAR. If you add a RATE_BASED rule, you need to set `type` as `RATE_BASED`. If you add a GROUP rule, you need to set `type` as `GROUP`.

#### `action` / `override_action` Configuration Block

* `type` - (Required) Specifies how you want AWS WAF Regional to respond to requests that match the settings in a rule. e.g. `ALLOW`, `BLOCK` or `COUNT`

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - Amazon Resource Name (ARN) of the WAF Regional WebACL.
* `id` - The ID of the WAF Regional WebACL.

## Import

WAF Regional Web ACL can be imported using the id, e.g.

```
$ terraform import aws_wafregional_web_acl.wafacl a1b2c3d4-d5f6-7777-8888-9999aaaabbbbcccc
```
