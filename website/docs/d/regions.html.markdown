---
subcategory: ""
layout: "aws"
page_title: "AWS: aws_regions"
description: |-
    Provides information about AWS Regions.
---

# Data Source: aws_regions

Provides information about AWS Regions. Can be used to filter regions i.e. by Opt-In status or only regions enabled for current account. To get details like endpoint and description of each region the data source can be combined with the [`aws_region` data source](/docs/providers/aws/d/region.html).

## Example Usage

Enabled AWS Regions:

```hcl
data "aws_regions" "current" {}
```

All the regions regardless of the availability

```hcl
data "aws_regions" "current" {
  all_regions = true
}
```

To see regions that are filtered by `"not-opted-in"`, the `all_regions` argument needs to be set to `true` or no results will be returned.

```hcl
data "aws_regions" "current" {
  all_regions = true

  filter {
    name   = "opt-in-status"
    values = ["not-opted-in"]
  }
}
```

## Argument Reference

The following arguments are supported:

* `all_regions` - (Optional) If true the source will query all regions regardless of availability.

* `filter` - (Optional) Configuration block(s) to use as filters. Detailed below.

### filter Configuration Block

The following arguments are supported by the `filter` configuration block:

* `name` - (Required) The name of the filter field. Valid values can be found in the [describe-regions AWS CLI Reference][1].
* `values` - (Required) Set of values that are accepted for the given filter field. Results will be selected if any given value matches.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - Identifier of the current partition (e.g. `aws` in AWS Commercial, `aws-cn` in AWS China).
* `names` - Names of regions that meets the criteria.

[1]: https://docs.aws.amazon.com/cli/latest/reference/ec2/describe-regions.html
