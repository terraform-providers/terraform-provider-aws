---
subcategory: "AppMesh"
layout: "aws"
page_title: "AWS: aws_appmesh_mesh"
description: |-
  Provides an AWS App Mesh service mesh resource.
---

# Resource: aws_appmesh_mesh

Provides an AWS App Mesh service mesh resource.

## Example Usage

### Basic

```hcl
resource "aws_appmesh_mesh" "simple" {
  name = "simpleapp"
}
```

### Egress Filter

```hcl
resource "aws_appmesh_mesh" "simple" {
  name = "simpleapp"

  spec {
    egress_filter {
      type = "ALLOW_ALL"
    }
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name to use for the service mesh.
* `spec` - (Optional) The service mesh specification to apply.
* `tags` - (Optional) A map of tags to assign to the resource.

The `spec` object supports the following:

* `egress_filter`- (Optional) The egress filter rules for the service mesh.

The `egress_filter` object supports the following:

* `type` - (Optional) The egress filter type. By default, the type is `DROP_ALL`.
Valid values are `ALLOW_ALL` and `DROP_ALL`.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the service mesh.
* `arn` - The ARN of the service mesh.
* `created_date` - The creation date of the service mesh.
* `last_updated_date` - The last update date of the service mesh.
* `mesh_owner` - The AWS account ID of the service mesh's owner.
* `resource_owner` - The resource owner's AWS account ID.

## Import

App Mesh service meshes can be imported using the `name`, e.g.

```
$ terraform import aws_appmesh_mesh.simple simpleapp
```
