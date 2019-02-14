---
layout: "aws"
page_title: "AWS: aws_service_discovery_private_dns_namespace"
sidebar_current: "docs-aws-resource-service-discovery-private-dns-namespace"
description: |-
  Provides a Service Discovery Private DNS Namespace resource.
---

# aws_service_discovery_private_dns_namespace

Provides a Service Discovery Private DNS Namespace resource.

## Example Usage

```hcl
resource "aws_vpc" "example" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_service_discovery_private_dns_namespace" "example" {
  name        = "hoge.example.local"
  description = "example"
  vpc         = "${aws_vpc.example.id}"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the namespace.
* `vpc` - (Required) The ID of VPC that you want to associate the namespace with.
* `description` - (Optional) The description that you specify for the namespace when you create it.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of a namespace.
* `arn` - The ARN that Amazon Route 53 assigns to the namespace when you create it.
* `hosted_zone` - The ID for the hosted zone that Amazon Route 53 creates when you create a namespace.
