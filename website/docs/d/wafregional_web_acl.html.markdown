---
subcategory: "WAF Regional"
layout: "aws"
page_title: "AWS: aws_wafregional_web_acl"
description: |-
  Retrieves a WAF Regional Web ACL id.
---

# Data Source: aws_wafregional_web_acl

`aws_wafregional_web_acl` Retrieves a WAF Regional Web ACL Resource Id.

## Example Usage

```hcl
data "aws_wafregional_web_acl" "example" {
  name = "tfWAFRegionalWebACL"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the WAF Regional Web ACL.

## Attributes Reference
In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the WAF Regional Web ACL.
