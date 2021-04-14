---
subcategory: "Application Autoscaling"
layout: "aws"
page_title: "AWS: aws_appautoscaling_scheduled_action"
description: |-
  Provides an Application AutoScaling ScheduledAction resource.
---

# Resource: aws_appautoscaling_scheduled_action

Provides an Application AutoScaling ScheduledAction resource.

## Example Usage

### DynamoDB Table Autoscaling

```terraform
resource "aws_appautoscaling_target" "dynamodb" {
  max_capacity       = 100
  min_capacity       = 5
  resource_id        = "table/tableName"
  scalable_dimension = "dynamodb:table:ReadCapacityUnits"
  service_namespace  = "dynamodb"
}

resource "aws_appautoscaling_scheduled_action" "dynamodb" {
  name               = "dynamodb"
  service_namespace  = aws_appautoscaling_target.dynamodb.service_namespace
  resource_id        = aws_appautoscaling_target.dynamodb.resource_id
  scalable_dimension = aws_appautoscaling_target.dynamodb.scalable_dimension
  schedule           = "at(2006-01-02T15:04:05)"

  scalable_target_action {
    min_capacity = 1
    max_capacity = 200
  }
}
```

### ECS Service Autoscaling

```terraform
resource "aws_appautoscaling_target" "ecs" {
  max_capacity       = 4
  min_capacity       = 1
  resource_id        = "service/clusterName/serviceName"
  scalable_dimension = "ecs:service:DesiredCount"
  service_namespace  = "ecs"
}

resource "aws_appautoscaling_scheduled_action" "ecs" {
  name               = "ecs"
  service_namespace  = aws_appautoscaling_target.ecs.service_namespace
  resource_id        = aws_appautoscaling_target.ecs.resource_id
  scalable_dimension = aws_appautoscaling_target.ecs.scalable_dimension
  schedule           = "at(2006-01-02T15:04:05)"

  scalable_target_action {
    min_capacity = 1
    max_capacity = 10
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the scheduled action.
* `service_namespace` - (Required) The namespace of the AWS service. Documentation can be found in the parameter at: [AWS Application Auto Scaling API Reference](https://docs.aws.amazon.com/ApplicationAutoScaling/latest/APIReference/API_PutScheduledAction.html#ApplicationAutoScaling-PutScheduledAction-request-ServiceNamespace) Example: ecs
* `resource_id` - (Required) The identifier of the resource associated with the scheduled action. Documentation can be found in the parameter at: [AWS Application Auto Scaling API Reference](https://docs.aws.amazon.com/ApplicationAutoScaling/latest/APIReference/API_PutScheduledAction.html#ApplicationAutoScaling-PutScheduledAction-request-ResourceId)
* `scalable_dimension` - (Required) The scalable dimension. Documentation can be found in the parameter at: [AWS Application Auto Scaling API Reference](https://docs.aws.amazon.com/ApplicationAutoScaling/latest/APIReference/API_PutScheduledAction.html#ApplicationAutoScaling-PutScheduledAction-request-ScalableDimension) Example: ecs:service:DesiredCount
* `scalable_target_action` - (Required) The new minimum and maximum capacity. You can set both values or just one. See [below](#scalable-target-action-arguments)
* `schedule` - (Required) The schedule for this action. The following formats are supported: At expressions - at(yyyy-mm-ddThh:mm:ss), Rate expressions - rate(valueunit), Cron expressions - cron(fields). Times for at expressions and cron expressions are evaluated using the time zone configured in `timezone`. Documentation can be found in the parameter at: [AWS Application Auto Scaling API Reference](https://docs.aws.amazon.com/ApplicationAutoScaling/latest/APIReference/API_PutScheduledAction.html#ApplicationAutoScaling-PutScheduledAction-request-Schedule)
* `start_time` - (Optional) The date and time for the scheduled action to start in RFC 3339 format. The timezone is not affected by the setting of `timezone`.
* `end_time` - (Optional) The date and time for the scheduled action to end in RFC 3339 format. The timezone is not affected by the setting of `timezone`.
* `timezone` - (Optional) The time zone used when setting a scheduled action by using an at or cron expression. Does not affect timezone for `start_time` and `end_time`. Valid values are the [canonical names of the IANA time zones supported by Joda-Time](https://www.joda.org/joda-time/timezones.html), such as `Etc/GMT+9` or `Pacific/Tahiti`. Default is `UTC`.

### Scalable Target Action Arguments

* `max_capacity` - (Optional) The maximum capacity. At least one of `max_capacity` or `min_capacity` must be set.
* `min_capacity` - (Optional) The minimum capacity. At least one of `min_capacity` or `max_capacity` must be set.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The Amazon Resource Name (ARN) of the scheduled action.
