---
subcategory: "Glue"
layout: "aws"
page_title: "AWS: aws_glue_registry"
description: |-
  Provides a Glue Registry resource.
---

# Resource: aws_glue_registry

Provides a Glue Registry resource.

## Example Usage

```terraform
resource "aws_glue_registry" "example" {
  registry_name = "example"
}
```

## Argument Reference

The following arguments are supported:

* `registry_name` – (Required) The Name of the registry.
* `description` – (Optional) A description of the registry.
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - Amazon Resource Name (ARN) of Glue Registry.
* `id` - Amazon Resource Name (ARN) of Glue Registry.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://www.terraform.io/docs/providers/aws/index.html#default_tags-configuration-block).

## Import

Glue Registries can be imported using `arn`, e.g.

```
$ terraform import aws_glue_registry.example arn:aws:glue:us-west-2:123456789012:registry/example
```
