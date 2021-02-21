---
subcategory: "SES"
layout: "aws"
page_title: "AWS: aws_ses_active_receipt_rule_set"
description: |-
  Provides a resource to designate the active SES receipt rule set
---

# Resource: aws_ses_active_receipt_rule_set

Provides a resource to designate the active SES receipt rule set

## Example Usage

```hcl
resource "aws_ses_active_receipt_rule_set" "main" {
  rule_set_name = "primary-rules"
}
```

## Argument Reference

The following arguments are supported:

* `rule_set_name` - (Required) The name of the rule set

## Attributes Reference

In addition to the arguments, which are exported, the following attributes are exported:

* `id` - The SES receipt rule set name.
* `arn` - The SES receipt rule set ARN.
