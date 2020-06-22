---
subcategory: "WorkSpaces"
layout: "aws"
page_title: "AWS: aws_workspaces_bundle"
description: |-
  Get information on a WorkSpaces Bundle.
---

# Data Source: aws_workspaces_bundle

Use this data source to get information about a WorkSpaces Bundle.

## Example Usage

```hcl
data "aws_workspaces_bundle" "example" {
  bundle_id = "wsb-b0s22j3d7"
}

data "aws_workspaces_bundle" "example" {
  owner = "AMAZON"
  name  = "Value with Windows 10 and Office 2016"
}
```

## Argument Reference

The following arguments are supported:

* `bundle_id` – (Optional) The ID of the bundle.
* `owner` – (Optional) The owner of the bundles. You have to leave it blank for own bundles. You cannot combine this parameter with `bundle_id`.
* `name` – (Optional) The name of the bundle. You cannot combine this parameter with `bundle_id`.

## Attributes Reference

The following attributes are exported:

* `description` – The description of the bundle.
* `bundle_id` – The ID of the bundle.
* `name` – The name of the bundle.
* `owner` – The owner of the bundle.
* `compute_type` – The compute type. See supported fields below.
* `root_storage` – The root volume. See supported fields below.
* `user_storage` – The user storage. See supported fields below.

### `compute_type`

* `name` - The name of the compute type.

### `root_storage`

* `capacity` - The size of the root volume.

### `user_storage`

* `capacity` - The size of the user storage.
