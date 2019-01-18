---
layout: "aws"
page_title: "AWS: aws_codedeploy_deployment_config"
sidebar_current: "docs-aws-resource-codedeploy-deployment-config"
description: |-
  Provides a CodeDeploy deployment config.
---

# aws_codedeploy_deployment_config

Provides a CodeDeploy deployment config for an application

## Example Usage

### Server Usage

```hcl
resource "aws_codedeploy_deployment_config" "foo" {
  deployment_config_name = "test-deployment-config"

  minimum_healthy_hosts {
    type  = "HOST_COUNT"
    value = 2
  }
}

resource "aws_codedeploy_deployment_group" "foo" {
  app_name               = "${aws_codedeploy_app.foo_app.name}"
  deployment_group_name  = "bar"
  service_role_arn       = "${aws_iam_role.foo_role.arn}"
  deployment_config_name = "${aws_codedeploy_deployment_config.foo.id}"

  ec2_tag_filter {
    key   = "filterkey"
    type  = "KEY_AND_VALUE"
    value = "filtervalue"
  }

  trigger_configuration {
    trigger_events     = ["DeploymentFailure"]
    trigger_name       = "foo-trigger"
    trigger_target_arn = "foo-topic-arn"
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

### Lambda Usage

```hcl
resource "aws_codedeploy_deployment_config" "foo" {
  deployment_config_name = "test-deployment-config"
  compute_platform = "Lambda"

  traffic_routing_config {
    type = "TimeBasedLinear"

    time_based_linear {
      interval = 10
      percentage = 10
    }
  }
}

resource "aws_codedeploy_deployment_group" "foo" {
  app_name = "${aws_codedeploy_app.foo_app.name}"
  deployment_group_name = "bar"
  service_role_arn = "${aws_iam_role.foo_role.arn}"
  deployment_config_name = "${aws_codedeploy_deployment_config.foo.id}"

  auto_rollback_configuration {
    enabled = true
    events = ["DEPLOYMENT_STOP_ON_ALARM"]
  }

  alarm_configuration {
    alarms = ["my-alarm-name"]
    enabled = true
  }
}
```

## Argument Reference

The following arguments are supported:

* `deployment_config_name` - (Required) The name of the deployment config.
* `compute_platform` - (Optional) The compute platform can be `Server`, `Lambda`, or `ECS`. Default is `Server`.
* `minimum_healthy_hosts` - (Optional) A minimum_healthy_hosts block. Minimum Healthy Hosts are documented below.
* `traffic_routing_config` - (Optional) A traffic_routing_config block. Traffic Routing Config is documented below.

The `minimum_healthy_hosts` block supports the following:

* `type` - (Required) The type can either be `FLEET_PERCENT` or `HOST_COUNT`.
* `value` - (Required) The value when the type is `FLEET_PERCENT` represents the minimum number of healthy instances as
a percentage of the total number of instances in the deployment. If you specify FLEET_PERCENT, at the start of the
deployment, AWS CodeDeploy converts the percentage to the equivalent number of instance and rounds up fractional instances.
When the type is `HOST_COUNT`, the value represents the minimum number of healthy instances as an absolute value.

The `traffic_routing_config` block supports the following:

* `type` - (Optional) Type of traffic routing config. One of `TimeBasedCanary`, `TimeBasedLinear`, `AllAtOnce`.
* `time_based_canary` - (Optional) The time based canary configuration information. If `type` is `TimeBasedLinear`, use `time_based_linear` instead.
* `time_based_linear` - (Optional) The time based linear configuration information. If `type` is `TimeBasedCanary`, use `time_based_canary` instead.

The `time_based_canary` block supports the following:

* `interval` - (Optional) The number of minutes between the first and second traffic shifts of a `TimeBasedCanary` deployment.
* `percentage` - (Optional) The percentage of traffic to shift in the first increment of a `TimeBasedCanary` deployment.

The `time_based_linear` block supports the following:

* `interval` - (Optional) The number of minutes between each incremental traffic shift of a `TimeBasedLinear` deployment.
* `percentage` - (Optional) The percentage of traffic that is shifted at the start of each increment of a `TimeBasedLinear` deployment.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The deployment group's config name.
* `deployment_config_id` - The AWS Assigned deployment config id

## Import

CodeDeploy Deployment Configurations can be imported using the `deployment_config_name`, e.g.

```
$ terraform import aws_codedeploy_app.example my-deployment-config
```
