---
subcategory: "CodeCommit"
layout: "aws"
page_title: "AWS: aws_codecommit_approval_rule"
description: |-
  Provides a CodeCommit Approval Rule Resource.
---

# Resource: aws_codecommit_approval_rule

Provides a CodeCommit Approval Rule Resource.

~> **NOTE on CodeCommit Availability**: CodeCommit is not yet rolled out
in all regions - available regions are listed
[the AWS Docs](https://docs.aws.amazon.com/general/latest/gr/rande.html#codecommit_region).

## Example Usage

```hcl
resource "aws_codecommit_approval_rule" "test" {
  name        = "MyTestApprovalRule"
  description = "This is a test approval rule template"

  content = <<EOF
{
    "Version": "2018-11-08",
    "DestinationReferences": ["refs/heads/master"],
    "Statements": [{
        "Type": "Approv ers",
        "NumberOfApprovalsNeeded": 2,
        "ApprovalPoolMembers": ["arn:aws:sts::123456789012:assumed-role/CodeCommitReview/*"]}]
}
EOF
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name for the approval rule template. This needs to be less than 100 characters.
* `description` - (Optional) The description of the approval rule template. This needs to be less than 1000 characters
* `content` - (Optional) The content of the approval rule template.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `approval_rule_template_id` - The ID of the approval rule template

## Import

Codecommit approval rule templates can be imported using approval rule template name, e.g.

```
$ terraform import aws_codecommit_approval_rule.imported ExistingApprovalRuleTemplateName
```
