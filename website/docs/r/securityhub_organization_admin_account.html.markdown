---
subcategory: "Security Hub"
layout: "aws"
page_title: "AWS: aws_securityhub_organization_admin_account"
description: |-
  Manages a Security Hub administrator account for an organization.
---

# Resource: aws_securityhub_organization_admin_account

Manages a Security Hub administrator account for an organization. The AWS account utilizing this resource must be an Organizations primary account. More information about Organizations support in Security Hub can be found in the [Security Hub User Guide](https://docs.aws.amazon.com/securityhub/latest/userguide/designate-orgs-admin-account.html).

## Example Usage

```hcl
resource "aws_organizations_organization" "example" {
  aws_service_access_principals = ["securityhub.amazonaws.com"]
  feature_set                   = "ALL"
}

resource "aws_securityhub_account" "example" {}

resource "aws_securityhub_organization_admin_account" "example" {
  depends_on = [aws_organizations_organization.example]

  admin_account_id = "123456789012"
}
```

## Argument Reference

The following arguments are supported:

* `admin_account_id` - (Required) The AWS account identifier of the account to designate as the Security Hub administrator account.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - AWS account identifier.

## Import

Security Hub Organization Admin Accounts can be imported using the AWS account ID, e.g.

```
$ terraform import aws_securityhub_organization_admin_account.example 123456789012
```
