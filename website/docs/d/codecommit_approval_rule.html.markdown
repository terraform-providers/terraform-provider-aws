---
subcategory: "CodeCommit"
layout: "aws"
page_title: "AWS: aws_codecommit_approval_rule"
description: |-
  Provides details about CodeCommit Approval Rule Templates.
---

# Data Source: aws_codecommit_approval_rule

The CodeCommit Approval Rule data source allows the Approval Rule Template ID, Approval Rule Name and Approval Rule Content to be retrieved for an CodeCommit Approval Rule Template.

## Example Usage

```hcl
data "aws_codecommit_approval_rule" "test" {
  name = "MyTestApprovalRule"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name for the approval rule template. This needs to be less than 100 characters.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `approval_rule_template_id` - The ID of the approval rule template.
* `description` - The description of the approval rule template.
* `content` - The content of the approval rule template.
