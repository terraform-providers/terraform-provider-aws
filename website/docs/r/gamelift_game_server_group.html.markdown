---
subcategory: "Gamelift"
layout: "aws"
page_title: "AWS: aws_gamelift_game_server_group"
description: |-
  Provides a Gamelift Game Server Group resource.
---

# Resource: aws_gamelift_game_server_group

Provides an Gamelift Game Server Group resource.

## Example Usage

Basic usage:

```hcl
resource "aws_gamelift_game_server_group" "example" {
  game_server_group_name = "example"

  instance_definition {
    instance_type = "c5.large"
  }

  instance_definition {
    instance_type = "c5a.large"
  }

  launch_template {
    id = aws_launch_template.example.id
  }

  max_size = 1
  min_size = 1
  role_arn = aws_iam_role.example.arn

  depends_on = [
    aws_iam_role_policy_attachment.example
  ]
}
```

Full usage:

```hcl
resource "aws_gamelift_game_server_group" "example" {
  auto_scaling_policy {
    estimated_instance_warmup = 60

    target_tracking_configuration {
      target_value = 75
    }
  }

  balancing_strategy            = "SPOT_ONLY"
  game_server_group_name        = "example"
  game_server_protection_policy = "FULL_PROTECTION"

  instance_definition {
    instance_type     = "c5.large"
    weighted_capacity = "1"
  }

  instance_definition {
    instance_type     = "c5.2xlarge"
    weighted_capacity = "2"
  }

  launch_template {
    id      = aws_launch_template.example.id
    version = "1"
  }

  max_size = 1
  min_size = 1
  role_arn = aws_iam_role.example.arn

  tags = {
    Name = "example"
  }

  vpc_subnets = [
    "subnet-12345678",
    "subnet-23456789"
  ]

  depends_on = [
    aws_iam_role_policy_attachment.example
  ]
}
```

### Example IAM Role for Gamelift Game Server Group

```hcl
data "aws_partition" "current" {}

resource "aws_iam_role" "example" {
  assume_role_policy = <<-EOF
    {
      "Version": "2012-10-17",
      "Statement": [
        {
          "Effect": "Allow",
          "Principal": {
            "Service": [
              "autoscaling.amazonaws.com",
              "gamelift.amazonaws.com"
            ]
          },
          "Action": "sts:AssumeRole"
        }
      ]
    }
  EOF

  name = "gamelift-game-server-group-example"
}

resource "aws_iam_role_policy_attachment" "example" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/GameLiftGameServerGroupPolicy"
  role       = aws_iam_role.example.name
}
```

## Argument Reference

The following arguments are supported:

* `balancing_strategy` - (Optional) Indicates how GameLift FleetIQ balances the use of Spot Instances and On-Demand Instances.
  Valid values: `SPOT_ONLY`, `SPOT_PREFERRED`, `ON_DEMAND_ONLY`. Defaults to `SPOT_PREFERRED`.
* `game_server_group_name` - (Required) Name of the game server group.
  This value is used to generate unique ARN identifiers for the EC2 Auto Scaling group and the GameLift FleetIQ game server group.
* `game_server_protection_policy` - (Optional) Indicates whether instances in the game server group are protected from early termination.
  Unprotected instances that have active game servers running might be terminated during a scale-down event,
  causing players to be dropped from the game.
  Protected instances cannot be terminated while there are active game servers running except in the event
  of a forced game server group deletion.
  Valid values: `NO_PROTECTION`, `FULL_PROTECTION`. Defaults to `NO_PROTECTION`.
* `max_size` - (Required) The maximum number of instances allowed in the EC2 Auto Scaling group.
  During automatic scaling events, GameLift FleetIQ and EC2 do not scale up the group above this maximum.
* `min_size` - (Required) The minimum number of instances allowed in the EC2 Auto Scaling group.
  During automatic scaling events, GameLift FleetIQ and EC2 do not scale down the group below this minimum.
* `role_arn` - (Required) ARN for an IAM role that allows Amazon GameLift to access your EC2 Auto Scaling groups.
* `tags` - (Optional) Key-value map of resource tags
* `vpc_subnets` - (Optional) A list of VPC subnets to use with instances in the game server group.
  By default, all GameLift FleetIQ-supported Availability Zones are used.

### `auto_scaling_policy`

Configuration settings to define a scaling policy for the Auto Scaling group that is optimized for game hosting.
The scaling policy uses the metric `PercentUtilizedGameServers` to maintain a buffer of idle game servers that
can immediately accommodate new games and players.

* `estimated_instance_warmup` - (Optional) Length of time, in seconds, it takes for a new instance to start
  new game server processes and register with GameLift FleetIQ.
  Specifying a warm-up time can be useful, particularly with game servers that take a long time to start up,
  because it avoids prematurely starting new instances. Defaults to `60`.

#### `target_tracking_configuration`

Settings for a target-based scaling policy applied to Auto Scaling group.
These settings are used to create a target-based policy that tracks the GameLift FleetIQ metric `PercentUtilizedGameServers`
and specifies a target value for the metric.

* `target_value` - (Required) Desired value to use with a game server group target-based scaling policy.

### `instance_definition`

The EC2 instance types and sizes to use in the Auto Scaling group.
The instance definitions must specify at least two different instance types that are supported by GameLift FleetIQ.

* `instance_type` - (Required) An EC2 instance type.
* `weighted_capacity` - (Optional) Instance weighting that indicates how much this instance type contributes
  to the total capacity of a game server group.
  Instance weights are used by GameLift FleetIQ to calculate the instance type's cost per unit hour and better identify
  the most cost-effective options.

### `launch_template`

The EC2 launch template that contains configuration settings and game server code to be deployed to all instances in the game server group.
You can specify the template using either the template name or ID.

* `id` - (Optional) A unique identifier for an existing EC2 launch template.
* `name` - (Optional) A readable identifier for an existing EC2 launch template.
* `version` - (Optional) The version of the EC2 launch template to use. If none is set, the default is the first version created.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The name of the Gamelift Game Server Group.
* `arn` - The ARN of the Gamelift Game Server Group.
* `auto_scaling_group_arn` - The ARN of the created EC2 Auto Scaling group.

## Timeouts

`aws_gamelift_game_server_group` provides the following
[Timeouts](https://www.terraform.io/docs/configuration/blocks/resources/syntax.html#operation-timeouts) configuration options:

* `create` - (Default `10 minutes`) How long to wait for the Gamelift Game Server Group to be created.
* `delete` - (Default `30 minutes`) How long to wait for the Gamelift Game Server Group to be deleted.

## Import

Gamelift Game Server Group can be imported using the `name`, e.g.

```
$ terraform import aws_gamelift_game_server_group.example example
```
