---
subcategory: "Service Discovery"
layout: "aws"
page_title: "AWS: aws_service_discovery_dns_namespace"
description: |-
  Provides details about a Service Discovery DNS Namespace.
---

# Data Source: aws_service_discovery_dns_namespace

`aws_service_discovery_dns_namespace` provides details about a specific Service Discovery DNS Namespace.

This resource can prove useful when you need to re-use a Service Discovery DNS Namespace for a new Service Discovery Service. 

## Example Usage

```hcl
data "aws_service_discovery_dns_namespace" "example" {
  name = "example.terraform.local"
  type = "private"
}

resource "aws_service_discovery_service" "example" {
  name = "example"

  dns_config {
    namespace_id = "${data.aws_service_discovery_dns_namespace.example.id}"

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
```

```hcl
data "aws_service_discovery_dns_namespace" "example" {
  name = "example.terraform.com"
  type = "public"
}

resource "aws_service_discovery_service" "example" {
  name = "example"

  dns_config {
    namespace_id = "${data.aws_service_discovery_public_dns_namespace.example.id}"

    dns_records {
      ttl  = 10
      type = "A"
    }
  }

  health_check_config {
    failure_threshold = 10
    resource_path     = "path"
    type              = "HTTP"
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` – (Required) The name of the namespace.
* `type` – (Required) The type of the namespace. Valid Values: public, private.
* `most_recent` – (Optional) If more than one result is returned, use the most recent namespace.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` – The ID of a namespace.
* `arn` – The ARN that Amazon Route 53 assigns to the namespace.
* `hosted_zone` – The ID for the hosted zone that Amazon Route 53 creates for a namespace.