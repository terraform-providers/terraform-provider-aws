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

```terraform
resource "aws_securityhub_account" "example" {}
```

## Argument Reference

* `enable_default_standards` - (Optional) Whether to enable the security standards that Security Hub has designated as automatically enabled including: ` AWS Foundational Security Best Practices v1.0.0` and `CIS AWS Foundations Benchmark v1.2.0`. Defaults to `true`.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - AWS Account ID.

## Import

An existing Security Hub enabled account can be imported using the AWS account ID, e.g.

```
$ terraform import aws_securityhub_account.example 123456789012
```
