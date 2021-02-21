---
subcategory: "EC2"
layout: "aws"
page_title: "AWS: aws_ec2_instance_type_offerings"
description: |-
  Information about EC2 Instance Type Offerings.
---

# Data Source: aws_ec2_instance_type_offerings

Information about EC2 Instance Type Offerings.

## Example Usage

```hcl
data "aws_ec2_instance_type_offerings" "example" {
  filter {
    name   = "instance-type"
    values = ["t2.micro", "t3.micro"]
  }

  filter {
    name   = "location"
    values = ["usw2-az4"]
  }

  location_type = "availability-zone-id"
}
```

## Argument Reference

The following arguments are supported:

* `filter` - (Optional) One or more configuration blocks containing name-values filters. See the [EC2 API Reference](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DescribeInstanceTypeOfferings.html) for supported filters. Detailed below.
* `location_type` - (Optional) Location type. Defaults to `region`. Valid values: `availability-zone`, `availability-zone-id`, and `region`.

### filter Argument Reference

* `name` - (Required) Name of the filter. The `location` filter depends on the top-level `location_type` argument and if not specified, defaults to the current region.
* `values` - (Required) List of one or more values for the filter.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - AWS Region.
* `instance_types` - Set of EC2 Instance Types.
