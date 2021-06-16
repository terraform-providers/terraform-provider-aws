---
subcategory: "Macie"
layout: "aws"
page_title: "AWS: aws_macie2_organization_admin_account"
description: |-
  Provides a resource to manage an Amazon Macie Organization Admin Account.
---

# Resource: aws_macie2_organization_admin_account

Provides a resource to manage an [Amazon Macie Organization Admin Account](https://docs.aws.amazon.com/macie/latest/APIReference/admin.html).

## Example Usage

```terraform
resource "aws_macie2_account" "example" {}

resource "aws_macie2_organization_admin_account" "test" {
  admin_account_id = "ID OF THE ADMIN ACCOUNT"
  depends_on       = [aws_macie2_account.test]
}
```

## Argument Reference

The following arguments are supported:

* `admin_account_id` - (Required) The AWS account ID for the account to designate as the delegated Amazon Macie administrator account for the organization.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The unique identifier (ID) of the macie organization admin account.

## Import

`aws_macie2_organization_admin_account` can be imported using the id, e.g.

```
$ terraform import aws_macie2_organization_admin_account.example abcd1
```
