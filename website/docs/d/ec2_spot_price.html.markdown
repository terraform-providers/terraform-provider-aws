---
subcategory: "EC2"
layout: "aws"
page_title: "AWS: aws_ec2_spot_price"
description: |-
  Information about most recent Spot Price for a given EC2 instance.
---

# Data Source: aws_ec2_spot_price

Information about most recent Spot Price for a given EC2 instance.

## Example Usage

```hcl
data "aws_ec2_spot_price" "example" {
  instance_type     = "t3.medium"
  availability_zone = "us-west-2a"

  filter {
    name   = "product-description"
    values = ["Linux/UNIX"]
  }
}
```

## Argument Reference

The following arguments are supported:

* `instance_type` - (Optional) The type of instance for which to query Spot Price information.
* `availability_zone` - (Optional) The availability zone in which to query Spot price information.
* `filter` - (Optional) One or more configuration blocks containing name-values filters. See the [EC2 API Reference](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DescribeSpotPriceHistory.html) for supported filters. Detailed below.

### filter Argument Reference

* `name` - (Required) Name of the filter.
* `values` - (Required) List of one or more values for the filter.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - AWS Region.
* `spot_price` - The most recent Spot Price value for the given instance type and AZ.
* `spot_price_timestamp` - The timestamp at which the Spot Price value was published.
