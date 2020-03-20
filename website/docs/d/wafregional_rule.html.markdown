---
subcategory: "WAF Regional"
layout: "aws"
page_title: "AWS: aws_wafregional_rule"
description: |-
  Retrieves an AWS WAF Regional rule id.
---

# Data Source: aws_wafregional_rule

`aws_wafregional_rule` Retrieves a WAF Regional Rule Resource Id.

## Example Usage

```hcl
data "aws_wafregional_rule" "example" {
  name = "tfWAFRegionalRule"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the WAF Regional rule.

## Attributes Reference
In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the WAF Regional rule.
