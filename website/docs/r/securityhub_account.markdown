---
subcategory: "Security Hub"
layout: "aws"
page_title: "AWS: aws_securityhub_account"
description: |-
  Enables Security Hub for an AWS account.
---

# Resource: aws_securityhub_account

Enables Security Hub for this AWS account.

~> **NOTE:** Destroying this resource will disable Security Hub for this AWS account.

## Example Usage

```hcl
resource "aws_securityhub_account" "example" {}
```

## Argument Reference

The resource does not support any arguments.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - AWS Account ID.

## Import

An existing Security Hub enabled account can be imported using the AWS account ID, e.g.

```
$ terraform import aws_securityhub_account.example 123456789012
```
