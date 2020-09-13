---
subcategory: "GuardDuty"
layout: "aws"
page_title: "AWS: aws_guardduty_detector"
description: |-
  Provides a resource to manage a GuardDuty detector
---

# Resource: aws_guardduty_detector

Provides a resource to manage a GuardDuty detector.

~> **NOTE:** Deleting this resource is equivalent to "disabling" GuardDuty for an AWS region, which removes all existing findings. You can set the `enable` attribute to `false` to instead "suspend" monitoring and feedback reporting while keeping existing data. See the [Suspending or Disabling Amazon GuardDuty documentation](https://docs.aws.amazon.com/guardduty/latest/ug/guardduty_suspend-disable.html) for more information.

## Example Usage

```hcl
resource "aws_guardduty_detector" "MyDetector" {
  enable = true
}
```

## Argument Reference

The following arguments are supported:

* `enable` - (Optional) Enable monitoring and feedback reporting. Setting to `false` is equivalent to "suspending" GuardDuty. Defaults to `true`.
* `finding_publishing_frequency` - (Optional) Specifies the frequency of notifications sent for subsequent finding occurrences. If the detector is a GuardDuty member account, the value is determined by the GuardDuty primary account and cannot be modified, otherwise defaults to `SIX_HOURS`. For standalone and GuardDuty primary accounts, it must be configured in Terraform to enable drift detection. Valid values for standalone and primary accounts: `FIFTEEN_MINUTES`, `ONE_HOUR`, `SIX_HOURS`. See [AWS Documentation](https://docs.aws.amazon.com/guardduty/latest/ug/guardduty_findings_cloudwatch.html#guardduty_findings_cloudwatch_notification_frequency) for more information.
* `tags` - (Optional) Key-value map of resource tags.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `account_id` - The AWS account ID of the GuardDuty detector
* `arn` - Amazon Resource Name (ARN) of the GuardDuty detector
* `id` - The ID of the GuardDuty detector

## Import

GuardDuty detectors can be imported using the detector ID, e.g.

```
$ terraform import aws_guardduty_detector.MyDetector 00b00fd5aecc0ab60a708659477e9617
```
