---
subcategory: "AppMesh"
layout: "aws"
page_title: "AWS: aws_appmesh_virtual_gateway"
description: |-
  Provides an AWS App Mesh virtual gateway resource.
---

# Resource: aws_appmesh_virtual_gateway

Provides an AWS App Mesh virtual gateway resource.

## Example Usage

### Basic

```hcl
resource "aws_appmesh_virtual_gateway" "example" {
  name      = "example-virtual-gateway"
  mesh_name = "example-service-mesh"

  spec {
    listener {
      port_mapping {
        port     = 8080
        protocol = "http"
      }
    }
  }

  tags = {
    Environment = "test"
  }
}
```

### Access Logs and TLS

```hcl
resource "aws_appmesh_virtual_gateway" "example" {
  name      = "example-virtual-gateway"
  mesh_name = "example-service-mesh"

  spec {
    listener {
      port_mapping {
        port     = 8080
        protocol = "http"
      }

      tls {
        certificate {
          acm {
            certificate_arn = aws_acm_certificate.example.arn
          }
        }

        mode = "STRICT"
      }
    }

    logging {
      access_log {
        file {
          path = "/var/log/access.log"
        }
      }
    }
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name to use for the virtual gateway. Must be between 1 and 255 characters in length.
* `mesh_name` - (Required) The name of the service mesh in which to create the virtual gateway. Must be between 1 and 255 characters in length.
* `mesh_owner` - (Optional) The AWS account ID of the service mesh's owner. Defaults to the account ID the [AWS provider][1] is currently connected to.
* `spec` - (Required) The virtual gateway specification to apply.
* `tags` - (Optional) A map of tags to assign to the resource.

The `spec` object supports the following:

* `listener` - (Required) The listeners that the mesh endpoint is expected to receive inbound traffic from. You can specify one listener.
* `backend_defaults` - (Optional) The defaults for backends.
* `logging` - (Optional) The inbound and outbound access logging information for the virtual gateway.

The `backend_defaults` object supports the following:

* `client_policy` - (Optional) The default client policy for virtual gateway backends.

The `client_policy` object supports the following:

* `tls` - (Optional) The Transport Layer Security (TLS) client policy.

The `tls` object supports the following:

* `enforced` - (Optional) Whether the policy is enforced. Default is `true`.
* `ports` - (Optional) One or more ports that the policy is enforced for.
* `validation` - (Required) The TLS validation context.

The `validation` object supports the following:

* `trust` - (Required) The TLS validation context trust.

The `trust` object supports the following:

* `acm` - (Optional) The TLS validation context trust for an AWS Certificate Manager (ACM) certificate.
* `file` - (Optional) The TLS validation context trust for a local file.

The `acm` object supports the following:

* `certificate_authority_arns` - (Required) One or more ACM Amazon Resource Name (ARN)s.

The `file` object supports the following:

* `certificate_chain` - (Required) The certificate trust chain for a certificate stored on the file system of the mesh endpoint that the proxy is running on. Must be between 1 and 255 characters in length.

The `listener` object supports the following:

* `port_mapping` - (Required) The port mapping information for the listener.
* `connection_pool` - (Optional) The connection pool information for the listener.
* `health_check` - (Optional) The health check information for the listener.
* `tls` - (Optional) The Transport Layer Security (TLS) properties for the listener

The `logging` object supports the following:

* `access_log` - (Optional) The access log configuration for a virtual gateway.

The `access_log` object supports the following:

* `file` - (Optional) The file object to send virtual gateway access logs to.

The `file` object supports the following:

* `path` - (Required) The file path to write access logs to. You can use `/dev/stdout` to send access logs to standard out. Must be between 1 and 255 characters in length.

The `port_mapping` object supports the following:

* `port` - (Required) The port used for the port mapping.
* `protocol` - (Required) The protocol used for the port mapping. Valid values are `http`, `http2`, `tcp` and `grpc`.

The `connection_pool` object supports the following:

* `grpc` - (Optional) Connection pool information for gRPC listeners.
* `http` - (Optional) Connection pool information for HTTP listeners.
* `http2` - (Optional) Connection pool information for HTTP2 listeners.

The `grpc` connection pool object supports the following:

* `max_requests` - (Required) Maximum number of inflight requests Envoy can concurrently support across hosts in upstream cluster. Minimum value of `1`.

The `http` connection pool object supports the following:

* `max_connections` - (Required) Maximum number of outbound TCP connections Envoy can establish concurrently with all hosts in upstream cluster. Minimum value of `1`.
* `max_pending_requests` - (Optional) Number of overflowing requests after `max_connections` Envoy will queue to upstream cluster. Minimum value of `1`.

The `http2` connection pool object supports the following:

* `max_requests` - (Required) Maximum number of inflight requests Envoy can concurrently support across hosts in upstream cluster. Minimum value of `1`.

The `health_check` object supports the following:

* `healthy_threshold` - (Required) The number of consecutive successful health checks that must occur before declaring listener healthy.
* `interval_millis`- (Required) The time period in milliseconds between each health check execution.
* `protocol` - (Required) The protocol for the health check request. Valid values are `http`, `http2`, and `grpc`.
* `timeout_millis` - (Required) The amount of time to wait when receiving a response from the health check, in milliseconds.
* `unhealthy_threshold` - (Required) The number of consecutive failed health checks that must occur before declaring a virtual gateway unhealthy.
* `path` - (Optional) The destination path for the health check request. This is only required if the specified protocol is `http` or `http2`.
* `port` - (Optional) The destination port for the health check request. This port must match the port defined in the `port_mapping` for the listener.

The `tls` object supports the following:

* `certificate` - (Required) The listener's TLS certificate.
* `mode`- (Required) The listener's TLS mode. Valid values: `DISABLED`, `PERMISSIVE`, `STRICT`.

The `certificate` object supports the following:

* `acm` - (Optional) An AWS Certificate Manager (ACM) certificate.
* `file` - (optional) A local file certificate.

The `acm` object supports the following:

* `certificate_arn` - (Required) The Amazon Resource Name (ARN) for the certificate.

The `file` object supports the following:

* `certificate_chain` - (Required) The certificate chain for the certificate. Must be between 1 and 255 characters in length.
* `private_key` - (Required) The private key for a certificate stored on the file system of the mesh endpoint that the proxy is running on. Must be between 1 and 255 characters in length.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the virtual gateway.
* `arn` - The ARN of the virtual gateway.
* `created_date` - The creation date of the virtual gateway.
* `last_updated_date` - The last update date of the virtual gateway.
* `resource_owner` - The resource owner's AWS account ID.

## Import

App Mesh virtual gateway can be imported using `mesh_name` together with the virtual gateway's `name`,
e.g.

```
$ terraform import aws_appmesh_virtual_gateway.example mesh/gw1
```

[1]: /docs/providers/aws/index.html
