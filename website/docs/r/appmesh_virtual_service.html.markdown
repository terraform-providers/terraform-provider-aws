---
subcategory: "AppMesh"
layout: "aws"
page_title: "AWS: aws_appmesh_virtual_service"
description: |-
  Provides an AWS App Mesh virtual service resource.
---

# Resource: aws_appmesh_virtual_service

Provides an AWS App Mesh virtual service resource.

## Example Usage

### Virtual Node Provider

```hcl
resource "aws_appmesh_virtual_service" "servicea" {
  name      = "servicea.simpleapp.local"
  mesh_name = aws_appmesh_mesh.simple.id

  spec {
    provider {
      virtual_node {
        virtual_node_name = aws_appmesh_virtual_node.serviceb1.name
      }
    }
  }
}
```

### Virtual Router Provider

```hcl
resource "aws_appmesh_virtual_service" "servicea" {
  name      = "servicea.simpleapp.local"
  mesh_name = aws_appmesh_mesh.simple.id

  spec {
    provider {
      virtual_router {
        virtual_router_name = aws_appmesh_virtual_router.serviceb.name
      }
    }
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name to use for the virtual service.
* `mesh_name` - (Required) The name of the service mesh in which to create the virtual service.
* `mesh_owner` - (Optional) The AWS account ID of the service mesh's owner. Defaults to the account ID the [AWS provider][1] is currently connected to.
* `spec` - (Required) The virtual service specification to apply.
* `tags` - (Optional) A map of tags to assign to the resource.

The `spec` object supports the following:

* `provider`- (Optional) The App Mesh object that is acting as the provider for a virtual service. You can specify a single virtual node or virtual router.

The `provider` object supports the following:

* `virtual_node` - (Optional) The virtual node associated with a virtual service.
* `virtual_router` - (Optional) The virtual router associated with a virtual service.

The `virtual_node` object supports the following:

* `virtual_node_name` - (Required) The name of the virtual node that is acting as a service provider.

The `virtual_router` object supports the following:

* `virtual_router_name` - (Required) The name of the virtual router that is acting as a service provider.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the virtual service.
* `arn` - The ARN of the virtual service.
* `created_date` - The creation date of the virtual service.
* `last_updated_date` - The last update date of the virtual service.
* `resource_owner` - The resource owner's AWS account ID.

## Import

App Mesh virtual services can be imported using `mesh_name` together with the virtual service's `name`,
e.g.

```
$ terraform import aws_appmesh_virtual_service.servicea simpleapp/servicea.simpleapp.local
```

[1]: /docs/providers/aws/index.html
