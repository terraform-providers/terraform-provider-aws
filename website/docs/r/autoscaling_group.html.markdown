---
subcategory: "Autoscaling"
layout: "aws"
page_title: "AWS: aws_autoscaling_group"
description: |-
  Provides an Auto Scaling Group resource.
---

# Resource: aws_autoscaling_group

Provides an Auto Scaling Group resource.

-> **Note:** You must specify either `launch_configuration`, `launch_template`, or `mixed_instances_policy`.

~> **NOTE on Auto Scaling Groups and ASG Attachments:** Terraform currently provides
both a standalone [`aws_autoscaling_attachment`](autoscaling_attachment.html) resource
(describing an ASG attached to an ELB or ALB), and an [`aws_autoscaling_group`](autoscaling_group.html)
with `load_balancers` and `target_group_arns` defined in-line. These two methods are not
mutually-exclusive. If `aws_autoscaling_attachment` resources are used, either alone or with inline
`load_balancers` or `target_group_arns`, the `aws_autoscaling_group` resource must be configured
to ignore changes to the `load_balancers` and `target_group_arns` arguments within a
[`lifecycle` configuration block](https://www.terraform.io/docs/configuration/meta-arguments/lifecycle.html).

## Example Usage

```hcl
resource "aws_placement_group" "test" {
  name     = "test"
  strategy = "cluster"
}

resource "aws_autoscaling_group" "bar" {
  name                      = "foobar3-terraform-test"
  max_size                  = 5
  min_size                  = 2
  health_check_grace_period = 300
  health_check_type         = "ELB"
  desired_capacity          = 4
  force_delete              = true
  placement_group           = aws_placement_group.test.id
  launch_configuration      = aws_launch_configuration.foobar.name
  vpc_zone_identifier       = [aws_subnet.example1.id, aws_subnet.example2.id]

  initial_lifecycle_hook {
    name                 = "foobar"
    default_result       = "CONTINUE"
    heartbeat_timeout    = 2000
    lifecycle_transition = "autoscaling:EC2_INSTANCE_LAUNCHING"

    notification_metadata = <<EOF
{
  "foo": "bar"
}
EOF

    notification_target_arn = "arn:aws:sqs:us-east-1:444455556666:queue1*"
    role_arn                = "arn:aws:iam::123456789012:role/S3Access"
  }

  tag {
    key                 = "foo"
    value               = "bar"
    propagate_at_launch = true
  }

  timeouts {
    delete = "15m"
  }

  tag {
    key                 = "lorem"
    value               = "ipsum"
    propagate_at_launch = false
  }
}
```

### With Latest Version Of Launch Template

```hcl
resource "aws_launch_template" "foobar" {
  name_prefix   = "foobar"
  image_id      = "ami-1a2b3c"
  instance_type = "t2.micro"
}

resource "aws_autoscaling_group" "bar" {
  availability_zones = ["us-east-1a"]
  desired_capacity   = 1
  max_size           = 1
  min_size           = 1

  launch_template {
    id      = aws_launch_template.foobar.id
    version = "$Latest"
  }
}
```

### Mixed Instances Policy

```hcl
resource "aws_launch_template" "example" {
  name_prefix   = "example"
  image_id      = data.aws_ami.example.id
  instance_type = "c5.large"
}

resource "aws_autoscaling_group" "example" {
  availability_zones = ["us-east-1a"]
  desired_capacity   = 1
  max_size           = 1
  min_size           = 1

  mixed_instances_policy {
    launch_template {
      launch_template_specification {
        launch_template_id = aws_launch_template.example.id
      }

      override {
        instance_type     = "c4.large"
        weighted_capacity = "3"
      }

      override {
        instance_type     = "c3.large"
        weighted_capacity = "2"
      }
    }
  }
}
```

### Mixed Instances Policy with Spot Instances and Capacity Rebalance

```hcl
resource "aws_launch_template" "example" {
  name_prefix   = "example"
  image_id      = data.aws_ami.example.id
  instance_type = "c5.large"
}

resource "aws_autoscaling_group" "example" {
  capacity_rebalance  = true
  desired_capacity    = 12
  max_size            = 15
  min_size            = 12
  vpc_zone_identifier = [aws_subnet.example1.id, aws_subnet.example2.id]

  mixed_instances_policy {
    instances_distribution {
      on_demand_base_capacity                  = 0
      on_demand_percentage_above_base_capacity = 25
      spot_allocation_strategy                 = "capacity-optimized"
    }

    launch_template {
      launch_template_specification {
        launch_template_id = aws_launch_template.example.id
      }

      override {
        instance_type     = "c4.large"
        weighted_capacity = "3"
      }

      override {
        instance_type     = "c3.large"
        weighted_capacity = "2"
      }
    }
  }
}
```

### Interpolated tags

```hcl
variable "extra_tags" {
  default = [
    {
      key                 = "Foo"
      value               = "Bar"
      propagate_at_launch = true
    },
    {
      key                 = "Baz"
      value               = "Bam"
      propagate_at_launch = true
    },
  ]
}

resource "aws_autoscaling_group" "bar" {
  name                 = "foobar3-terraform-test"
  max_size             = 5
  min_size             = 2
  launch_configuration = aws_launch_configuration.foobar.name
  vpc_zone_identifier  = [aws_subnet.example1.id, aws_subnet.example2.id]

  tags = concat(
    [
      {
        "key"                 = "interpolation1"
        "value"               = "value3"
        "propagate_at_launch" = true
      },
      {
        "key"                 = "interpolation2"
        "value"               = "value4"
        "propagate_at_launch" = true
      },
    ],
    var.extra_tags,
  )
}
```

### Automatically refresh all instances after the group is updated

```hcl
resource "aws_autoscaling_group" "example" {
  availability_zones = ["us-east-1a"]
  desired_capacity   = 1
  max_size           = 2
  min_size           = 1

  launch_template {
    id      = aws_launch_template.example.id
    version = aws_launch_template.example.latest_version
  }

  tag {
    key                 = "Key"
    value               = "Value"
    propagate_at_launch = true
  }

  instance_refresh {
    strategy = "Rolling"
    preferences {
      min_healthy_percentage = 50
    }
    triggers = ["tag"]
  }
}

data "aws_ami" "example" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn-ami-hvm-*-x86_64-gp2"]
  }
}

resource "aws_launch_template" "example" {
  image_id      = data.aws_ami.example.id
  instance_type = "t3.nano"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Optional) The name of the Auto Scaling Group. By default generated by Terraform.
* `name_prefix` - (Optional) Creates a unique name beginning with the specified
  prefix. Conflicts with `name`.
* `max_size` - (Required) The maximum size of the Auto Scaling Group.
* `min_size` - (Required) The minimum size of the Auto Scaling Group.
    (See also [Waiting for Capacity](#waiting-for-capacity) below.)
* `availability_zones` - (Optional) A list of one or more availability zones for the group. Used for EC2-Classic and default subnets when not specified with `vpc_zone_identifier` argument. Conflicts with `vpc_zone_identifier`.
* `capacity_rebalance` - (Optional) Indicates whether capacity rebalance is enabled. Otherwise, capacity rebalance is disabled.
* `default_cooldown` - (Optional) The amount of time, in seconds, after a scaling activity completes before another scaling activity can start.
* `launch_configuration` - (Optional) The name of the launch configuration to use.
* `launch_template` - (Optional) Nested argument with Launch template specification to use to launch instances. Defined below.
* `mixed_instances_policy` (Optional) Configuration block containing settings to define launch targets for Auto Scaling groups. Defined below.
* `initial_lifecycle_hook` - (Optional) One or more
  [Lifecycle Hooks](http://docs.aws.amazon.com/autoscaling/latest/userguide/lifecycle-hooks.html)
  to attach to the Auto Scaling Group **before** instances are launched. The
  syntax is exactly the same as the separate
  [`aws_autoscaling_lifecycle_hook`](/docs/providers/aws/r/autoscaling_lifecycle_hook.html)
  resource, without the `autoscaling_group_name` attribute. Please note that this will only work when creating
  a new Auto Scaling Group. For all other use-cases, please use `aws_autoscaling_lifecycle_hook` resource.
* `health_check_grace_period` - (Optional, Default: 300) Time (in seconds) after instance comes into service before checking health.
* `health_check_type` - (Optional) "EC2" or "ELB". Controls how health checking is done.
* `desired_capacity` - (Optional) The number of Amazon EC2 instances that
    should be running in the group. (See also [Waiting for
    Capacity](#waiting-for-capacity) below.)
* `force_delete` - (Optional) Allows deleting the Auto Scaling Group without waiting
   for all instances in the pool to terminate.  You can force an Auto Scaling Group to delete
   even if it's in the process of scaling a resource. Normally, Terraform
   drains all the instances before deleting the group.  This bypasses that
   behavior and potentially leaves resources dangling.
* `load_balancers` (Optional) A list of elastic load balancer names to add to the autoscaling
   group names. Only valid for classic load balancers. For ALBs, use `target_group_arns` instead.
* `vpc_zone_identifier` (Optional) A list of subnet IDs to launch resources in. Subnets automatically determine which availability zones the group will reside. Conflicts with `availability_zones`.
* `target_group_arns` (Optional) A set of `aws_alb_target_group` ARNs, for use with Application or Network Load Balancing.
* `termination_policies` (Optional) A list of policies to decide how the instances in the Auto Scaling Group should be terminated. The allowed values are `OldestInstance`, `NewestInstance`, `OldestLaunchConfiguration`, `ClosestToNextInstanceHour`, `OldestLaunchTemplate`, `AllocationStrategy`, `Default`.
* `suspended_processes` - (Optional) A list of processes to suspend for the Auto Scaling Group. The allowed values are `Launch`, `Terminate`, `HealthCheck`, `ReplaceUnhealthy`, `AZRebalance`, `AlarmNotification`, `ScheduledActions`, `AddToLoadBalancer`.
Note that if you suspend either the `Launch` or `Terminate` process types, it can prevent your Auto Scaling Group from functioning properly.
* `tag` (Optional) Configuration block(s) containing resource tags. Conflicts with `tags`. Documented below.
* `tags` (Optional) Set of maps containing resource tags. Conflicts with `tag`. Documented below.
* `placement_group` (Optional) The name of the placement group into which you'll launch your instances, if any.
* `metrics_granularity` - (Optional) The granularity to associate with the metrics to collect. The only valid value is `1Minute`. Default is `1Minute`.
* `enabled_metrics` - (Optional) A list of metrics to collect. The allowed values are `GroupDesiredCapacity`, `GroupInServiceCapacity`, `GroupPendingCapacity`, `GroupMinSize`, `GroupMaxSize`, `GroupInServiceInstances`, `GroupPendingInstances`, `GroupStandbyInstances`, `GroupStandbyCapacity`, `GroupTerminatingCapacity`, `GroupTerminatingInstances`, `GroupTotalCapacity`, `GroupTotalInstances`.
* `wait_for_capacity_timeout` (Default: "10m") A maximum
  [duration](https://golang.org/pkg/time/#ParseDuration) that Terraform should
  wait for ASG instances to be healthy before timing out.  (See also [Waiting
  for Capacity](#waiting-for-capacity) below.) Setting this to "0" causes
  Terraform to skip all Capacity Waiting behavior.
* `min_elb_capacity` - (Optional) Setting this causes Terraform to wait for
  this number of instances from this Auto Scaling Group to show up healthy in the
  ELB only on creation. Updates will not wait on ELB instance number changes.
  (See also [Waiting for Capacity](#waiting-for-capacity) below.)
* `wait_for_elb_capacity` - (Optional) Setting this will cause Terraform to wait
  for exactly this number of healthy instances from this Auto Scaling Group in
  all attached load balancers on both create and update operations. (Takes
  precedence over `min_elb_capacity` behavior.)
  (See also [Waiting for Capacity](#waiting-for-capacity) below.)
* `protect_from_scale_in` (Optional) Allows setting instance protection. The
   Auto Scaling Group will not select instances with this setting for termination
   during scale in events.
* `service_linked_role_arn` (Optional) The ARN of the service-linked role that the ASG will use to call other AWS services
* `max_instance_lifetime` (Optional) The maximum amount of time, in seconds, that an instance can be in service, values must be either equal to 0 or between 604800 and 31536000 seconds.
* `instance_refresh` - (Optional) If this block is configured, start an
   [Instance Refresh](https://docs.aws.amazon.com/autoscaling/ec2/userguide/asg-instance-refresh.html)
   when this Auto Scaling Group is updated. Defined [below](#instance_refresh).

### launch_template

~> **NOTE:** Either `id` or `name` must be specified.

The top-level `launch_template` block supports the following:

* `id` - (Optional) The ID of the launch template. Conflicts with `name`.
* `name` - (Optional) The name of the launch template. Conflicts with `id`.
* `version` - (Optional) Template version. Can be version number, `$Latest`, or `$Default`. (Default: `$Default`).

### mixed_instances_policy

* `instances_distribution` - (Optional) Nested argument containing settings on how to mix on-demand and Spot instances in the Auto Scaling group. Defined below.
* `launch_template` - (Required) Nested argument containing launch template settings along with the overrides to specify multiple instance types and weights. Defined below.

#### mixed_instances_policy instances_distribution

This configuration block supports the following:

* `on_demand_allocation_strategy` - (Optional) Strategy to use when launching on-demand instances. Valid values: `prioritized`. Default: `prioritized`.
* `on_demand_base_capacity` - (Optional) Absolute minimum amount of desired capacity that must be fulfilled by on-demand instances. Default: `0`.
* `on_demand_percentage_above_base_capacity` - (Optional) Percentage split between on-demand and Spot instances above the base on-demand capacity. Default: `0`.
* `spot_allocation_strategy` - (Optional) How to allocate capacity across the Spot pools. Valid values: `lowest-price`, `capacity-optimized`. Default: `lowest-price`.
* `spot_instance_pools` - (Optional) Number of Spot pools per availability zone to allocate capacity. EC2 Auto Scaling selects the cheapest Spot pools and evenly allocates Spot capacity across the number of Spot pools that you specify. Default: `2`.
* `spot_max_price` - (Optional) Maximum price per unit hour that the user is willing to pay for the Spot instances. Default: an empty string which means the on-demand price.

#### mixed_instances_policy launch_template

This configuration block supports the following:

* `launch_template_specification` - (Required) Nested argument defines the Launch Template. Defined below.
* `override` - (Optional) List of nested arguments provides the ability to specify multiple instance types. This will override the same parameter in the launch template. For on-demand instances, Auto Scaling considers the order of preference of instance types to launch based on the order specified in the overrides list. Defined below.

##### mixed_instances_policy launch_template launch_template_specification

~> **NOTE:** Either `launch_template_id` or `launch_template_name` must be specified.

This configuration block supports the following:

* `launch_template_id` - (Optional) The ID of the launch template. Conflicts with `launch_template_name`.
* `launch_template_name` - (Optional) The name of the launch template. Conflicts with `launch_template_id`.
* `version` - (Optional) Template version. Can be version number, `$Latest`, or `$Default`. (Default: `$Default`).

##### mixed_instances_policy launch_template override

This configuration block supports the following:

* `instance_type` - (Optional) Override the instance type in the Launch Template.
* `weighted_capacity` - (Optional) The number of capacity units, which gives the instance type a proportional weight to other instance types.

### tag and tags

The `tag` attribute accepts exactly one tag declaration with the following fields:

* `key` - (Required) Key
* `value` - (Required) Value
* `propagate_at_launch` - (Required) Enables propagation of the tag to
   Amazon EC2 instances launched via this ASG

To declare multiple tags additional `tag` blocks can be specified.
Alternatively the `tags` attributes can be used, which accepts a list of maps containing the above field names as keys and their respective values.
This allows the construction of dynamic lists of tags which is not possible using the single `tag` attribute.
`tag` and `tags` are mutually exclusive, only one of them can be specified.

~> **NOTE:** Other AWS APIs may automatically add special tags to their associated Auto Scaling Group for management purposes, such as ECS Capacity Providers adding the `AmazonECSManaged` tag. These generally should be included in the configuration so Terraform does not attempt to remove them and so if the `min_size` was greater than zero on creation, that these tag(s) are applied to any initial EC2 Instances in the Auto Scaling Group. If these tag(s) were missing in the Auto Scaling Group configuration on creation, affected EC2 Instances missing the tags may require manual intervention of adding the tags to ensure they work properly with the other AWS service.

### instance_refresh

This configuration block supports the following:

* `strategy` - (Required) The strategy to use for instance refresh. The only allowed value is `Rolling`. See [StartInstanceRefresh Action](https://docs.aws.amazon.com/autoscaling/ec2/APIReference/API_StartInstanceRefresh.html#API_StartInstanceRefresh_RequestParameters) for more information.
* `preferences` - (Optional) Override default parameters for Instance Refresh.
    * `instance_warmup` - (Optional) The number of seconds until a newly launched instance is configured and ready to use. Default behavior is to use the Auto Scaling Group's health check grace period.
    * `min_healthy_percentage` - (Optional) The amount of capacity in the Auto Scaling group that must remain healthy during an instance refresh to allow the operation to continue, as a percentage of the desired capacity of the Auto Scaling group. Defaults to `90`.
* `triggers` - (Optional) Set of additional property names that will trigger an Instance Refresh. A refresh will always be triggered by a change in any of `launch_configuration`, `launch_template`, or `mixed_instances_policy`.
  
~> **NOTE:** A refresh is started when any of the following Auto Scaling Group properties change: `launch_configuration`, `launch_template`, `mixed_instances_policy`. Additional properties can be specified in the `triggers` property of `instance_refresh`.

~> **NOTE:** Auto Scaling Groups support up to one active instance refresh at a time. When this resource is updated, any existing refresh is cancelled.

~> **NOTE:** Depending on health check settings and group size, an instance refresh may take a long time or fail. This resource does not wait for the instance refresh to complete.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The Auto Scaling Group id.
* `arn` - The ARN for this Auto Scaling Group
* `availability_zones` - The availability zones of the Auto Scaling Group.
* `min_size` - The minimum size of the Auto Scaling Group
* `max_size` - The maximum size of the Auto Scaling Group
* `default_cooldown` - Time between a scaling activity and the succeeding scaling activity.
* `name` - The name of the Auto Scaling Group
* `health_check_grace_period` - Time after instance comes into service before checking health.
* `health_check_type` - "EC2" or "ELB". Controls how health checking is done.
* `desired_capacity` -The number of Amazon EC2 instances that should be running in the group.
* `launch_configuration` - The launch configuration of the Auto Scaling Group
* `vpc_zone_identifier` (Optional) - The VPC zone identifier

~> **NOTE:** When using `ELB` as the `health_check_type`, `health_check_grace_period` is required.

~> **NOTE:** Terraform has two types of ways you can add lifecycle hooks - via
the `initial_lifecycle_hook` attribute from this resource, or via the separate
[`aws_autoscaling_lifecycle_hook`](/docs/providers/aws/r/autoscaling_lifecycle_hook.html)
resource. `initial_lifecycle_hook` exists here because any lifecycle hooks
added with `aws_autoscaling_lifecycle_hook` will not be added until the
Auto Scaling Group has been created, and depending on your
[capacity](#waiting-for-capacity) settings, after the initial instances have
been launched, creating unintended behavior. If you need hooks to run on all
instances, add them with `initial_lifecycle_hook` here, but take
care to not duplicate these hooks in `aws_autoscaling_lifecycle_hook`.

## Timeouts

`autoscaling_group` provides the following
[Timeouts](https://www.terraform.io/docs/configuration/blocks/resources/syntax.html#operation-timeouts) configuration options:

- `delete` - (Default `10 minutes`) Used for destroying ASG.


## Waiting for Capacity

A newly-created ASG is initially empty and begins to scale to `min_size` (or
`desired_capacity`, if specified) by launching instances using the provided
Launch Configuration. These instances take time to launch and boot.

On ASG Update, changes to these values also take time to result in the target
number of instances providing service.

Terraform provides two mechanisms to help consistently manage ASG scale up
time across dependent resources.

#### Waiting for ASG Capacity

The first is default behavior. Terraform waits after ASG creation for
`min_size` (or `desired_capacity`, if specified) healthy instances to show up
in the ASG before continuing.

If `min_size` or `desired_capacity` are changed in a subsequent update,
Terraform will also wait for the correct number of healthy instances before
continuing.

Terraform considers an instance "healthy" when the ASG reports `HealthStatus:
"Healthy"` and `LifecycleState: "InService"`. See the [AWS AutoScaling
Docs](https://docs.aws.amazon.com/AutoScaling/latest/DeveloperGuide/AutoScalingGroupLifecycle.html)
for more information on an ASG's lifecycle.

Terraform will wait for healthy instances for up to
`wait_for_capacity_timeout`. If ASG creation is taking more than a few minutes,
it's worth investigating for scaling activity errors, which can be caused by
problems with the selected Launch Configuration.

Setting `wait_for_capacity_timeout` to `"0"` disables ASG Capacity waiting.

#### Waiting for ELB Capacity

The second mechanism is optional, and affects ASGs with attached ELBs specified
via the `load_balancers` attribute or with ALBs specified with `target_group_arns`.

The `min_elb_capacity` parameter causes Terraform to wait for at least the
requested number of instances to show up `"InService"` in all attached ELBs
during ASG creation.  It has no effect on ASG updates.

If `wait_for_elb_capacity` is set, Terraform will wait for exactly that number
of Instances to be `"InService"` in all attached ELBs on both creation and
updates.

These parameters can be used to ensure that service is being provided before
Terraform moves on. If new instances don't pass the ELB's health checks for any
reason, the Terraform apply will time out, and the ASG will be marked as
tainted (i.e. marked to be destroyed in a follow up run).

As with ASG Capacity, Terraform will wait for up to `wait_for_capacity_timeout`
for the proper number of instances to be healthy.

#### Troubleshooting Capacity Waiting Timeouts

If ASG creation takes more than a few minutes, this could indicate one of a
number of configuration problems. See the [AWS Docs on Load Balancer
Troubleshooting](https://docs.aws.amazon.com/ElasticLoadBalancing/latest/DeveloperGuide/elb-troubleshooting.html)
for more information.


## Import

Auto Scaling Groups can be imported using the `name`, e.g.

```
$ terraform import aws_autoscaling_group.web web-asg
```
