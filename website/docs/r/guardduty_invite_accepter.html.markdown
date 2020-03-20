---
subcategory: "GuardDuty"
layout: "aws"
page_title: "AWS: aws_guardduty_invite_accepter"
description: |-
  Provides a resource to accept a pending GuardDuty invite on creation, ensure the detector has the correct master account on read, and disassociate with the master account upon removal.
---

# Resource: aws_guardduty_invite_accepter

Provides a resource to accept a pending GuardDuty invite on creation, ensure the detector has the correct master account on read, and disassociate with the master account upon removal.

## Example Usage

```hcl
resource "aws_guardduty_detector" "master" {}

resource "aws_guardduty_detector" "member" {
  provider = "aws.dev"
}

resource "aws_guardduty_member" "dev" {
  account_id  = "${aws_guardduty_detector.member.account_id}"
  detector_id = "${aws_guardduty_detector.master.id}"
  email       = "required@example.com"
  invite      = true
}

resource "aws_guardduty_invite_accepter" "member" {
  depends_on = ["aws_guardduty_member.dev"]
  provider   = "aws.dev"

  detector_id       = "${aws_guardduty_detector.member.id}"
  master_account_id = "${aws_guardduty_detector.master.account_id}"
}
```

## Argument Reference

The following arguments are supported:

* `detector_id` - (Required) The detector ID of the member GuardDuty account.
* `master_account_id` - (Required) AWS account ID for master account.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - GuardDuty member detector ID

## Timeouts

`aws_guardduty_invite_accepter` provides the following [Timeouts](/docs/configuration/resources.html#timeouts)
configuration options:

- `create` - (Default `1m`) How long to wait for an invite to accept.

## Import

`aws_guardduty_invite_accepter` can be imported using the the member GuardDuty detector ID, e.g.

```
$ terraform import aws_guardduty_invite_accepter.member 00b00fd5aecc0ab60a708659477e9617
```
