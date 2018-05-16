---
layout: "aws"
page_title: "AWS: aws_ecs_service"
sidebar_current: "docs-aws-resource-ecs-service"
description: |-
  Provides an ECS service.
---

# aws_ecs_service

-> **Note:** To prevent a race condition during service deletion, make sure to set `depends_on` to the related `aws_iam_role_policy`; otherwise, the policy may be destroyed too soon and the ECS service will then get stuck in the `DRAINING` state.

Provides an ECS service - effectively a task that is expected to run until an error occurs or a user terminates it (typically a webserver or a database).

See [ECS Services section in AWS developer guide](https://docs.aws.amazon.com/AmazonECS/latest/developerguide/ecs_services.html).

## Example Usage

```hcl
resource "aws_ecs_service" "mongo" {
  name            = "mongodb"
  cluster         = "${aws_ecs_cluster.foo.id}"
  task_definition = "${aws_ecs_task_definition.mongo.arn}"
  desired_count   = 3
  iam_role        = "${aws_iam_role.foo.arn}"
  depends_on      = ["aws_iam_role_policy.foo"]

  ordered_placement_strategy {
    type  = "binpack"
    field = "cpu"
  }

  load_balancer {
    target_group_arn = "${aws_lb_target_group.foo.arn}"
    container_name   = "mongo"
    container_port   = 8080
  }

  placement_constraints {
    type       = "memberOf"
    expression = "attribute:ecs.availability-zone in [us-west-2a, us-west-2b]"
  }
}
```

### Ignoring Changes to Desired Count

You can utilize the generic Terraform resource [lifecycle configuration block](/docs/configuration/resources.html#lifecycle) with `ignore_changes` to create an ECS service with an initial count of running instances, then ignore any changes to that count caused externally (e.g. Application Autoscaling).

```hcl
resource "aws_ecs_service" "example" {
  # ... other configurations ...

  # Example: Create service with 2 instances to start
  desired_count = 2

  # Optional: Allow external changes without Terraform plan difference
  lifecycle {
    ignore_changes = ["desired_count"]
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the service (up to 255 letters, numbers, hyphens, and underscores)
* `task_definition` - (Required) The family and revision (`family:revision`) or full ARN of the task definition that you want to run in your service.
* `desired_count` - (Required) The number of instances of the task definition to place and keep running
* `launch_type` - (Optional) The launch type on which to run your service. The valid values are `EC2` and `FARGATE`. Defaults to `EC2`.
* `cluster` - (Optional) ARN of an ECS cluster
* `iam_role` - (Optional) The ARN of IAM role that allows your Amazon ECS container agent to make calls to your load balancer on your behalf. This parameter is only required if you are using a load balancer with your service.
* `deployment_maximum_percent` - (Optional) The upper limit (as a percentage of the service's desiredCount) of the number of running tasks that can be running in a service during a deployment.
* `deployment_minimum_healthy_percent` - (Optional) The lower limit (as a percentage of the service's desiredCount) of the number of running tasks that must remain running and healthy in a service during a deployment.
* `placement_strategy` - (Optional) **Deprecated**, use `ordered_placement_strategy` instead.
* `ordered_placement_strategy` - (Optional) Service level strategy rules that are taken into consideration during task placement. List from top to bottom in order of precedence. The maximum number of `ordered_placement_strategy` blocks is `5`. Defined below.
* `health_check_grace_period_seconds` - (Optional) Seconds to ignore failing load balancer health checks on newly instantiated tasks to prevent premature shutdown, up to 7200. Only valid for services configured to use load balancers.
* `load_balancer` - (Optional) A load balancer block. Load balancers documented below.
* `placement_constraints` - (Optional) rules that are taken into consideration during task placement. Maximum number of
`placement_constraints` is `10`. Defined below.
* `network_configuration` - (Optional) The network configuration for the service. This parameter is required for task definitions that use the awsvpc network mode to receive their own Elastic Network Interface, and it is not supported for other network modes.
* `service_registries` - (Optional) The service discovery registries for the service. The maximum number of `service_registries` blocks is `1`.

-> **Note:** As a result of an AWS limitation, a single `load_balancer` can be attached to the ECS service at most. See [related docs](https://docs.aws.amazon.com/AmazonECS/latest/developerguide/service-load-balancing.html#load-balancing-concepts).

Load balancers support the following:

* `elb_name` - (Required for ELB Classic) The name of the ELB (Classic) to associate with the service.
* `target_group_arn` - (Required for ALB/NLB) The ARN of the Load Balancer target group to associate with the service.
* `container_name` - (Required) The name of the container to associate with the load balancer (as it appears in a container definition).
* `container_port` - (Required) The port on the container to associate with the load balancer.

## ordered_placement_strategy

`ordered_placement_strategy` supports the following:

* `type` - (Required) The type of placement strategy. Must be one of: `binpack`, `random`, or `spread`
* `field` - (Optional) For the `spread` placement strategy, valid values are instanceId (or host,
 which has the same effect), or any platform or custom attribute that is applied to a container instance.
 For the `binpack` type, valid values are `memory` and `cpu`. For the `random` type, this attribute is not
 needed. For more information, see [Placement Strategy](https://docs.aws.amazon.com/AmazonECS/latest/APIReference/API_PlacementStrategy.html).

-> **Note:** for `spread`, `host` and `instanceId` will be normalized, by AWS, to be `instanceId`. This means the statefile will show `instanceId` but your config will differ if you use `host`.

## placement_constraints

`placement_constraints` support the following:

* `type` - (Required) The type of constraint. The only valid values at this time are `memberOf` and `distinctInstance`.
* `expression` -  (Optional) Cluster Query Language expression to apply to the constraint. Does not need to be specified
for the `distinctInstance` type.
For more information, see [Cluster Query Language in the Amazon EC2 Container
Service Developer
Guide](https://docs.aws.amazon.com/AmazonECS/latest/developerguide/cluster-query-language.html).

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

## Attributes Reference

The following attributes are exported:

* `id` - The Amazon Resource Name (ARN) that identifies the service
* `name` - The name of the service
* `cluster` - The Amazon Resource Name (ARN) of cluster which the service runs on
* `iam_role` - The ARN of IAM role used for ELB
* `desired_count` - The number of instances of the task definition

## Import

ECS services can be imported using the `name` together with ecs cluster `name`, e.g.

```
$ terraform import aws_ecs_service.imported cluster-name/service-name
```
