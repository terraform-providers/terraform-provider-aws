---
subcategory: "Security Hub"
layout: "aws"
page_title: "AWS: aws_securityhub_member"
description: |-
  Provides a Security Hub member resource.
---

# Resource: aws_securityhub_member

Provides a Security Hub member resource.

## Example Usage

```hcl
resource "aws_securityhub_account" "example" {}

resource "aws_securityhub_member" "example" {
  depends_on = [aws_securityhub_account.example]
  account_id = "123456789012"
  email      = "example@example.com"
  invite     = true
}
```

## Argument Reference

The following arguments are supported:

* `account_id` - (Required) The ID of the member AWS account.
* `email` - (Required) The email of the member AWS account.
* `invite` - (Optional) Boolean whether to invite the account to Security Hub as a member. Defaults to `false`.

## Attributes Reference

The following attributes are exported in addition to the arguments listed above:

* `id` - The ID of the member AWS account (matches `account_id`).
* `master_id` - The ID of the master Security Hub AWS account.
* `member_status` - The status of the member account relationship.

## Import

Security Hub members can be imported using their account ID, e.g.

```
$ terraform import aws_securityhub_member.example 123456789012
```
