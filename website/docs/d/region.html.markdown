---
layout: "aws"
page_title: "AWS: aws_region"
sidebar_current: "docs-aws-datasource-region"
description: |-
    Provides details about a specific service region
---

# Data Source: aws_region

`aws_region` provides details about a specific AWS region.

As well as validating a given region name this resource can be used to
discover the name of the region configured within the provider. The latter
can be useful in a child module which is inheriting an AWS provider
configuration from its parent module.

## Example Usage

The following example shows how the resource might be used to obtain
the name of the AWS region configured on the provider.

```hcl
data "aws_region" "current" {}
```

## Argument Reference

The arguments of this data source act as filters for querying the available
regions. The given filters must match exactly one region whose data will be
exported as attributes.

* `name` - (Optional) The full name of the region to select.

* `endpoint` - (Optional) The EC2 endpoint of the region to select.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `name` - The name of the selected region.

* `current` - `true` if the selected region is the one configured on the
  provider, or `false` otherwise.

* `endpoint` - The EC2 endpoint for the selected region.

* `description` - The region's description in this format: "Location (Region name)".
