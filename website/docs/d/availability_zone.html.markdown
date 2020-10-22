---
subcategory: ""
layout: "aws"
page_title: "AWS: aws_availability_zone"
description: |-
    Provides details about a specific availability zone
---

# Data Source: aws_availability_zone

`aws_availability_zone` provides details about a specific availability zone (AZ)
in the current region.

This can be used both to validate an availability zone given in a variable
and to split the AZ name into its component parts of an AWS region and an
AZ identifier letter. The latter may be useful e.g. for implementing a
consistent subnet numbering scheme across several regions by mapping both
the region and the subnet letter to network numbers.

This is different from the `aws_availability_zones` (plural) data source,
which provides a list of the available zones.

## Example Usage

The following example shows how this data source might be used to derive
VPC and subnet CIDR prefixes systematically for an availability zone.

```hcl
variable "region_number" {
  # Arbitrary mapping of region name to number to use in
  # a VPC's CIDR prefix.
  default = {
    us-east-1      = 1
    us-west-1      = 2
    us-west-2      = 3
    eu-central-1   = 4
    ap-northeast-1 = 5
  }
}

variable "az_number" {
  # Assign a number to each AZ letter used in our configuration
  default = {
    a = 1
    b = 2
    c = 3
    d = 4
    e = 5
    f = 6
  }
}

# Retrieve the AZ where we want to create network resources
# This must be in the region selected on the AWS provider.
data "aws_availability_zone" "example" {
  name = "eu-central-1a"
}

# Create a VPC for the region associated with the AZ
resource "aws_vpc" "example" {
  cidr_block = cidrsubnet("10.0.0.0/8", 4, var.region_number[data.aws_availability_zone.example.region])
}

# Create a subnet for the AZ within the regional VPC
resource "aws_subnet" "example" {
  vpc_id     = aws_vpc.example.id
  cidr_block = cidrsubnet(aws_vpc.example.cidr_block, 4, var.az_number[data.aws_availability_zone.example.name_suffix])
}
```

## Argument Reference

The arguments of this data source act as filters for querying the available
availability zones. The given filters must match exactly one availability
zone whose data will be exported as attributes.

* `all_availability_zones` - (Optional) Set to `true` to include all Availability Zones and Local Zones regardless of your opt in status.
* `filter` - (Optional) Configuration block(s) for filtering. Detailed below.
* `name` - (Optional) The full name of the availability zone to select.
* `state` - (Optional) A specific availability zone state to require. May be any of `"available"`, `"information"` or `"impaired"`.
* `zone_id` - (Optional) The zone ID of the availability zone to select.

### filter Configuration Block

The following arguments are supported by the `filter` configuration block:

* `name` - (Required) The name of the filter field. Valid values can be found in the [EC2 DescribeAvailabilityZones API Reference](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DescribeAvailabilityZones.html).
* `values` - (Required) Set of values that are accepted for the given filter field. Results will be selected if any given value matches.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `group_name` - For Availability Zones, this is the same value as the Region name. For Local Zones, the name of the associated group, for example `us-west-2-lax-1`.
* `name_suffix` - The part of the AZ name that appears after the region name, uniquely identifying the AZ within its region.
* `network_border_group` - The name of the location from which the address is advertised.
* `opt_in_status` - For Availability Zones, this always has the value of `opt-in-not-required`. For Local Zones, this is the opt in status. The possible values are `opted-in` and `not-opted-in`.
* `region` - The region where the selected availability zone resides. This is always the region selected on the provider, since this data source searches only within that region.
