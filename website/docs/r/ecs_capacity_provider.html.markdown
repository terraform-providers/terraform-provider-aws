---
subcategory: "ECS"
layout: "aws"
page_title: "AWS: aws_ecs_capacity_provider"
description: |-
  Provides an ECS cluster capacity provider.
---

# Resource: aws_ecs_capacity_provider

Provides an ECS cluster capacity provider. More information can be found on the [ECS Developer Guide](https://docs.aws.amazon.com/AmazonECS/latest/developerguide/cluster-capacity-providers.html).

~> **NOTE:** The AWS API does not currently support deleting ECS cluster capacity providers. Removing this Terraform resource will only remove the Terraform state for it.

## Example Usage

```hcl
resource "aws_ecs_capacity_provider" "test" {
  name = "test"

  auto_scaling_group_provider {
    auto_scaling_group_arn         = aws_autoscaling_group.test.arn
    managed_termination_protection = "ENABLED"

    managed_scaling {
      maximum_scaling_step_size = 1000
      minimum_scaling_step_size = 1
      status                    = "ENABLED"
      target_capacity           = 10
    }
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the capacity provider.
* `auto_scaling_group_provider` - (Required) Nested argument defining the provider for the ECS auto scaling group. Defined below.
* `tags` - (Optional) Key-value mapping of resource tags.

## auto_scaling_group_provider

The `auto_scaling_group_provider` block supports the following:

* `auto_scaling_group_arn` - (Required) - The Amazon Resource Name (ARN) of the associated auto scaling group.
* `managed_scaling` - (Optional) - Nested argument defining the parameters of the auto scaling. Defined below.
* `managed_termination_protection` - (Optional) - Enables or disables container-aware termination of instances in the auto scaling group when scale-in happens. Valid values are `ENABLED` and `DISABLED`.

## managed_scaling

The `managed_scaling` block supports the following:

* `maximum_scaling_step_size` - (Optional) The maximum step adjustment size. A number between 1 and 10,000.
* `minimum_scaling_step_size` - (Optional) The minimum step adjustment size. A number between 1 and 10,000.
* `status` - (Optional) Whether auto scaling is managed by ECS. Valid values are `ENABLED` and `DISABLED`.
* `target_capacity` - (Optional) The target utilization for the capacity provider. A number between 1 and 100.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The Amazon Resource Name (ARN) that identifies the capacity provider.
* `arn` - The Amazon Resource Name (ARN) that identifies the capacity provider.

## Import

ECS Capacity Providers can be imported using the `name`, e.g.

```
$ terraform import aws_ecs_capacity_provider.example example
```
