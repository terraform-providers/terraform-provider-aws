---
subcategory: "ElastiCache"
layout: "aws"
page_title: "AWS: aws_elasticache_parameter_group"
description: |-
  Provides an ElastiCache parameter group resource.
---

# Resource: aws_elasticache_parameter_group

Provides an ElastiCache parameter group resource.

~> **NOTE:** Attempting to remove the `reserved-memory` parameter when `family` is set to `redis2.6` or `redis2.8` may show a perpetual difference in Terraform due to an Elasticache API limitation. Leave that parameter configured with any value to workaround the issue.

## Example Usage

```terraform
resource "aws_elasticache_parameter_group" "default" {
  name   = "cache-params"
  family = "redis2.8"

  parameter {
    name  = "activerehashing"
    value = "yes"
  }

  parameter {
    name  = "min-slaves-to-write"
    value = "2"
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the ElastiCache parameter group.
* `family` - (Required) The family of the ElastiCache parameter group.
* `description` - (Optional) The description of the ElastiCache parameter group. Defaults to "Managed by Terraform".
* `parameter` - (Optional) A list of ElastiCache parameters to apply.
* `tags` - (Optional) Key-value mapping of resource tags. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

Parameter blocks support the following:

* `name` - (Required) The name of the ElastiCache parameter.
* `value` - (Required) The value of the ElastiCache parameter.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ElastiCache parameter group name.
* `arn` - The AWS ARN associated with the parameter group.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block).


## Import

ElastiCache Parameter Groups can be imported using the `name`, e.g.

```
$ terraform import aws_elasticache_parameter_group.default redis-params
```
