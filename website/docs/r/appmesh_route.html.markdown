---
subcategory: "AppMesh"
layout: "aws"
page_title: "AWS: aws_appmesh_route"
description: |-
  Provides an AWS App Mesh route resource.
---

# Resource: aws_appmesh_route

Provides an AWS App Mesh route resource.

## Example Usage

### HTTP Routing

```hcl
resource "aws_appmesh_route" "serviceb" {
  name                = "serviceB-route"
  mesh_name           = "${aws_appmesh_mesh.simple.id}"
  virtual_router_name = "${aws_appmesh_virtual_router.serviceb.name}"

  spec {
    http_route {
      match {
        prefix = "/"
      }

      action {
        weighted_target {
          virtual_node = "${aws_appmesh_virtual_node.serviceb1.name}"
          weight       = 90
        }

        weighted_target {
          virtual_node = "${aws_appmesh_virtual_node.serviceb2.name}"
          weight       = 10
        }
      }
    }
  }
}
```

### HTTP Header Routing

```hcl
resource "aws_appmesh_route" "serviceb" {
  name                = "serviceB-route"
  mesh_name           = "${aws_appmesh_mesh.simple.id}"
  virtual_router_name = "${aws_appmesh_virtual_router.serviceb.name}"

  spec {
    http_route {
      match {
        method = "POST"
        prefix = "/"
        scheme = "https"

        header {
          name = "clientRequestId"

          match {
            prefix = "123"
          }
        }
      }

      action {
        weighted_target {
          virtual_node = "${aws_appmesh_virtual_node.serviceb.name}"
          weight       = 100
        }
      }
    }
  }
}
```

### TCP Routing

```hcl
resource "aws_appmesh_route" "serviceb" {
  name                = "serviceB-route"
  mesh_name           = "${aws_appmesh_mesh.simple.id}"
  virtual_router_name = "${aws_appmesh_virtual_router.serviceb.name}"

  spec {
    tcp_route {
      action {
        weighted_target {
          virtual_node = "${aws_appmesh_virtual_node.serviceb1.name}"
          weight       = 100
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
* `virtual_router_name` - (Required) The name of the virtual router in which to create the route.
* `spec` - (Required) The route specification to apply.
* `tags` - (Optional) A mapping of tags to assign to the resource.

The `spec` object supports the following:

* `http_route` - (Optional) The HTTP routing information for the route.
* `priority` - (Optional) The priority for the route, between `0` and `1000`.
Routes are matched based on the specified value, where `0` is the highest priority.
* `tcp_route` - (Optional) The TCP routing information for the route.

The `http_route` object supports the following:

* `action` - (Required) The action to take if a match is determined.
* `match` - (Required) The criteria for determining an HTTP request match.

The `tcp_route` object supports the following:

* `action` - (Required) The action to take if a match is determined.

The `action` object supports the following:

* `weighted_target` - (Required) The targets that traffic is routed to when a request matches the route.
You can specify one or more targets and their relative weights with which to distribute traffic.

The `http_route`'s `match` object supports the following:

* `prefix` - (Required) Specifies the path with which to match requests.
This parameter must always start with /, which by itself matches all requests to the virtual router service name.
* `header` - (Optional) The client request headers to match on.
* `method` - (Optional) The client request header method to match on. Valid values: `GET`, `HEAD`, `POST`, `PUT`, `DELETE`, `CONNECT`, `OPTIONS`, `TRACE`, `PATCH`.
* `scheme` - (Optional) The client request header scheme to match on. Valid values: `http`, `https`.

The `weighted_target` object supports the following:

* `virtual_node` - (Required) The virtual node to associate with the weighted target.
* `weight` - (Required) The relative weight of the weighted target. An integer between 0 and 100.

The `header` object supports the following:

* `name` - (Required) A name for the HTTP header in the client request that will be matched on.
* `invert` - (Optional) If `true`, the match is on the opposite of the `match` method and value. Default is `false`.
* `match` - (Optional) The method and value to match the header value sent with a request. Specify one match method.

The `header`'s `match` object supports the following:

* `exact` - (Optional) The header value sent by the client must match the specified value exactly.
* `prefix` - (Optional) The header value sent by the client must begin with the specified characters.
* `range`- (Optional) The object that specifies the range of numbers that the header value sent by the client must be included in.
* `regex` - (Optional) The header value sent by the client must include the specified characters.
* `suffix` - (Optional) The header value sent by the client must end with the specified characters.

The `range` object supports the following:

* `end` - (Required) The end of the range.
* `start` - (Requited) The start of the range.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the route.
* `arn` - The ARN of the route.
* `created_date` - The creation date of the route.
* `last_updated_date` - The last update date of the route.

## Import

App Mesh virtual routes can be imported using `mesh_name` and `virtual_router_name` together with the route's `name`,
e.g.

```
$ terraform import aws_appmesh_virtual_route.serviceb simpleapp/serviceB/serviceB-route
```
