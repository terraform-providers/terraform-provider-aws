---
subcategory: "SES"
layout: "aws"
page_title: "AWS: aws_ses_receipt_rule_set"
description: |-
  Provides an SES receipt rule set resource
---

# Resource: aws_ses_receipt_rule_set

Provides an SES receipt rule set resource

## Example Usage

```hcl
resource "aws_ses_receipt_rule_set" "main" {
  rule_set_name = "primary-rules"
}
```

## Argument Reference

The following arguments are supported:

* `rule_set_name` - (Required) The name of the rule set

## Import

SES receipt rule sets can be imported using the rule set name.

```
$ terraform import aws_ses_receipt_rule_set.my_rule_set my_rule_set_name
```
