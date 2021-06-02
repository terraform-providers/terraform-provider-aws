---
subcategory: "Service Discovery"
layout: "aws"
page_title: "AWS: aws_service_discovery_dns_namespace"
description: |-
  Retrieves information about a Service Discovery private or public DNS namespace.
---

# Data Source: aws_service_discovery_dns_namespace

Retrieves information about a Service Discovery private or public DNS namespace.

## Example Usage

```hcl
data "aws_service_discovery_dns_namespace" "test" {
  name = "example.terraform.local"
  type = "DNS_PRIVATE"
}
```

## Argument Reference

* `name` - (Required) The name of the namespace.
* `type` - (Required) The type of the namespace. Allowed values are `DNS_PUBLIC` or `DNS_PRIVATE`.

## Attributes Reference

* `arn` - The Amazon Resource Name (ARN) of the namespace.
* `description` - A description of the namespace.
* `id` - The namespace ID.
* `hosted_zone` - The ID for the hosted zone that Amazon Route 53 creates when you create a namespace.
