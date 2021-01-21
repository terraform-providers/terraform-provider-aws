---
subcategory: "Identity Store"
layout: "aws"
page_title: "AWS: aws_identitystore_user"
description: |-
  Get information on an Identity Store User
---

# Data Source: aws_identitystore_user

Use this data source to get an Identity Store User.

## Example Usage

```hcl
data "aws_ssoadmin_instances" "example" {}

data "aws_identitystore_user" "example" {
  identity_store_id = tolist(data.aws_ssoadmin_instances.example.identity_store_ids)[0]

  filter {
    attribute_path  = "UserName"
    attribute_value = "ExampleUser"
  }
}

output "user_id" {
  value = data.aws_identitystore_user.example.user_id
}
```

## Argument Reference

The following arguments are supported:

* `filter` - (Required) Configuration block(s) for filtering. Currently, the AWS Identity Store API supports only 1 filter. Detailed below.
* `user_id` - (Optional)  The identifier for a user in the Identity Store.
* `identity_store_id` - (Required) The Identity Store ID associated with the Single Sign-On Instance.

### `filter` Configuration Block

The following arguments are supported by the `filter` configuration block:

* `attribute_path` - (Required) The attribute path that is used to specify which attribute name to search. Currently, `UserName` is the only valid attribute path.
* `attribute_value` - (Required) The value for an attribute.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The identifier of the user in the Identity Store.
* `user_name` - The user's user name value.
