---
subcategory: "Organizations"
layout: "aws"
page_title: "AWS: aws_organizations_accounts"
description: |-
  Get all direct child organizational units under a parent organizational unit. This only provides immediate children, not all children.
---

# Data Source: aws_organizations_accounts

Get all direct child organizational units under a parent organizational unit. This only provides immediate children, not all children.

## Example Usage

```terraform
data "aws_organizations_organization" "org" {}

data "aws_organizations_accounts" "ou" {
  parent_id = data.aws_organizations_organization.org.roots[0].id
}
```

## Argument Reference

* `parent_id` - (Required) The parent ID of the accounts.

## Attributes Reference

* `children` - List of child accounts, which have the following attributes:
    * `arn` - ARN of the organizational account
    * `name` - Name of the organizational account
    * `id` - ID of the organizational account
    * `email` - `email` - Email of the account
* `id` - Parent identifier of the accounts.
