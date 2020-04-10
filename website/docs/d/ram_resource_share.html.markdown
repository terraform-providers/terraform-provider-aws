---
subcategory: "RAM"
layout: "aws"
page_title: "AWS: aws_ram_resource_share"
description: |-
  Retrieve information about a RAM Resource Share
---

# Data Source: aws_ram_resource_share

`aws_ram_resource_share` Retrieve information about a RAM Resource Share.

## Example Usage

```hcl
data "aws_ram_resource_share" "example" {
  name           = "example"
  resource_owner = "SELF"
}
```

## Search by filters

```hcl
data "aws_ram_resource_share" "tag_filter" {
  name           = "MyResourceName"
  resource_owner = "SELF"

  filter {
    name   = "NameOfTag"
    values = ["exampleNameTagValue"]
  }
}
```

## Argument Reference

The following Arguments are supported

* `name` - (Required) The name of the resource share to retrieve.
* `resource_owner` (Required) The owner of the resource share. Valid values are SELF or OTHER-ACCOUNTS

* `filter` - (Optional) A filter used to scope the list e.g. by tags. See [related docs] (https://docs.aws.amazon.com/ram/latest/APIReference/API_TagFilter.html).
  * `name` - (Required) The name of the tag key to filter on.
  * `values` - (Required) The value of the tag key.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The Amazon Resource Name (ARN) of the resource share.
* `id` - The Amazon Resource Name (ARN) of the resource share.
* `status` - The Status of the RAM share.
* `tags` - The Tags attached to the RAM share
