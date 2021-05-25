---
subcategory: "Lake Formation"
layout: "aws"
page_title: "AWS: aws_lakeformation_policy_tag"
description: |-
    Creates a tag with the specified name and values.
---

# Resource: aws_lakeformation_policy_tag

Creates a policy tag with the specified name and values. Each key must have at least one value. The maximum number of values permitted is 15.

## Example Usage

```terraform
resource "aws_lakeformation_policy_tag" "example" {
  key    = "module"
  values = ["Orders", "Sales", "Customers"]
}
```

## Argument Reference

The following arguments are supported:

* `catalog_id` - (Optional) ID of the Data Catalog to create the tag in. If omitted, this defaults to the AWS Account ID.
* `key` - (Required) The key-name for the tag.
* `values` - (Required) A list of possible values an attribute can take.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - Catalog ID and key-name of the tag

## Import

Lake Formation Policy Tags can be imported using the `catalog_id:key`. If you have not set a Catalog ID specify the AWS Account ID that the database is in, e.g.

```
$ terraform import aws_lakeformation_policy_tag.example 123456789012:some_key
```

