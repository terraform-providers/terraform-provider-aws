---
subcategory: "AppMesh"
layout: "aws"
page_title: "AWS: aws_appmesh_virtual_node"
description: |-
  Provides an AWS App Mesh virtual node resource.
---

# Resource: aws_appmesh_virtual_node

Provides an AWS App Mesh virtual node resource.

## Breaking Changes

Because of backward incompatible API changes (read [here](https://github.com/awslabs/aws-app-mesh-examples/issues/92)), `aws_appmesh_virtual_node` resource definitions created with provider versions earlier than v2.3.0 will need to be modified:

* Rename the `service_name` attribute of the `dns` object to `hostname`.

* Replace the `backends` attribute of the `spec` object with one or more `backend` configuration blocks,
setting `virtual_service_name` to the name of the service.

The Terraform state associated with existing resources will automatically be migrated.

## Example Usage

### Basic

```hcl
resource "aws_appmesh_virtual_node" "serviceb1" {
  name      = "serviceBv1"
  mesh_name = aws_appmesh_mesh.simple.id

  spec {
    backend {
      virtual_service {
        virtual_service_name = "servicea.simpleapp.local"
      }
    }

    listener {
      port_mapping {
        port     = 8080
        protocol = "http"
      }
    }

    service_discovery {
      dns {
        hostname = "serviceb.simpleapp.local"
      }
    }
  }
}
```

### AWS Cloud Map Service Discovery

```hcl
resource "aws_service_discovery_http_namespace" "example" {
  name = "example-ns"
}

resource "aws_appmesh_virtual_node" "serviceb1" {
  name      = "serviceBv1"
  mesh_name = aws_appmesh_mesh.simple.id

  spec {
    backend {
      virtual_service {
        virtual_service_name = "servicea.simpleapp.local"
      }
    }

    listener {
      port_mapping {
        port     = 8080
        protocol = "http"
      }
    }

    service_discovery {
      aws_cloud_map {
        attributes = {
          stack = "blue"
        }

        service_name   = "serviceb1"
        namespace_name = aws_service_discovery_http_namespace.example.name
      }
    }
  }
}
```

### Listener Health Check

```hcl
resource "aws_appmesh_virtual_node" "serviceb1" {
  name      = "serviceBv1"
  mesh_name = aws_appmesh_mesh.simple.id

  spec {
    backend {
      virtual_service {
        virtual_service_name = "servicea.simpleapp.local"
      }
    }

    listener {
      port_mapping {
        port     = 8080
        protocol = "http"
      }

      health_check {
        protocol            = "http"
        path                = "/ping"
        healthy_threshold   = 2
        unhealthy_threshold = 2
        timeout_millis      = 2000
        interval_millis     = 5000
      }
    }

    service_discovery {
      dns {
        hostname = "serviceb.simpleapp.local"
      }
    }
  }
}
```

### Logging

```hcl
resource "aws_appmesh_virtual_node" "serviceb1" {
  name      = "serviceBv1"
  mesh_name = aws_appmesh_mesh.simple.id

  spec {
    backend {
      virtual_service {
        virtual_service_name = "servicea.simpleapp.local"
      }
    }

    listener {
      port_mapping {
        port     = 8080
        protocol = "http"
      }
    }

    service_discovery {
      dns {
        hostname = "serviceb.simpleapp.local"
      }
    }

    logging {
      access_log {
        file {
          path = "/dev/stdout"
        }
      }
    }
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name to use for the virtual node.
* `mesh_name` - (Required) The name of the service mesh in which to create the virtual node.
* `mesh_owner` - (Optional) The AWS account ID of the service mesh's owner. Defaults to the account ID the [AWS provider][1] is currently connected to.
* `spec` - (Required) The virtual node specification to apply.
* `tags` - (Optional) A map of tags to assign to the resource.

The `spec` object supports the following:

* `backend` - (Optional) The backends to which the virtual node is expected to send outbound traffic.
* `backend_defaults` - (Optional) The defaults for backends.
* `listener` - (Optional) The listeners from which the virtual node is expected to receive inbound traffic.
* `logging` - (Optional) The inbound and outbound access logging information for the virtual node.
* `service_discovery` - (Optional) The service discovery information for the virtual node.

The `backend` object supports the following:

* `virtual_service` - (Required) Specifies a virtual service to use as a backend for a virtual node.

The `virtual_service` object supports the following:

* `client_policy` - (Optional) The client policy for the backend.
* `virtual_service_name` - (Required) The name of the virtual service that is acting as a virtual node backend.

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

* `certificate_chain` - (Required) The certificate trust chain for a certificate stored on the file system of the virtual node that the proxy is running on.

The `backend_defaults` object supports the following:

* `client_policy` - (Optional) The default client policy for virtual service backends. See above for details.

The `listener` object supports the following:

* `port_mapping` - (Required) The port mapping information for the listener.
* `health_check` - (Optional) The health check information for the listener.
* `timeout` - (Optional) Timeouts for different protocols.
* `tls` - (Optional) The Transport Layer Security (TLS) properties for the listener

The `logging` object supports the following:

* `access_log` - (Optional) The access log configuration for a virtual node.

The `access_log` object supports the following:

* `file` - (Optional) The file object to send virtual node access logs to.

The `file` object supports the following:

* `path` - (Required) The file path to write access logs to. You can use `/dev/stdout` to send access logs to standard out.

The `service_discovery` object supports the following:

* `aws_cloud_map` - (Optional) Specifies any AWS Cloud Map information for the virtual node.
* `dns` - (Optional) Specifies the DNS service name for the virtual node.

The `aws_cloud_map` object supports the following:

* `attributes` - (Optional) A string map that contains attributes with values that you can use to filter instances by any custom attribute that you specified when you registered the instance. Only instances that match all of the specified key/value pairs will be returned.
* `namespace_name` - (Required) The name of the AWS Cloud Map namespace to use.
Use the [`aws_service_discovery_http_namespace`](/docs/providers/aws/r/service_discovery_http_namespace.html) resource to configure a Cloud Map namespace.
* `service_name` - (Required) The name of the AWS Cloud Map service to use. Use the [`aws_service_discovery_service`](/docs/providers/aws/r/service_discovery_service.html) resource to configure a Cloud Map service.

The `dns` object supports the following:

* `hostname` - (Required) The DNS host name for your virtual node.

The `port_mapping` object supports the following:

* `port` - (Required) The port used for the port mapping.
* `protocol` - (Required) The protocol used for the port mapping. Valid values are `http`, `http2`, `tcp` and `grpc`.

The `health_check` object supports the following:

* `healthy_threshold` - (Required) The number of consecutive successful health checks that must occur before declaring listener healthy.
* `interval_millis`- (Required) The time period in milliseconds between each health check execution.
* `protocol` - (Required) The protocol for the health check request. Valid values are `http`, `http2`, `tcp` and `grpc`.
* `timeout_millis` - (Required) The amount of time to wait when receiving a response from the health check, in milliseconds.
* `unhealthy_threshold` - (Required) The number of consecutive failed health checks that must occur before declaring a virtual node unhealthy.
* `path` - (Optional) The destination path for the health check request. This is only required if the specified protocol is `http` or `http2`.
* `port` - (Optional) The destination port for the health check request. This port must match the port defined in the `port_mapping` for the listener.

The `timeout` object supports the following:

* `grpc` - (Optional) Timeouts for gRPC listeners.
* `http` - (Optional) Timeouts for HTTP listeners.
* `http2` - (Optional) Timeouts for HTTP2 listeners.
* `tcp` - (Optional) Timeouts for TCP listeners.

The `grpc` timeout object supports the following:

* `idle` - (Optional) The idle timeout. An idle timeout bounds the amount of time that a connection may be idle.
* `per_request` - (Optional) The per request timeout.

The `idle` and `per_request` objects support the following:

* `unit` - (Required) The unit of time. Valid values: `ms`, `s`.
* `value` - (Required) The number of time units. Minimum value of `0`.

The `http` and `http2` timeout objects support the following:

* `idle` - (Optional) The idle timeout. An idle timeout bounds the amount of time that a connection may be idle.
* `per_request` - (Optional) The per request timeout.

The `idle` and `per_request` objects support the following:

* `unit` - (Required) The unit of time. Valid values: `ms`, `s`.
* `value` - (Required) The number of time units. Minimum value of `0`.

The `tcp` timeout object supports the following:

* `idle` - (Optional) The idle timeout. An idle timeout bounds the amount of time that a connection may be idle.

The `idle` object supports the following:

* `unit` - (Required) The unit of time. Valid values: `ms`, `s`.
* `value` - (Required) The number of time units. Minimum value of `0`.

The `tls` object supports the following:

* `certificate` - (Required) The listener's TLS certificate.
* `mode`- (Required) The listener's TLS mode. Valid values: `DISABLED`, `PERMISSIVE`, `STRICT`.

The `certificate` object supports the following:

* `acm` - (Optional) An AWS Certificate Manager (ACM) certificate.
* `file` - (optional) A local file certificate.

The `acm` object supports the following:

* `certificate_arn` - (Required) The Amazon Resource Name (ARN) for the certificate.

The `file` object supports the following:

* `certificate_chain` - (Required) The certificate chain for the certificate.
* `private_key` - (Required) The private key for a certificate stored on the file system of the virtual node that the proxy is running on.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the virtual node.
* `arn` - The ARN of the virtual node.
* `created_date` - The creation date of the virtual node.
* `last_updated_date` - The last update date of the virtual node.
* `resource_owner` - The resource owner's AWS account ID.

## Import

App Mesh virtual nodes can be imported using `mesh_name` together with the virtual node's `name`,
e.g.

```
$ terraform import aws_appmesh_virtual_node.serviceb1 simpleapp/serviceBv1
```

[1]: /docs/providers/aws/index.html
