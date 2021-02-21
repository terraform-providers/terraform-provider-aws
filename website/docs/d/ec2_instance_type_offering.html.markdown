---
subcategory: "EC2"
layout: "aws"
page_title: "AWS: aws_ec2_instance_type_offering"
description: |-
  Information about single EC2 Instance Type Offering.
---

# Data Source: aws_ec2_instance_type_offering

Information about single EC2 Instance Type Offering.

## Example Usage

```hcl
data "aws_ec2_instance_type_offering" "example" {
  filter {
    name   = "instance-type"
    values = ["t2.micro", "t3.micro"]
  }

  preferred_instance_types = ["t3.micro", "t2.micro"]
}
```

## Argument Reference

The following arguments are supported:

* `filter` - (Optional) One or more configuration blocks containing name-values filters. See the [EC2 API Reference](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DescribeInstanceTypeOfferings.html) for supported filters. Detailed below.
* `location_type` - (Optional) Location type. Defaults to `region`. Valid values: `availability-zone`, `availability-zone-id`, and `region`.
* `preferred_instance_types` - (Optional) Ordered list of preferred EC2 Instance Types. The first match in this list will be returned. If no preferred matches are found and the original search returned more than one result, an error is returned.

### filter Argument Reference

* `name` - (Required) Name of the filter. The `location` filter depends on the top-level `location_type` argument and if not specified, defaults to the current region.
* `values` - (Required) List of one or more values for the filter.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - EC2 Instance Type.
* `instance_type` - EC2 Instance Type.
