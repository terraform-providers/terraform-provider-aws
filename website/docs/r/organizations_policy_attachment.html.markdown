---
layout: "aws"
page_title: "AWS: aws_organizations_policy_attachment"
sidebar_current: "docs-aws-resource-organizations-policy-attachment"
description: |-
  Provides a resource to attach an AWS Organizations policy to an organization account, root, or unit.
---

# aws_organizations_policy_attachment

Provides a resource to attach an AWS Organizations policy to an organization account, root, or unit.

## Example Usage

### Organization Account

```hcl
resource "aws_organizations_policy_attachment" "account" {
  policy_id = "${aws_organizations_policy.example.id}"
  target_id = "123456789012"
}
```

### Organization Root

```hcl
resource "aws_organizations_policy_attachment" "root" {
  policy_id = "${aws_organizations_policy.example.id}"
  target_id = "r-12345678"
}
```

### Organization Unit

```hcl
resource "aws_organizations_policy_attachment" "unit" {
  policy_id = "${aws_organizations_policy.example.id}"
  target_id = "ou-12345678"
}
```

## Argument Reference

The following arguments are supported:

* `policy_id` - (Required) The unique identifier (ID) of the policy that you want to attach to the target.
* `target_id` - (Required) The unique identifier (ID) of the root, organizational unit, or account number that you want to attach the policy to.

## Import

`aws_organizations_policy_attachment` can be imported by using the target ID and policy ID, e.g. with an account target

```
$ terraform import aws_organizations_policy_attachment.account 123456789012:p-12345678
```
