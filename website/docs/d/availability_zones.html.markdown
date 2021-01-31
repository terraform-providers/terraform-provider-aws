---
subcategory: ""
layout: "aws"
page_title: "AWS: aws_availability_zones"
description: |-
    Provides a list of Availability Zones which can be used by an AWS account.
---

# Data Source: aws_availability_zones

The Availability Zones data source allows access to the list of AWS
Availability Zones which can be accessed by an AWS account within the region
configured in the provider.

This is different from the `aws_availability_zone` (singular) data source,
which provides some details about a specific availability zone.

-> When [Local Zones](https://aws.amazon.com/about-aws/global-infrastructure/localzones/) are enabled in a region, by default the API and this data source include both Local Zones and Availability Zones. To return only Availability Zones, see the example section below.

## Example Usage

### By State

```hcl
# Declare the data source
data "aws_availability_zones" "available" {
  state = "available"
}

# e.g. Create subnets in the first two available availability zones

resource "aws_subnet" "primary" {
  availability_zone = data.aws_availability_zones.available.names[0]

  # ...
}

resource "aws_subnet" "secondary" {
  availability_zone = data.aws_availability_zones.available.names[1]

  # ...
}
```

### By Filter

All Local Zones (regardless of opt-in status):

```hcl
data "aws_availability_zones" "example" {
  all_availability_zones = true

  filter {
    name   = "opt-in-status"
    values = ["not-opted-in", "opted-in"]
  }
}
```

Only Availability Zones (no Local Zones):

```
data "aws_availability_zones" "example" {
  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}
```

## Argument Reference

The following arguments are supported:

* `all_availability_zones` - (Optional) Set to `true` to include all Availability Zones and Local Zones regardless of your opt in status.
* `filter` - (Optional) Configuration block(s) for filtering. Detailed below.
* `exclude_names` - (Optional) List of Availability Zone names to exclude.
* `exclude_zone_ids` - (Optional) List of Availability Zone IDs to exclude.
* `state` - (Optional) Allows to filter list of Availability Zones based on their
current state. Can be either `"available"`, `"information"`, `"impaired"` or
`"unavailable"`. By default the list includes a complete set of Availability Zones
to which the underlying AWS account has access, regardless of their state.

### filter Configuration Block

The following arguments are supported by the `filter` configuration block:

* `name` - (Required) The name of the filter field. Valid values can be found in the [EC2 DescribeAvailabilityZones API Reference](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DescribeAvailabilityZones.html).
* `values` - (Required) Set of values that are accepted for the given filter field. Results will be selected if any given value matches.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `group_names` A set of the Availability Zone Group names. For Availability Zones, this is the same value as the Region name. For Local Zones, the name of the associated group, for example `us-west-2-lax-1`.
* `id` - Region of the Availability Zones.
* `names` - A list of the Availability Zone names available to the account.
* `zone_ids` - A list of the Availability Zone IDs available to the account.

Note that the indexes of Availability Zone names and IDs correspond.
