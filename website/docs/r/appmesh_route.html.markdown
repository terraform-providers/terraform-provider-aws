---
layout: "aws"
page_title: "AWS: aws_appmesh_route"
sidebar_current: "docs-aws-resource-appmesh-route"
description: |-
  Provides an AWS App Mesh route resource.
---

# aws_appmesh_route

Provides an AWS App Mesh route resource.

## Example Usage

```hcl
resource "aws_appmesh_route" "serviceb" {
  name                = "serviceB-route"
  mesh_name           = "simpleapp"
  virtual_router_name = "serviceB"

  spec {
    http_route {
      match {
        prefix = "/"
      }

      action {
        weighted_target {
          virtual_node = "serviceBv1"
          weight       = 90
        }

        weighted_target {
          virtual_node = "serviceBv2"
          weight       = 10
        }
      }
    }
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name to use for the route.
* `mesh_name` - (Required) The name of the service mesh in which to create the route.
* `virtual_route_name` - (Required) The name of the virtual router in which to create the route.
* `spec` - (Required) The route specification to apply.

The `spec` object supports the following:

* `http_route` - (Optional) The HTTP routing information for the route.

The `http_route` object supports the following:

* `action` - (Required) The action to take if a match is determined.
* `match` - (Required) The criteria for determining an HTTP request match.

The `action` object supports the following:

* `weighted_target` - (Required) The targets that traffic is routed to when a request matches the route.
You can specify one or more targets and their relative weights with which to distribute traffic.

The `match` object supports the following:

* `prefix` - (Required) Specifies the path with which to match requests.
This parameter must always start with /, which by itself matches all requests to the virtual router service name.

The `weighted_target` object supports the following:

* `virtual_node` - (Required) The virtual node to associate with the weighted target.
* `weight` - (Required) The relative weight of the weighted target. An integer between 0 and 100.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the route.
* `arn` - The ARN of the route.
* `created_date` - The creation date of the route.
* `last_updated_date` - The last update date of the route.
