---
subcategory: ""
layout: "aws"
page_title: "AWS: aws_hosted_zones"
description: |-
    Returns AWS hosted zone data for all supported services.
---

# Data Source: aws_hosted_zones

Returns AWS hosted zone data for all supported services.

## Example Usage

```terraform
data "aws_hosted_zones" "current" {}
```

## Argument Reference

The following arguments are supported:

* `region` - (Optional) The region you'd like the zones for. Uses the current region by default. Note that some hosted zone IDs are the same in all regions (Cloudfront, Global Accelerator).

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

~> **NOTE:** Some regions do not support all services. In these cases the hosted zone will be an empty string.

* `alb` - The hosted zone ID to use with Application Load Balancers.
* `cloudfront` - The hosted zone ID to use with Cloudfront. Value is the same in all regions.
* `elastic_beanstalk` - The hosted zone ID to use with Elastic Beanstalk.
* `elb` - The hosted zone ID to use with Elastic Load Balancers. Always the same value as `alb`.
* `global_accelerator` - The hosted zone ID to use with Global Accelerator. Value is the same in all regions.
* `nlb` - The hosted zone ID to use with Network Load Balancers.
* `s3` - The hosted zone ID to use with S3.
