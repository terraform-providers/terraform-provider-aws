---
subcategory: "AppMesh"
layout: "aws"
page_title: "AWS: aws_appmesh_mesh"
description: |-
    Provides details about an App Mesh Mesh service mesh resource.
---

# Data Source: aws_appmesh_mesh

The App Mesh Mesh data source allows details of an App Mesh Mesh to be retrieved by its name and optionally the mesh_owner.

## Example Usage

```hcl
data "aws_appmesh_mesh" "simple" {
  name = "simpleapp"
}
```

```hcl
data "aws_caller_identity" "current" {}

data "aws_appmesh_mesh" "simple" {
  name       = "simpleapp"
  mesh_owner = data.aws_caller_identity.current.account_id
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the service mesh.
* `mesh_owner` - (Optional) The AWS account ID of the service mesh's owner.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The ARN of the service mesh.
* `created_date` - The creation date of the service mesh.
* `last_updated_date` - The last update date of the service mesh.
* `resource_owner` - The resource owner's AWS account ID.
* `spec` - The service mesh specification.
* `tags` - A map of tags.

### Spec

* `egress_filter` - The egress filter rules for the service mesh.

### Egress Filter

* `type` - The egress filter type.