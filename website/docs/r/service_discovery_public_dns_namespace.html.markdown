---
layout: "aws"
page_title: "AWS: aws_service_discovery_public_dns_namespace"
sidebar_current: "docs-aws-resource-service-discovery-public-dns-namespace"
description: |-
  Provides a Service Discovery Public DNS Namespace resource.
---

# aws_service_discovery_public_dns_namespace

Provides a Service Discovery Public DNS Namespace resource.

## Example Usage

```hcl
resource "aws_service_discovery_public_dns_namespace" "example" {
  name        = "hoge.example.com"
  description = "example"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the namespace.
* `description` - (Optional) The description that you specify for the namespace when you create it.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of a namespace.
* `arn` - The ARN that Amazon Route 53 assigns to the namespace when you create it.
* `hosted_zone` - The ID for the hosted zone that Amazon Route 53 creates when you create a namespace.

## Import

Service Discovery Public DNS Namespace can be imported using the namespace ID, e.g.

```
$ terraform import aws_service_discovery_public_dns_namespace.example 0123456789
```
