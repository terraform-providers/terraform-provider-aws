---
subcategory: "Service Discovery"
layout: "aws"
page_title: "AWS: aws_service_discovery_http_namespace"
description: |-
  Provides a Service Discovery HTTP Namespace resource.
---

# Resource: aws_service_discovery_http_namespace


## Example Usage

```terraform
resource "aws_service_discovery_http_namespace" "example" {
  name        = "development"
  description = "example"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the http namespace.
* `description` - (Optional) The description that you specify for the namespace when you create it.
* `tags` - (Optional) A map of tags to assign to the namespace. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of a namespace.
* `arn` - The ARN that Amazon Route 53 assigns to the namespace when you create it.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block).

## Import

Service Discovery HTTP Namespace can be imported using the namespace ID, e.g.

```
$ terraform import aws_service_discovery_http_namespace.example ns-1234567890
```
