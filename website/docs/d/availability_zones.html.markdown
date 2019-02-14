---
layout: "aws"
page_title: "AWS: aws_availability_zones"
sidebar_current: "docs-aws-datasource-availability-zones"
description: |-
    Provides a list of Availability Zones which can be used by an AWS account.
---

# Data Source: aws_availability_zones

The Availability Zones data source allows access to the list of AWS
Availability Zones which can be accessed by an AWS account within the region
configured in the provider.

This is different from the `aws_availability_zone` (singular) data source,
which provides some details about a specific availability zone.

## Example Usage

```hcl
# Declare the data source
data "aws_availability_zones" "available" {}

# e.g. Create subnets in the first two available availability zones

resource "aws_subnet" "primary" {
  availability_zone = "${data.aws_availability_zones.available.names[0]}"

  # ...
}

resource "aws_subnet" "secondary" {
  availability_zone = "${data.aws_availability_zones.available.names[1]}"

  # ...
}
```

## Argument Reference

The following arguments are supported:

* `state` - (Optional) Allows to filter list of Availability Zones based on their
current state. Can be either `"available"`, `"information"`, `"impaired"` or
`"unavailable"`. By default the list includes a complete set of Availability Zones
to which the underlying AWS account has access, regardless of their state.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `names` - A list of the Availability Zone names available to the account.
* `zone_ids` - A list of the Availability Zone IDs available to the account.

Note that the indexes of Availability Zone names and IDs correspond.
