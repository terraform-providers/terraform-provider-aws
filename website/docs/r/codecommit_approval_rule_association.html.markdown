---
subcategory: "CodeCommit"
layout: "aws"
page_title: "AWS: aws_codecommit_approval_rule_association"
description: |-
  Provides a CodeCommit Approval Rule Association Resource.
---

# Resource: aws_codecommit_approval_rule

Provides a CodeCommit Approval Rule Association Resource.

~> **NOTE on CodeCommit Availability**: CodeCommit is not yet rolled out
in all regions - available regions are listed
[the AWS Docs](https://docs.aws.amazon.com/general/latest/gr/rande.html#codecommit_region).

## Example Usage

```hcl
resource "aws_codecommit_approval_rule_association" "example" {
  template_name    = "my-approval-rule"
  repository_names = ["repo1", "repo2"]
}
```

## Argument Reference

The following arguments are supported:

* `template_name` - (Required) The name of the approval rule template to associate with CodeCommit repositories
* `repository_names` - (Required) A list of CodeCommit repositories to associate the approval rule template with
