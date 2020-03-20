---
subcategory: "WAF Regional"
layout: "aws"
page_title: "AWS: aws_wafregional_ipset"
description: |-
  Retrieves an AWS WAF Regional IP set id.
---

# Data Source: aws_wafregional_ipset

`aws_wafregional_ipset` Retrieves a WAF Regional IP Set Resource Id.

## Example Usage

```hcl
data "aws_wafregional_ipset" "example" {
  name = "tfWAFRegionalIPSet"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the WAF Regional IP set.

## Attributes Reference
In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the WAF Regional IP set.
