---
subcategory: "WAF"
layout: "aws"
page_title: "AWS: aws_waf_rule"
description: |-
  Retrieves an AWS WAF rule id.
---

# Data Source: aws_waf_rule

`aws_waf_rule` Retrieves a WAF Rule Resource Id.

## Example Usage

```hcl
data "aws_waf_rule" "example" {
  name = "tfWAFRule"
}

```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the WAF rule.

## Attributes Reference
In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the WAF rule.
