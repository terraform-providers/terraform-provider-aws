---
subcategory: "AppMesh"
layout: "aws"
page_title: "AWS: aws_appmesh_virtual_router"
description: |-
  Provides an AWS App Mesh virtual router resource.
---

# Resource: aws_appmesh_virtual_router

Provides an AWS App Mesh virtual router resource.

## Breaking Changes

Because of backward incompatible API changes (read [here](https://github.com/awslabs/aws-app-mesh-examples/issues/92) and [here](https://github.com/awslabs/aws-app-mesh-examples/issues/94)), `aws_appmesh_virtual_router` resource definitions created with provider versions earlier than v2.3.0 will need to be modified:

* Remove service `service_names` from the `spec` argument.
AWS has created a `aws_appmesh_virtual_service` resource for each of service names.
These resource can be imported using `terraform import`.

* Add a `listener` configuration block to the `spec` argument.

The Terraform state associated with existing resources will automatically be migrated.

## Example Usage

```hcl
resource "aws_appmesh_virtual_router" "serviceb" {
  name      = "serviceB"
  mesh_name = aws_appmesh_mesh.simple.id

  spec {
    listener {
      port_mapping {
        port     = 8080
        protocol = "http"
      }
    }
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name to use for the virtual router. Must be between 1 and 255 characters in length.
* `mesh_name` - (Required) The name of the service mesh in which to create the virtual router. Must be between 1 and 255 characters in length.
* `mesh_owner` - (Optional) The AWS account ID of the service mesh's owner. Defaults to the account ID the [AWS provider][1] is currently connected to.
* `spec` - (Required) The virtual router specification to apply.
* `tags` - (Optional) A map of tags to assign to the resource.

The `spec` object supports the following:

* `listener` - (Required) The listeners that the virtual router is expected to receive inbound traffic from.
Currently only one listener is supported per virtual router.

The `listener` object supports the following:

* `port_mapping` - (Required) The port mapping information for the listener.

The `port_mapping` object supports the following:

* `port` - (Required) The port used for the port mapping.
* `protocol` - (Required) The protocol used for the port mapping. Valid values are `http`,`http2`, `tcp` and `grpc`.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the virtual router.
* `arn` - The ARN of the virtual router.
* `created_date` - The creation date of the virtual router.
* `last_updated_date` - The last update date of the virtual router.
* `resource_owner` - The resource owner's AWS account ID.

## Import

App Mesh virtual routers can be imported using `mesh_name` together with the virtual router's `name`,
e.g.

```
$ terraform import aws_appmesh_virtual_router.serviceb simpleapp/serviceB
```

[1]: /docs/providers/aws/index.html
