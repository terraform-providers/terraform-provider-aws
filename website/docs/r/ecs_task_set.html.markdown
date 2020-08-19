---
subcategory: "ECS"
layout: "aws"
page_title: "AWS: aws_ecs_task_set"
description: |-
  Provides an ECS task set.
---

# Resource: aws_ecs_task_set

Provides an ECS task set - effectively a task that is expected to run until an error occurs or a user terminates it (typically a webserver or a database).

See [ECS Task Set section in AWS developer guide](https://docs.amazonaws.cn/en_us/AmazonECS/latest/userguide/deployment-type-external.html).

## Example Usage

```hcl
resource "aws_ecs_task_set" "mongo" {
  service         = "${aws_ecs_service.foo.id}"
  cluster         = "${aws_ecs_cluster.foo.id}"
  task_definition = "${aws_ecs_task_definition.mongo.arn}"

  load_balancer {
    target_group_arn = "${aws_lb_target_group.foo.arn}"
    container_name   = "mongo"
    container_port   = 8080
  }
}
```

### Ignoring Changes to Scale

You can utilize the generic Terraform resource [lifecycle configuration block](/docs/configuration/resources.html#lifecycle) with `ignore_changes` to create an ECS service with an initial count of running instances, then ignore any changes to that count caused externally (e.g. Application Autoscaling).

```hcl
resource "aws_ecs_task_set" "example" {
  # ... other configurations ...

  # Example: Run 50% of the servcie's desired count
  scale {
    value = 50.0
  }

  # Optional: Allow external changes without Terraform plan difference
  lifecycle {
    ignore_changes = ["scale"]
  }
}
```

## Argument Reference

The following arguments are supported:

* `service` - (Required) The name or ARN of the ECS service.
* `cluster` - (Required) The name or ARN of an ECS cluster.
* `external_id` - (Optional) The external ID associated with the task set.
* `task_definition` - (Required) The family and revision (`family:revision`) or full ARN of the task definition that you want to run in your service.
* `network_configuration` - (Optional) The network configuration for the service. This parameter is required for task definitions that use the `awsvpc` network mode to receive their own Elastic Network Interface, and it is not supported for other network modes.
* `load_balancer` - (Optional) A load balancer block. Load balancers documented below.
* `service_registries` - (Optional) The service discovery registries for the service. The maximum number of `service_registries` blocks is `1`.
* `launch_type` - (Optional) The launch type on which to run your service. The valid values are `EC2` and `FARGATE`. Defaults to `EC2`.
* `capacity_provider_strategy` - (Optional) The capacity provider strategy to use for the service. Can be one or more.  Defined below.
* `platform_version` - (Optional) The platform version on which to run your service. Only applicable for `launch_type` set to `FARGATE`. Defaults to `LATEST`. More information about Fargate platform versions can be found in the [AWS ECS User Guide](https://docs.aws.amazon.com/AmazonECS/latest/developerguide/platform_versions.html).
* `scale` - (Optional) A floating-point percentage of the desired number of tasks to place and keep running in the task set. `unit` determines the interpretation of this number (a percentage of the service's desired count). Accepted values are numbers between 0 and 100.
* `force_delete` - (Optional) Allows deleting the task set without waiting for scaling down to 0. You can force a task set to delete even if it's in the process of scaling a resource. Normally, Terraform drains all the tasks before deleting the task set. This bypasses that behavior and potentially leaves resources dangling.
* `wait_until_stable` - (Optional) Apply will wait until the task set has reached `STEADY_STATE`
* `wait_until_stable_timeout` - (Optional) Wait timeout for task set to reach `STEADY_STATE`. Default `10m`
* `tags` - (Optional) Key-value map of resource tags

## capacity_provider_strategy

The `capacity_provider_strategy` configuration block supports the following:

* `capacity_provider` - (Required) The short name or full Amazon Resource Name (ARN) of the capacity provider.
* `weight` - (Required) The relative percentage of the total number of launched tasks that should use the specified capacity provider.
* `base` - (Optional) The number of tasks, at a minimum, to run on the specified capacity provider. Only one capacity provider in a capacity provider strategy can have a base defined.

## scale

The `scale` configuration block supports the following:

* `unit` - (Optional) The unit of measure for the scale value. Default: `PERCENT`
* `value` - (Required) The value, specified as a percent total of a service's `desiredCount`, to scale the task set. Accepted values are numbers between 0.0 and 100.0.

## load_balancer

`load_balancer` supports the following:

* `elb_name` - (Required for ELB Classic) The name of the ELB (Classic) to associate with the service.
* `target_group_arn` - (Required for ALB/NLB) The ARN of the Load Balancer target group to associate with the service.
* `container_name` - (Required) The name of the container to associate with the load balancer (as it appears in a container definition).
* `container_port` - (Required) The port on the container to associate with the load balancer.

-> **Note:** Multiple `load_balancer` configuration is still not supported by AWS for ECS task set.

## network_configuration

`network_configuration` support the following:

* `subnets` - (Required) The subnets associated with the task or service.
* `security_groups` - (Optional) The security groups associated with the task or service. If you do not specify a security group, the default security group for the VPC is used.
* `assign_public_ip` - (Optional) Assign a public IP address to the ENI (Fargate launch type only). Valid values are `true` or `false`. Default `false`.

For more information, see [Task Networking](https://docs.aws.amazon.com/AmazonECS/latest/developerguide/task-networking.html)

## service_registries

`service_registries` support the following:

* `registry_arn` - (Required) The ARN of the Service Registry. The currently supported service registry is Amazon Route 53 Auto Naming Service(`aws_service_discovery_service`). For more information, see [Service](https://docs.aws.amazon.com/Route53/latest/APIReference/API_autonaming_Service.html)
* `port` - (Optional) The port value used if your Service Discovery service specified an SRV record.
* `container_port` - (Optional) The port value, already specified in the task definition, to be used for your service discovery service.
* `container_name` - (Optional) The container name value, already specified in the task definition, to be used for your service discovery service.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the task set
* `arn` - The Amazon Resource Name (ARN) that identifies the task set
