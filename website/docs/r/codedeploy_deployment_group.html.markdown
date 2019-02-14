---
layout: "aws"
page_title: "AWS: aws_codedeploy_deployment_group"
sidebar_current: "docs-aws-resource-codedeploy-deployment-group"
description: |-
  Provides a CodeDeploy deployment group.
---

# aws_codedeploy_deployment_group

Provides a CodeDeploy Deployment Group for a CodeDeploy Application

~> **NOTE on blue/green deployments:** When using `green_fleet_provisioning_option` with the `COPY_AUTO_SCALING_GROUP` action, CodeDeploy will create a new ASG with a different name. This ASG is _not_ managed by terraform and will conflict with existing configuration and state. You may want to use a different approach to managing deployments that involve multiple ASG, such as `DISCOVER_EXISTING` with separate blue and green ASG.

## Example Usage

```hcl
resource "aws_iam_role" "example" {
  name = "example-role"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "codedeploy.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

resource "aws_iam_role_policy_attachment" "AWSCodeDeployRole" {
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSCodeDeployRole"
  role       = "${aws_iam_role.example.name}"
}

resource "aws_codedeploy_app" "example" {
  name = "example-app"
}

resource "aws_sns_topic" "example" {
  name = "example-topic"
}

resource "aws_codedeploy_deployment_group" "example" {
  app_name              = "${aws_codedeploy_app.example.name}"
  deployment_group_name = "example-group"
  service_role_arn      = "${aws_iam_role.example.arn}"

  ec2_tag_set {
    ec2_tag_filter {
      key   = "filterkey1"
      type  = "KEY_AND_VALUE"
      value = "filtervalue"
    }

    ec2_tag_filter {
      key   = "filterkey2"
      type  = "KEY_AND_VALUE"
      value = "filtervalue"
    }
  }

  trigger_configuration {
    trigger_events     = ["DeploymentFailure"]
    trigger_name       = "example-trigger"
    trigger_target_arn = "${aws_sns_topic.example.arn}"
  }

  auto_rollback_configuration {
    enabled = true
    events  = ["DEPLOYMENT_FAILURE"]
  }

  alarm_configuration {
    alarms  = ["my-alarm-name"]
    enabled = true
  }
}
```

### Blue Green Deployments with ECS

```hcl
resource "aws_codedeploy_app" "example" {
  compute_platform = "ECS"
  name             = "example"
}

resource "aws_codedeploy_deployment_group" "example" {
  app_name               = "${aws_codedeploy_app.example.name}"
  deployment_config_name = "CodeDeployDefault.ECSAllAtOnce"
  deployment_group_name  = "example"
  service_role_arn       = "${aws_iam_role.example.arn}"

  auto_rollback_configuration {
    enabled = true
    events  = ["DEPLOYMENT_FAILURE"]
  }

  blue_green_deployment_config {
    deployment_ready_option {
      action_on_timeout = "CONTINUE_DEPLOYMENT"
    }

    terminate_blue_instances_on_deployment_success {
      action                           = "TERMINATE"
      termination_wait_time_in_minutes = 5
    }
  }

  deployment_style {
    deployment_option = "WITH_TRAFFIC_CONTROL"
    deployment_type   = "BLUE_GREEN"
  }

  ecs_service {
    cluster_name = "${aws_ecs_cluster.example.name}"
    service_name = "${aws_ecs_service.example.name}"
  }

  load_balancer_info {
    target_group_pair_info {
      prod_traffic_route {
        listener_arns = ["${aws_lb_listener.example.arn}"]
      }

      target_group {
        name = "${aws_lb_target_group.blue.name}"
      }

      target_group {
        name = "${aws_lb_target_group.green.name}"
      }
    }
  }
}
```

### Blue Green Deployments with Servers and Classic ELB

```hcl
resource "aws_codedeploy_app" "example" {
  name = "example-app"
}

resource "aws_codedeploy_deployment_group" "example" {
  app_name              = "${aws_codedeploy_app.example.name}"
  deployment_group_name = "example-group"
  service_role_arn      = "${aws_iam_role.example.arn}"

  deployment_style {
    deployment_option = "WITH_TRAFFIC_CONTROL"
    deployment_type   = "BLUE_GREEN"
  }

  load_balancer_info {
    elb_info {
      name = "${aws_elb.example.name}"
    }
  }

  blue_green_deployment_config {
    deployment_ready_option {
      action_on_timeout    = "STOP_DEPLOYMENT"
      wait_time_in_minutes = 60
    }

    green_fleet_provisioning_option {
      action = "DISCOVER_EXISTING"
    }

    terminate_blue_instances_on_deployment_success {
      action = "KEEP_ALIVE"
    }
  }
}
```

## Argument Reference

The following arguments are supported:

* `app_name` - (Required) The name of the application.
* `deployment_group_name` - (Required) The name of the deployment group.
* `service_role_arn` - (Required) The service role ARN that allows deployments.
* `alarm_configuration` - (Optional) Configuration block of alarms associated with the deployment group (documented below).
* `auto_rollback_configuration` - (Optional) Configuration block of the automatic rollback configuration associated with the deployment group (documented below).
* `autoscaling_groups` - (Optional) Autoscaling groups associated with the deployment group.
* `blue_green_deployment_config` - (Optional) Configuration block of the blue/green deployment options for a deployment group (documented below).
* `deployment_config_name` - (Optional) The name of the group's deployment config. The default is "CodeDeployDefault.OneAtATime".
* `deployment_style` - (Optional) Configuration block of the type of deployment, either in-place or blue/green, you want to run and whether to route deployment traffic behind a load balancer (documented below).
* `ec2_tag_filter` - (Optional) Tag filters associated with the deployment group. See the AWS docs for details.
* `ec2_tag_set` - (Optional) Configuration block(s) of Tag filters associated with the deployment group, which are also referred to as tag groups (documented below). See the AWS docs for details.
* `ecs_service` - (Optional) Configuration block(s) of the ECS services for a deployment group (documented below).
* `load_balancer_info` - (Optional) Single configuration block of the load balancer to use in a blue/green deployment (documented below).
* `on_premises_instance_tag_filter` - (Optional) On premise tag filters associated with the group. See the AWS docs for details.
* `trigger_configuration` - (Optional) Configuration block(s) of the triggers for the deployment group (documented below).

### alarm_configuration Argument Reference

You can configure a deployment to stop when a **CloudWatch** alarm detects that a metric has fallen below or exceeded a defined threshold. `alarm_configuration` supports the following:

* `alarms` - (Optional) A list of alarms configured for the deployment group. _A maximum of 10 alarms can be added to a deployment group_.
* `enabled` - (Optional) Indicates whether the alarm configuration is enabled. This option is useful when you want to temporarily deactivate alarm monitoring for a deployment group without having to add the same alarms again later.
* `ignore_poll_alarm_failure` - (Optional) Indicates whether a deployment should continue if information about the current state of alarms cannot be retrieved from CloudWatch. The default value is `false`.
  * `true`: The deployment will proceed even if alarm status information can't be retrieved.
  * `false`: The deployment will stop if alarm status information can't be retrieved.

_Only one `alarm_configuration` is allowed_.

### auto_rollback_configuration Argument Reference

You can configure a deployment group to automatically rollback when a deployment fails or when a monitoring threshold you specify is met. In this case, the last known good version of an application revision is deployed. `auto_rollback_configuration` supports the following:

* `enabled` - (Optional) Indicates whether a defined automatic rollback configuration is currently enabled for this Deployment Group. If you enable automatic rollback, you must specify at least one event type.
* `events` - (Optional) The event type or types that trigger a rollback. Supported types are `DEPLOYMENT_FAILURE` and `DEPLOYMENT_STOP_ON_ALARM`.

_Only one `auto_rollback_ configuration` is allowed_.

### blue_green_deployment_config Argument Reference

You can configure options for a blue/green deployment. `blue_green_deployment_config` supports the following:

* `deployment_ready_option` - (Optional) Information about the action to take when newly provisioned instances are ready to receive traffic in a blue/green deployment (documented below).
* `green_fleet_provisioning_option` - (Optional) Information about how instances are provisioned for a replacement environment in a blue/green deployment (documented below).
* `terminate_blue_instances_on_deployment_success` - (Optional) Information about whether to terminate instances in the original fleet during a blue/green deployment (documented below).

_Only one `blue_green_deployment_config` is allowed_.

You can configure how traffic is rerouted to instances in a replacement environment in a blue/green deployment. `deployment_ready_option` supports the following:

* `action_on_timeout` - (Optional) When to reroute traffic from an original environment to a replacement environment in a blue/green deployment.
  * `CONTINUE_DEPLOYMENT`: Register new instances with the load balancer immediately after the new application revision is installed on the instances in the replacement environment.
  * `STOP_DEPLOYMENT`: Do not register new instances with load balancer unless traffic is rerouted manually. If traffic is not rerouted manually before the end of the specified wait period, the deployment status is changed to Stopped.
* `wait_time_in_minutes` - (Optional) The number of minutes to wait before the status of a blue/green deployment changed to Stopped if rerouting is not started manually. Applies only to the `STOP_DEPLOYMENT` option for `action_on_timeout`.

You can configure how instances will be added to the replacement environment in a blue/green deployment. `green_fleet_provisioning_option` supports the following:

* `action` - (Optional) The method used to add instances to a replacement environment.
  * `DISCOVER_EXISTING`: Use instances that already exist or will be created manually.
  * `COPY_AUTO_SCALING_GROUP`: Use settings from a specified **Auto Scaling** group to define and create instances in a new Auto Scaling group. _Exactly one Auto Scaling group must be specified_ when selecting `COPY_AUTO_SCALING_GROUP`. Use `autoscaling_groups` to specify the Auto Scaling group.

You can configure how instances in the original environment are terminated when a blue/green deployment is successful. `terminate_blue_instances_on_deployment_success` supports the following:

* `action` - (Optional) The action to take on instances in the original environment after a successful blue/green deployment.
  * `TERMINATE`: Instances are terminated after a specified wait time.
  * `KEEP_ALIVE`: Instances are left running after they are deregistered from the load balancer and removed from the deployment group.
* `termination_wait_time_in_minutes` - (Optional) The number of minutes to wait after a successful blue/green deployment before terminating instances from the original environment.

### deployment_style Argument Reference

You can configure the type of deployment, either in-place or blue/green, you want to run and whether to route deployment traffic behind a load balancer. `deployment_style` supports the following:

* `deployment_option` - (Optional) Indicates whether to route deployment traffic behind a load balancer. Valid Values are `WITH_TRAFFIC_CONTROL` or `WITHOUT_TRAFFIC_CONTROL`.
* `deployment_type` - (Optional) Indicates whether to run an in-place deployment or a blue/green deployment. Valid Values are `IN_PLACE` or `BLUE_GREEN`.

_Only one `deployment_style` is allowed_.

### ec2_tag_filter Argument Reference

The `ec2_tag_filter` configuration block supports the following:

* `key` - (Optional) The key of the tag filter.
* `type` - (Optional) The type of the tag filter, either `KEY_ONLY`, `VALUE_ONLY`, or `KEY_AND_VALUE`.
* `value` - (Optional) The value of the tag filter.

Multiple occurrences of `ec2_tag_filter` are allowed, where any instance that matches to at least one of the tag filters is selected.

### ec2_tag_set Argument Reference

You can form a tag group by putting a set of tag filters into `ec2_tag_set`. If multiple tag groups are specified, any instance that matches to at least one tag filter of every tag group is selected.

### load_balancer_info Argument Reference

You can configure the **Load Balancer** to use in a deployment. `load_balancer_info` supports the following:

* `elb_info` - (Optional) The Classic Elastic Load Balancer to use in a deployment. Conflicts with `target_group_info` and `target_group_pair_info`.
* `target_group_info` - (Optional) The (Application/Network Load Balancer) target group to use in a deployment. Conflicts with `elb_info` and `target_group_pair_info`.
* `target_group_pair_info` - (Optional) The (Application/Network Load Balancer) target group pair to use in a deployment. Conflicts with `elb_info` and `target_group_info`.

#### load_balancer_info elb_info Argument Reference

The `elb_info` configuration block supports the following:

* `name` - (Optional) The name of the load balancer that will be used to route traffic from original instances to replacement instances in a blue/green deployment. For in-place deployments, the name of the load balancer that instances are deregistered from so they are not serving traffic during a deployment, and then re-registered with after the deployment completes.

#### load_balancer_info target_group_info Argument Reference

The `target_group_info` configuration block supports the following:

* `name` - (Optional) The name of the target group that instances in the original environment are deregistered from, and instances in the replacement environment registered with. For in-place deployments, the name of the target group that instances are deregistered from, so they are not serving traffic during a deployment, and then re-registered with after the deployment completes.

#### load_balancer_info target_group_pair_info Argument Reference

The `target_group_pair_info` configuration block supports the following:

* `prod_traffic_route` - (Required) Configuration block for the production traffic route (documented below).
* `target_group` - (Required) Configuration blocks for a target group within a target group pair (documented below).
* `test_traffic_route` - (Optional) Configuration block for the test traffic route (documented below).

##### load_balancer_info target_group_pair_info prod_traffic_route Argument Reference

The `prod_traffic_route` configuration block supports the following:

* `listener_arns` - (Required) List of Amazon Resource Names (ARNs) of the load balancer listeners.

##### load_balancer_info target_group_pair_info target_group Argument Reference

The `target_group` configuration block supports the following:

* `name` - (Required) Name of the target group.

##### load_balancer_info target_group_pair_info test_traffic_route Argument Reference

The `test_traffic_route` configuration block supports the following:

* `listener_arns` - (Required) List of Amazon Resource Names (ARNs) of the load balancer listeners.

### on_premises_tag_filter Argument Reference

The `on_premises_tag_filter` configuration block supports the following:

* `key` - (Optional) The key of the tag filter.
* `type` - (Optional) The type of the tag filter, either `KEY_ONLY`, `VALUE_ONLY`, or `KEY_AND_VALUE`.
* `value` - (Optional) The value of the tag filter.

### trigger_configuration Argument Reference

Add triggers to a Deployment Group to receive notifications about events related to deployments or instances in the group. Notifications are sent to subscribers of the **SNS** topic associated with the trigger. _CodeDeploy must have permission to publish to the topic from this deployment group_. `trigger_configuration` supports the following:

* `trigger_events` - (Required) The event type or types for which notifications are triggered. Some values that are supported: `DeploymentStart`, `DeploymentSuccess`, `DeploymentFailure`, `DeploymentStop`, `DeploymentRollback`, `InstanceStart`, `InstanceSuccess`, `InstanceFailure`.  See [the CodeDeploy documentation][1] for all possible values.
* `trigger_name` - (Required) The name of the notification trigger.
* `trigger_target_arn` - (Required) The ARN of the SNS topic through which notifications are sent.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - Application name and deployment group name.

## Import

CodeDeploy Deployment Groups can be imported by their `app_name`, a colon, and `deployment_group_name`, e.g.

```
$ terraform import aws_codedeploy_deployment_group.example my-application:my-deployment-group
```

[1]: http://docs.aws.amazon.com/codedeploy/latest/userguide/monitoring-sns-event-notifications-create-trigger.html
