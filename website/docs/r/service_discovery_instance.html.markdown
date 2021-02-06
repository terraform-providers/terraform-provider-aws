---
subcategory: "Service Discovery"
layout: "aws"
page_title: "AWS: aws_service_discovery_instance"
description: |-
  Provides a Service Discovery Instance resource.
---

# Resource: aws_service_discovery_instance

Provides a Service Discovery Instance resource.

## Example Usage

```hcl
resource "aws_vpc" "example" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_support   = true
  enable_dns_hostnames = true
}

resource "aws_service_discovery_private_dns_namespace" "example" {
  name        = "example.terraform.local"
  description = "example"
  vpc         = aws_vpc.example.id
}

resource "aws_service_discovery_service" "example" {
  name = "example"

  dns_config {
    namespace_id = aws_service_discovery_private_dns_namespace.example.id

    dns_records {
      ttl  = 10
      type = "A"
    }

    routing_policy = "MULTIVALUE"
  }

  health_check_custom_config {
    failure_threshold = 1
  }
}

resource "aws_service_discovery_instance" "example" {
  instance_id = "example-instance-id"
  service_id  = aws_service_discovery_service.example.id
  attributes  = {
    AWS_INSTANCE_IPV4 = "172.18.0.1"
    custom_attribute  = "custom"
  }
} 
```

```hcl
resource "aws_service_discovery_http_namespace" "example" {
  name        = "example.terraform.com"
  description = "example"
}

resource "aws_service_discovery_service" "example" {
  name         = "example"
  namespace_id = aws_service_discovery_http_namespace.example.id
}

resource "aws_service_discovery_instance" "example" {
  instance_id = "example-instance-id"
  service_id  = aws_service_discovery_service.example.id
  attributes  = {
    AWS_EC2_INSTANCE_ID = "i-0abdg374kd892cj6dl"
  }
}
```

## Argument Reference

The following arguments are supported:

* `instance_id` - (Required, ForceNew) The ID of the service instance.
* `service_id` - (Required, ForceNew) The ID of the service that you want to use to create the instance.
* `dns_config` - (Optional) A complex type that contains information about the resource record sets that you want Amazon Route 53 to create when you register an instance.
* `attributes` - (Required) A complex type that contains the attributes of the instance. Check the [doc](https://docs.aws.amazon.com/cloud-map/latest/api/API_RegisterInstance.html#API_RegisterInstance_RequestSyntax) for the supported attributes and syntax.
* `creator_request_id` - (Optional) A unique string that identifies the request and that allows failed requests to be retried without the risk of executing the operation twice. You must use a unique string every time you submit a request if you're registering additional instances for the same namespace and service.

## Import

Service Discovery Instance can be imported using the service ID and instance ID, e.g.

```
$ terraform import aws_service_discovery_instance.example 0123456789/i-0123
```
