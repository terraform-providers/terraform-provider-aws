---
subcategory: "GuardDuty"
layout: "aws"
page_title: "AWS: aws_guardduty_organization_configuration"
description: |-
  Manages the GuardDuty Organization Configuration
---

# Resource: aws_guardduty_organization_configuration

Manages the GuardDuty Organization Configuration in the current AWS Region. The AWS account utilizing this resource must have been assigned as a delegated Organization administrator account, e.g. via the [`aws_guardduty_organization_admin_account` resource](/docs/providers/aws/r/guardduty_organization_admin_account.html). More information about Organizations support in GuardDuty can be found in the [GuardDuty User Guide](https://docs.aws.amazon.com/guardduty/latest/ug/guardduty_organizations.html).

~> **NOTE:** This is an advanced Terraform resource. Terraform will automatically assume management of the GuardDuty Organization Configuration without import and perform no actions on removal from the Terraform configuration.

## Example Usage

```hcl
resource "aws_guardduty_detector" "example" {
  enable = true
}

resource "aws_guardduty_organization_configuration" "example" {
  auto_enable = true
  detector_id = aws_guardduty_detector.example.id
}
```

## Argument Reference

The following arguments are supported:

* `auto_enable` - (Required) When this setting is enabled, all new accounts that are created in, or added to, the organization are added as a member accounts of the organization’s GuardDuty delegated administrator and GuardDuty is enabled in that AWS Region.
* `detector_id` - (Required) The detector ID of the GuardDuty account.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - Identifier of the GuardDuty Detector.

## Import

GuardDuty Organization Configurations can be imported using the GuardDuty Detector ID, e.g.

```
$ terraform import aws_guardduty_organization_configuration.example 00b00fd5aecc0ab60a708659477e9617
```
