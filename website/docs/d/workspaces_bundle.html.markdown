---
layout: "aws"
page_title: "AWS: aws_workspaces_bundle"
sidebar_current: "docs-aws-datasource-workspaces-bundle"
description: |-
  Get information on a Workspaces Bundle.
---

# Data Source: aws_workspaces_bundle

Use this data source to get information about a Workspaces Bundle.

## Example Usage

```hcl
data "aws_workspaces_bundle" "example" {
  bundle_id = "wsb-b0s22j3d7"
}
```

## Argument Reference

The following arguments are supported:

* `bundle_id` – (Required) The ID of the bundle.

## Attributes Reference

The following attributes are exported:

* `description` – The description of the bundle.
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
