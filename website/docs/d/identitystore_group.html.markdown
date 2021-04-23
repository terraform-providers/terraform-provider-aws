---
subcategory: "Identity Store"
layout: "aws"
page_title: "AWS: aws_identitystore_group"
description: |-
  Get information on an Identity Store Group
---

# Data Source: aws_identitystore_group

Use this data source to get an Identity Store Group.

## Example Usage

```hcl
data "aws_ssoadmin_instances" "example" {}

data "aws_identitystore_group" "example" {
  identity_store_id = tolist(data.aws_ssoadmin_instances.example.identity_store_ids)[0]

  filter {
    attribute_path  = "DisplayName"
    attribute_value = "ExampleGroup"
  }
}

output "group_id" {
  value = data.aws_identitystore_group.example.group_id
}
```

## Argument Reference

The following arguments are supported:

* `filter` - (Required) Configuration block(s) for filtering. Currently, the AWS Identity Store API supports only 1 filter. Detailed below.
* `group_id` - (Optional)  The identifier for a group in the Identity Store.
* `identity_store_id` - (Required) The Identity Store ID associated with the Single Sign-On Instance.

### `filter` Configuration Block

The following arguments are supported by the `filter` configuration block:

* `attribute_path` - (Required) The attribute path that is used to specify which attribute name to search. Currently, `DisplayName` is the only valid attribute path.
* `attribute_value` - (Required) The value for an attribute.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The identifier of the group in the Identity Store.
* `display_name` - The group's display name value.
