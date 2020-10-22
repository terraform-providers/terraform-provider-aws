---
subcategory: "AppMesh"
layout: "aws"
page_title: "AWS: aws_appmesh_gateway_route"
description: |-
  Provides an AWS App Mesh gateway route resource.
---

# Resource: aws_appmesh_gateway_route

Provides an AWS App Mesh gateway route resource.

## Example Usage

```hcl
resource "aws_appmesh_gateway_route" "example" {
  name                 = "example-gateway-route"
  mesh_name            = "example-service-mesh"
  virtual_gateway_name = aws_appmesh_virtual_gateway.example.name

  spec {
    http_route {
      action {
        target {
          virtual_service {
            virtual_service_name = aws_appmesh_virtual_service.example.name
          }
        }
      }

      match {
        prefix = "/"
      }
    }
  }

  tags = {
    Environment = "test"
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name to use for the gateway route.
* `mesh_name` - (Required) The name of the service mesh in which to create the gateway route.
* `virtual_gateway_name` - (Required) The name of the [virtual gateway](/docs/providers/aws/r/appmesh_virtual_gateway.html) to associate the gateway route with.
* `mesh_owner` - (Optional) The AWS account ID of the service mesh's owner. Defaults to the account ID the [AWS provider][1] is currently connected to.
* `spec` - (Required) The gateway route specification to apply.
* `tags` - (Optional) A map of tags to assign to the resource.

The `spec` object supports the following:

* `grpc_route` - (Optional) The specification of a gRPC gateway route.
* `http_route` - (Optional) The specification of an HTTP gateway route.
* `http2_route` - (Optional) The specification of an HTTP/2 gateway route.

The `grpc_route`, `http_route` and `http2_route` objects supports the following:

* `action` - (Required) The action to take if a match is determined.
* `match` - (Required) The criteria for determining a request match.

The `grpc_route`, `http_route` and `http2_route`'s `action` object supports the following:

* `target` - (Required) The target that traffic is routed to when a request matches the gateway route.

The `target` object supports the following:

* `virtual_service` - (Required) The virtual service gateway route target.

The `virtual_service` object supports the following:

* `virtual_service_name` - (Required) The name of the virtual service that traffic is routed to.

The `grpc_route`'s `match` object supports the following:

* `service_name` - (Required) The fully qualified domain name for the service to match from the request.

The `http_route` and `http2_route`'s `match` object supports the following:

* `prefix` - (Required) Specifies the path to match requests with. This parameter must always start with `/`, which by itself matches all requests to the virtual service name.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the gateway route.
* `arn` - The ARN of the gateway route.
* `created_date` - The creation date of the gateway route.
* `last_updated_date` - The last update date of the gateway route.
* `resource_owner` - The resource owner's AWS account ID.

## Import

App Mesh gateway routes can be imported using `mesh_name` and `virtual_gateway_name` together with the gateway route's `name`,
e.g.

```
$ terraform import aws_appmesh_gateway_route.example mesh/gw1/example-gateway-route
```

[1]: /docs/providers/aws/index.html
