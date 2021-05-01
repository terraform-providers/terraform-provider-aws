---
subcategory: "ECS"
layout: "aws"
page_title: "AWS: aws_ecs_cluster"
description: |-
  Provides an ECS cluster.
---

# Resource: aws_ecs_cluster

Provides an ECS cluster.

## Example Usage

```terraform
resource "aws_ecs_cluster" "foo" {
  name = "white-hart"

  setting {
    name  = "containerInsights"
    value = "enabled"
  }
}
```

## Argument Reference

The following arguments are supported:

* `capacity_providers` - (Optional) List of short names of one or more capacity providers to associate with the cluster. Valid values also include `FARGATE` and `FARGATE_SPOT`.
* `default_capacity_provider_strategy` - (Optional) Configuration block for capacity provider strategy to use by default for the cluster. Can be one or more. Detailed below.
* `name` - (Required) Name of the cluster (up to 255 letters, numbers, hyphens, and underscores)
* `setting` - (Optional) Configuration block(s) with cluster settings. For example, this can be used to enable CloudWatch Container Insights for a cluster. Detailed below.
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://www.terraform.io/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### `default_capacity_provider_strategy`

* `capacity_provider` - (Required) The short name of the capacity provider.
* `weight` - (Optional) The relative percentage of the total number of launched tasks that should use the specified capacity provider.
* `base` - (Optional) The number of tasks, at a minimum, to run on the specified capacity provider. Only one capacity provider in a capacity provider strategy can have a base defined.

### `setting`

* `name` - (Required) Name of the setting to manage. Valid values: `containerInsights`.
* `value` -  (Required) The value to assign to the setting. Value values are `enabled` and `disabled`.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - ARN that identifies the cluster.
* `id` - ARN that identifies the cluster.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://www.terraform.io/docs/providers/aws/index.html#default_tags-configuration-block).

## Import

ECS clusters can be imported using the `name`, e.g.

```
$ terraform import aws_ecs_cluster.stateless stateless-app
```
