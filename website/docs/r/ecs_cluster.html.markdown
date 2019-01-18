---
layout: "aws"
page_title: "AWS: aws_ecs_cluster"
sidebar_current: "docs-aws-resource-ecs-cluster"
description: |-
  Provides an ECS cluster.
---

# aws_ecs_cluster

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
* `tags` - (Optional) Key-value mapping of resource tags

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The Amazon Resource Name (ARN) that identifies the cluster
* `arn` - The Amazon Resource Name (ARN) that identifies the cluster

## Import

ECS clusters can be imported using the `name`, e.g.

```
$ terraform import aws_ecs_cluster.stateless stateless-app
```
