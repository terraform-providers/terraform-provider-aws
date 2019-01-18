---
layout: "aws"
page_title: "AWS: aws_appmesh_virtual_node"
sidebar_current: "docs-aws-resource-appmesh-virtual-node"
description: |-
  Provides an AWS App Mesh virtual node resource.
---

# aws_appmesh_virtual_node

Provides an AWS App Mesh virtual node resource.

## Example Usage

```hcl
resource "aws_appmesh_virtual_node" "serviceb1" {
  name                = "serviceBv1"
  mesh_name           = "simpleapp"

  spec {
    backends = ["servicea.simpleapp.local"]

    listener {
      port_mapping {
        port     = 8080
        protocol = "http"
      }
    }

    service_discovery {
      dns {
        service_name = "serviceb.simpleapp.local"
      }
    }
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name to use for the virtual node.
* `mesh_name` - (Required) The name of the service mesh in which to create the virtual node.
* `spec` - (Required) The virtual node specification to apply.

The `spec` object supports the following:

* `backends` - (Optional) The backends to which the virtual node is expected to send outbound traffic.
* `listener` - (Optional) The listeners from which the virtual node is expected to receive inbound traffic.
* `service_discovery`- (Optional) The service discovery information for the virtual node.

The `listener` object supports the following:

* `port_mapping` - (Required) The port mapping information for the listener.

The `service_discovery` object supports the following:

* `dns` - (Required) Specifies the DNS service name for the virtual node.

The `dns` object supports the following:

* `service_name` - (Required) The DNS service name for your virtual node.

The `port_mapping` object supports the following:

* `port` - (Required) The port used for the port mapping.
* `protocol` - (Required) The protocol used for the port mapping. Valid values are `http` and `tcp`.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the virtual node.
* `arn` - The ARN of the virtual node.
* `created_date` - The creation date of the virtual node.
* `last_updated_date` - The last update date of the virtual node.
