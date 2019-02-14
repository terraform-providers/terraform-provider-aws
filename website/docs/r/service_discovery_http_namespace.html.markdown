---
layout: "aws"
page_title: "AWS: aws_service_discovery_http_namespace"
sidebar_current: "docs-aws-resource-service-discovery-http-namespace"
description: |-
  Provides a Service Discovery HTTP Namespace resource.
---

# aws_service_discovery_http_namespace


## Example Usage

```hcl
resource "aws_service_discovery_http_namespace" "example" {
  name        = "development"
  description = "example"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the http namespace.
* `description` - (Optional) The description that you specify for the namespace when you create it.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of a namespace.
* `arn` - The ARN that Amazon Route 53 assigns to the namespace when you create it.

## Import

Service Discovery HTTP Namespace can be imported using the namespace ID, e.g.

```
$ terraform import aws_service_discovery_http_namespace.example ns-1234567890
```
