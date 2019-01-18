---
layout: "aws"
page_title: "AWS: aws_appmesh_mesh"
sidebar_current: "docs-aws-resource-appmesh-mesh"
description: |-
  Provides an AWS App Mesh service mesh resource.
---

# aws_appmesh_mesh

Provides an AWS App Mesh service mesh resource.

## Example Usage

```hcl
resource "aws_appmesh_mesh" "simple" {
  name = "simpleapp"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name to use for the service mesh.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the service mesh.
* `arn` - The ARN of the service mesh.
* `created_date` - The creation date of the service mesh.
* `last_updated_date` - The last update date of the service mesh.

## Import

App Mesh service meshes can be imported using the `name`, e.g.

```
$ terraform import aws_appmesh_mesh.simple simpleapp
```
