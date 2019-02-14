---
layout: "aws"
page_title: "AWS: aws_securityhub_account"
sidebar_current: "docs-aws-resource-securityhub-account"
description: |-
  Enables Security Hub for an AWS account.
---

# aws_securityhub_account

-> **Note:** Destroying this resource will disable Security Hub for this AWS account.

Enables Security Hub for this AWS account.

## Example Usage

```hcl
resource "aws_securityhub_account" "example" {}
```

## Argument Reference

The resource does not support any arguments.

## Attributes Reference

The following attributes are exported in addition to the arguments listed above:

* `id` - AWS Account ID.

## Import

An existing Security Hub enabled account can be imported using the AWS account ID, e.g.

```
$ terraform import aws_securityhub_account.example 123456789012
```
