---
layout: "aws"
page_title: "AWS: aws_autoscaling_group"
sidebar_current: "docs-aws-datasource-autoscaling-group"
description: |-
  Get information on an Amazon EC2 Autoscaling Group.
---

# Data Source: aws_autoscaling_group

Use this data source to get information on an existing autoscaling group.

## Example Usage

```hcl
data "aws_autoscaling_group" "foo" {
  name = "foo"
}
```

## Argument Reference

* `name` - Specify the exact name of the desired autoscaling group.

## Attributes Reference

~> **NOTE:** Some values are not always set and may not be available for
interpolation.

* `arn` - The Amazon Resource Name (ARN) of the Auto Scaling group.
* `name` - The name of the Auto Scaling group.
* `availability_zones` - One or more Availability Zones for the group.
* `default_cool_down` - The amount of time, in seconds, after a scaling activity completes before another scaling activity can start.
* `desired_capacity` - The desired size of the group.
* `health_check_grace_period` - The amount of time, in seconds, that Amazon EC2 Auto Scaling waits before checking the health status of an EC2 instance that has come into service.
* `health_check_type` - The service to use for the health checks. The valid values are EC2 and ELB.
* `launch_configuration` - The name of the associated launch configuration.
* `load_balancers` - One or more load balancers associated with the group.
* `max_size` - The maximum size of the group.
* `min_size` - The minimum size of the group.
* `placement_group` - The name of the placement group into which to launch your instances, if any. For more information, see Placement Groups (http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/placement-groups.html) in the Amazon Elastic Compute Cloud User Guide.
* `service_linked_role_arn` - The Amazon Resource Name (ARN) of the service-linked role that the Auto Scaling group uses to call other AWS services on your behalf.
* `status` - The current state of the group when DeleteAutoScalingGroup is in progress.
* `target_group_arns` - The Amazon Resource Names (ARN) of the target groups for your load balancer.
* `termination_policies` - The termination policies for the group.
* `vpc_zone_identifier` - VPC ID for the group.
