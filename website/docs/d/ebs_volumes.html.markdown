---
subcategory: "VPC"
layout: "aws"
page_title: "AWS: aws_ebs_volumes"
description: |-
    Provides identifying information for EBS volumes matching given criteria
---

# Data Source: aws_ebs_volumes

`aws_ebs_volumes` provides identifying information for EBS volumes matching given criteria.

This data source can be useful for getting a list of volume IDs with (for example) matching tags.

## Example Usage

The following demonstrates obtaining a map of availability zone to EBS volume ID for volumes with a given tag value.

```hcl
data "aws_ebs_volumes" "example" {
  tags = {
    VolumeSet = "TestVolumeSet"
  }
}

data "aws_ebs_volume" "example" {
  for_each = data.aws_ebs_volumes.example.ids
  filter {
    name   = "volume-id"
    values = [each.value]
  }
}

output "availability_zone_to_volume_id" {
  value = { for s in data.aws_ebs_volume.example : s.id => s.availability_zone }
}
```

## Argument Reference

* `filter` - (Optional) Custom filter block as described below.

* `tags` - (Optional) A map of tags, each pair of which must exactly match
  a pair on the desired volumes.

More complex filters can be expressed using one or more `filter` sub-blocks,
which take the following arguments:

* `name` - (Required) The name of the field to filter by, as defined by
  [the underlying AWS API](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DescribeVolumes.html).
  For example, if matching against the `size` filter, use:

```hcl
data "aws_ebs_volumes" "ten_or_twenty_gb_volumes" {
  filter {
    name   = "size"
    values = ["10", "20"]
  }
}
```

* `values` - (Required) Set of values that are accepted for the given field.
  EBS Volume IDs will be selected if any one of the given values match.

## Attributes Reference

* `id` - AWS Region.
* `ids` - A set of all the EBS Volume IDs found. This data source will fail if
  no volumes match the provided criteria.
