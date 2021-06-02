---
subcategory: "MediaStore"
layout: "aws"
page_title: "AWS: aws_media_store_container"
description: |-
  Provides a MediaStore Container.
---

# Resource: aws_media_store_container

Provides a MediaStore Container.

## Example Usage

```terraform
resource "aws_media_store_container" "example" {
  name = "example"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the container. Must contain alphanumeric characters or underscores.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The ARN of the container.
* `endpoint` - The DNS endpoint of the container.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block).

## Import

MediaStore Container can be imported using the MediaStore Container Name, e.g.

```
$ terraform import aws_media_store_container.example example
```
