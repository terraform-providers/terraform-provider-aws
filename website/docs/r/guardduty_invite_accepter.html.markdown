---
subcategory: "GuardDuty"
layout: "aws"
page_title: "AWS: aws_guardduty_invite_accepter"
description: |-
  Provides a resource to accept a pending GuardDuty invite on creation, ensure the detector has the correct primary account on read, and disassociate with the primary account upon removal.
---

# Resource: aws_guardduty_invite_accepter

Provides a resource to accept a pending GuardDuty invite on creation, ensure the detector has the correct primary account on read, and disassociate with the primary account upon removal.

## Example Usage

```hcl
provider "aws" {
  alias = "primary"
}

provider "aws" {
  alias = "member"
}

resource "aws_guardduty_invite_accepter" "member" {
  depends_on = [aws_guardduty_member.member]
  provider   = aws.member

  detector_id       = aws_guardduty_detector.member.id
  master_account_id = aws_guardduty_detector.primary.account_id
}

resource "aws_guardduty_member" "member" {
  provider    = aws.primary
  account_id  = aws_guardduty_detector.member.account_id
  detector_id = aws_guardduty_detector.primary.id
  email       = "required@example.com"
  invite      = true
}

resource "aws_guardduty_detector" "primary" {
  provider = aws.primary
}

resource "aws_guardduty_detector" "member" {
  provider = aws.member
}
```

## Argument Reference

The following arguments are supported:

* `detector_id` - (Required) The detector ID of the member GuardDuty account.
* `master_account_id` - (Required) AWS account ID for primary account.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - GuardDuty member detector ID

## Timeouts

`aws_guardduty_invite_accepter` provides the following [Timeouts](https://www.terraform.io/docs/configuration/blocks/resources/syntax.html#operation-timeouts)
configuration options:

- `create` - (Default `1m`) How long to wait for an invite to accept.

## Import

`aws_guardduty_invite_accepter` can be imported using the the member GuardDuty detector ID, e.g.

```
$ terraform import aws_guardduty_invite_accepter.member 00b00fd5aecc0ab60a708659477e9617
```
