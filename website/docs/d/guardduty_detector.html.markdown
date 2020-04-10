---
subcategory: "GuardDuty"
layout: "aws"
page_title: "AWS: aws_guardduty_detector"
description: |-
  Retrieve information about a GuardDuty detector.
---

# Data Source: aws_guardduty_detector

Retrieve information about a GuardDuty detector.

## Example Usage

```hcl
data "aws_guardduty_detector" "example" {}
```

## Argument Reference

* `id` - (Optional) The ID of the detector.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `finding_publishing_frequency` - The frequency of notifications sent about subsequent finding occurrences.
* `service_role_arn` - The service-linked role that grants GuardDuty access to the resources in the AWS account.
* `status` - The current status of the detector.
