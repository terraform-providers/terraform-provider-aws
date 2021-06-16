---
subcategory: "Step Function (SFN)"
layout: "aws"
page_title: "AWS: aws_sfn_activity"
description: |-
  Provides a Step Function Activity resource.
---

# Resource: aws_sfn_activity

Provides a Step Function Activity resource

## Example Usage

```terraform
resource "aws_sfn_activity" "sfn_activity" {
  name = "my-activity"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the activity to create.
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The Amazon Resource Name (ARN) that identifies the created activity.
* `name` - The name of the activity.
* `creation_date` - The date the activity was created.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block).

## Import

Activities can be imported using the `arn`, e.g.

```
$ terraform import aws_sfn_activity.foo arn:aws:states:eu-west-1:123456789098:activity:bar
```
