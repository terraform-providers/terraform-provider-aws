---
subcategory: "Security Hub"
layout: "aws"
page_title: "AWS: aws_securityhub_standards_subscription"
description: |-
  Subscribes to a Security Hub standard.
---

# Resource: aws_securityhub_standards_subscription

Subscribes to a Security Hub standard.

## Example Usage

```hcl
resource "aws_securityhub_account" "example" {}

resource "aws_securityhub_standards_subscription" "cis" {
  depends_on    = [aws_securityhub_account.example]
  standards_arn = "arn:aws:securityhub:::ruleset/cis-aws-foundations-benchmark/v/1.2.0"
}

resource "aws_securityhub_standards_subscription" "pci_321" {
  depends_on    = [aws_securityhub_account.example]
  standards_arn = "arn:aws:securityhub:us-east-1::standards/pci-dss/v/3.2.1"
}
```

## Argument Reference

The following arguments are supported:

* `standards_arn` - (Required) The ARN of a standard - see below.

Currently available standards:

| Name                                     | ARN                                                                                         |
|------------------------------------------|---------------------------------------------------------------------------------------------|
| AWS Foundational Security Best Practices | `arn:aws:securityhub:us-east-1::standards/aws-foundational-security-best-practices/v/1.0.0` |
| CIS AWS Foundations                      | `arn:aws:securityhub:::ruleset/cis-aws-foundations-benchmark/v/1.2.0`                       |
| PCI DSS                                  | `arn:aws:securityhub:us-east-1::standards/pci-dss/v/3.2.1`                                  |

## Attributes Reference

The following attributes are exported in addition to the arguments listed above:

* `id` - The ARN of a resource that represents your subscription to a supported standard.

## Import

Security Hub standards subscriptions can be imported using the standards subscription ARN, e.g.

```
$ terraform import aws_securityhub_standards_subscription.cis arn:aws:securityhub:eu-west-1:123456789012:subscription/cis-aws-foundations-benchmark/v/1.2.0
$ terraform import aws_securityhub_standards_subscription.pci_321 arn:aws:securityhub:eu-west-1:123456789012:subscription/pci-dss/v/3.2.1
```
