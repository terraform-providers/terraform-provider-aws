---
subcategory: "GuardDuty"
layout: "aws"
page_title: "AWS: aws_guardduty_organization_admin_account"
description: |-
  Manages a GuardDuty Organization Admin Account
---

# Resource: aws_guardduty_organization_admin_account

Manages a GuardDuty Organization Admin Account. The AWS account utilizing this resource must be an Organizations primary account. More information about Organizations support in GuardDuty can be found in the [GuardDuty User Guide](https://docs.aws.amazon.com/guardduty/latest/ug/guardduty_organizations.html).

## Example Usage

```hcl
resource "aws_organizations_organization" "example" {
  aws_service_access_principals = ["guardduty.amazonaws.com"]
  feature_set                   = "ALL"
}

resource "aws_guardduty_detector" "example" {}

resource "aws_guardduty_organization_admin_account" "example" {
  depends_on = [aws_organizations_organization.example]

  admin_account_id = "123456789012"
}
```

## Argument Reference

The following arguments are supported:

* `admin_account_id` - (Required) AWS account identifier to designate as a delegated administrator for GuardDuty.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - AWS account identifier.

## Import

GuardDuty Organization Admin Account can be imported using the AWS account ID, e.g.

```
$ terraform import aws_guardduty_organization_admin_account.example 123456789012
```
