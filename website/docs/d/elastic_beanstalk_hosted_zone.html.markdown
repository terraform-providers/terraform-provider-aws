---
subcategory: "Elastic Beanstalk"
layout: "aws"
page_title: "AWS: aws_elastic_beanstalk_hosted_zone"
description: |-
  Get an elastic beanstalk hosted zone.
---

# Data Source: aws_elastic_beanstalk_hosted_zone

Use this data source to get the ID of an [elastic beanstalk hosted zone](http://docs.aws.amazon.com/general/latest/gr/rande.html#elasticbeanstalk_region).

## Example Usage

```hcl
data "aws_elastic_beanstalk_hosted_zone" "current" {}
```

## Argument Reference

* `region` - (Optional) The region you'd like the zone for. By default, fetches the current region.

## Attributes Reference

* `id` - The ID of the hosted zone.

* `region` - The region of the hosted zone.
