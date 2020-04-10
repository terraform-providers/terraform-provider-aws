---
subcategory: "MediaConvert"
layout: "aws"
page_title: "AWS: aws_media_convert_queue"
description: |-
  Provides an AWS Elemental MediaConvert Queue.
---

# Resource: aws_media_convert_queue

Provides an AWS Elemental MediaConvert Queue.

## Example Usage

```hcl
resource "aws_media_convert_queue" "test" {
  name = "tf-test-queue"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) A unique identifier describing the queue
* `description` - (Optional) A description of the queue
* `pricing_plan` - (Optional) Specifies whether the pricing plan for the queue is on-demand or reserved. Valid values are `ON_DEMAND` or `RESERVED`. Default to `ON_DEMAND`.
* `reservation_plan_settings` - (Optional) A detail pricing plan of the  reserved queue. See below.
* `status` - (Optional) A status of the queue. Valid values are `ACTIVE` or `RESERVED`. Default to `PAUSED`.
* `tags` - (Optional) A mapping of tags to assign to the resource.

### Nested Fields

#### `reservation_plan_settings`

* `commitment` - (Required) The length of the term of your reserved queue pricing plan commitment. Valid value is `ONE_YEAR`.
* `renewal_type` - (Required) Specifies whether the term of your reserved queue pricing plan. Valid values are `AUTO_RENEW` or `EXPIRE`.
* `reserved_slots` - (Required) Specifies the number of reserved transcode slots (RTS) for queue.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The same as `name`
* `arn` - The Arn of the queue

## Import

Media Convert Queue can be imported via the queue name, e.g.

```
$ terraform import aws_media_convert_queue.test tf-test-queue
```
