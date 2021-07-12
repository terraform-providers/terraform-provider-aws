---
subcategory: "Location Service"
layout: "aws"
page_title: "AWS: aws_location_tracker"
description: |-
  Provides a Location Tracker
---

# Resource: aws_location_tracker

Provides a Location Tracker.

## Example Usage

```terraform
resource "aws_location_tracker" "default" {
  name         = "sample_location_tracker"
  description  = "Sample Location Tracker"
  pricing_plan = "RequestBasedUsage"

  tags = {
    Name = "sample_location_tracker"
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name for the tracker resource.
* `pricing_plan` - (Required) Specifies the pricing plan for the tracker resource. Valid values are `RequestBasedUsage`, `MobileAssetTracking` and `MobileAssetManagement`.
* `pricing_plan_data_source` - (Required only when `pricing_plan` is `MobileAssetTracking` or `MobileAssetManagement`) Specifies the data provider for the tracker resource. Valid values are `Esri` and `Here`.
* `description` - (Optional) An optional description for the tracker resource.
* `kms_key_id` - (Optional) The ARN of the KMS key to use when encrypting location data.
* `tags` - (Optional) A map of tags to assign to the resource.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the tracker resource.
* `arn` - The ARN of the tracker resource.

## Import

Location tracker can be imported using the `id`, e.g.

```
$ terraform import aws_location_tracker.test sample_location_tracker
```
