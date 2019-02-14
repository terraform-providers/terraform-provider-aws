---
layout: "aws"
page_title: "AWS: aws_securityhub_standards_subscription"
sidebar_current: "docs-aws-resource-securityhub-standards-subscription"
description: |-
  Subscribes to a Security Hub standard.
---

# aws_securityhub_standards_subscription

Subscribes to a Security Hub standard.

## Example Usage

```hcl
resource "aws_securityhub_account" "example" {}

resource "aws_securityhub_standards_subscription" "example" {
  depends_on    = ["aws_securityhub_account.example"]
  standards_arn = "arn:aws:securityhub:::ruleset/cis-aws-foundations-benchmark/v/1.2.0"
}
```

## Argument Reference

The following arguments are supported:

* `standards_arn` - (Required) The ARN of a standard - see below.

Currently available standards:

| Name                | ARN                                                                   |
|---------------------|-----------------------------------------------------------------------|
| CIS AWS Foundations | `arn:aws:securityhub:::ruleset/cis-aws-foundations-benchmark/v/1.2.0` |

## Attributes Reference

The following attributes are exported in addition to the arguments listed above:

* `id` - The ARN of a resource that represents your subscription to a supported standard.

## Import

Security Hub standards subscriptions can be imported using the standards subscription ARN, e.g.

```
$ terraform import aws_securityhub_standards_subscription.example arn:aws:securityhub:eu-west-1:123456789012:subscription/cis-aws-foundations-benchmark/v/1.2.0
```
