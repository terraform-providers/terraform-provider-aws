---
layout: "aws"
page_title: "AWS: aws_organizations_account"
sidebar_current: "docs-aws-resource-organizations-account"
description: |-
  Provides a resource to create a member account in the current AWS Organization.
---

# aws_organizations_account

Provides a resource to create a member account in the current organization.

~> **Note:** Account management must be done from the organization's master account.

!> **WARNING:** Deleting this Terraform resource will only remove an AWS account from an organization. Terraform will not close the account. The member account must be prepared to be a standalone account beforehand. See the [AWS Organizations documentation](https://docs.aws.amazon.com/organizations/latest/userguide/orgs_manage_accounts_remove.html) for more information.

## Example Usage:

```hcl
resource "aws_organizations_account" "account" {
  name  = "my_new_account"
  email = "john@doe.org"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) A friendly name for the member account.
* `email` - (Required) The email address of the owner to assign to the new member account. This email address must not already be associated with another AWS account.
* `iam_user_access_to_billing` - (Optional) If set to `ALLOW`, the new account enables IAM users to access account billing information if they have the required permissions. If set to `DENY`, then only the root user of the new account can access account billing information.
* `role_name` - (Optional) The name of an IAM role that Organizations automatically preconfigures in the new member account. This role trusts the master account, allowing users in the master account to assume the role, as permitted by the master account administrator. The role has administrator permissions in the new member account.

## Attributes Reference

The following additional attributes are exported:

* `arn` - The ARN for this account.

## Import

The AWS member account can be imported by using the `account_id`, e.g.

```
$ terraform import aws_organizations_account.my_org 111111111111
```
