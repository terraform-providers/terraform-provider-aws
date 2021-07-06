---
subcategory: "CloudFront"
layout: "aws"
page_title: "AWS: aws_cloudfront_monitoring_subscription"
description: |-
  Provides a CloudFront monitoring subscription resource.
---

# Resource: aws_cloudfront_monitoring_subscription

Provides a CloudFront real-time log configuration resource.

## Example Usage

```terraform
resource "aws_cloudfront_monitoring_subscription" "example" {
  distribution_id = aws_cloudfront_distribution.example.id

  monitoring_subscription {
    realtime_metrics_subscription_config {
      realtime_metrics_subscription_status = "Enabled"
    }
  }
}
```

## Argument Reference

The following arguments are supported:

* `distribution_id` - (Required) The ID of the distribution that you are enabling metrics for.
* `monitoring_subscription` - (Required) A monitoring subscription. This structure contains information about whether additional CloudWatch metrics are enabled for a given CloudFront distribution.

### monitoring_subscription

* `realtime_metrics_subscription_config` - (Required) A subscription configuration for additional CloudWatch metrics. See below.

### realtime_metrics_subscription_config

* `realtime_metrics_subscription_status` - (Required) A flag that indicates whether additional CloudWatch metrics are enabled for a given CloudFront distribution. Valid values are `Enabled` and `Disabled`. See below.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the CloudFront monitoring subscription, which corresponds to the `distribution_id`.

## Import

CloudFront monitoring subscription can be imported using the id, e.g.

```
$ terraform import aws_cloudfront_monitoring_subscription.example E3QYSUHO4VYRGB
```
