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

```hcl
resource "aws_ecs_cluster" "foo" {
  name = "white-hart"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the cluster (up to 255 letters, numbers, hyphens, and underscores)
* `capacity_providers` - (Optional) List of short names of one or more capacity providers to associate with the cluster. Valid values also include `FARGATE` and `FARGATE_SPOT`.
* `default_capacity_provider_strategy` - (Optional) The capacity provider strategy to use by default for the cluster. Can be one or more.  Defined below.
* `tags` - (Optional) Key-value map of resource tags
* `setting` - (Optional) Configuration block(s) with cluster settings. For example, this can be used to enable CloudWatch Container Insights for a cluster. Defined below.

## setting

The `setting` configuration block supports the following:

* `name` - (Required) Name of the setting to manage. Valid values: `containerInsights`.
* `value` -  (Required) The value to assign to the setting. Value values are `enabled` and `disabled`.

## default_capacity_provider_strategy

The `default_capacity_provider_strategy` configuration block supports the following:

* `capacity_provider` - (Required) The short name of the capacity provider.
* `weight` - (Optional) The relative percentage of the total number of launched tasks that should use the specified capacity provider.
* `base` - (Optional) The number of tasks, at a minimum, to run on the specified capacity provider. Only one capacity provider in a capacity provider strategy can have a base defined.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The Amazon Resource Name (ARN) that identifies the cluster
* `arn` - The Amazon Resource Name (ARN) that identifies the cluster

## Import

ECS clusters can be imported using the `name`, e.g.

```
$ terraform import aws_ecs_cluster.stateless stateless-app
```
