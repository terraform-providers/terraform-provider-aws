---
layout: "aws"
page_title: "AWS: aws_gamelift_fleet"
sidebar_current: "docs-aws-resource-gamelift-fleet"
description: |-
  Provides a Gamelift Fleet resource.
---

# aws_gamelift_fleet

Provides a Gamelift Fleet resource.

## Example Usage

```hcl
resource "aws_gamelift_fleet" "example" {
  build_id          = "${aws_gamelift_build.example.id}"
  ec2_instance_type = "t2.micro"
  name              = "example-fleet-name"

  runtime_configuration {
    server_process {
      concurrent_executions = 1
      launch_path           = "C:\\game\\GomokuServer.exe"
    }
  }
}
```

## Argument Reference

The following arguments are supported:

* `build_id` - (Required) ID of the Gamelift Build to be deployed on the fleet.
* `ec2_instance_type` - (Required) Name of an EC2 instance type. e.g. `t2.micro`
* `name` - (Required) The name of the fleet.
* `description` - (Optional) Human-readable description of the fleet.
* `ec2_inbound_permission` - (Optional) Range of IP addresses and port settings that permit inbound traffic to access server processes running on the fleet. See below.
* `metric_groups` - (Optional) List of names of metric groups to add this fleet to. A metric group tracks metrics across all fleets in the group. Defaults to `default`.
* `new_game_session_protection_policy` - (Optional) Game session protection policy to apply to all instances in this fleet. e.g. `FullProtection`. Defaults to `NoProtection`.
* `resource_creation_limit_policy` - (Optional) Policy that limits the number of game sessions an individual player can create over a span of time for this fleet. See below.
* `runtime_configuration` - (Optional) Instructions for launching server processes on each instance in the fleet. See below.

### Nested Fields

#### `ec2_inbound_permission`

* `from_port` - (Required) Starting value for a range of allowed port numbers.
* `ip_range` - (Required) Range of allowed IP addresses expressed in CIDR notation. e.g. `000.000.000.000/[subnet mask]` or `0.0.0.0/[subnet mask]`.
* `protocol` - (Required) Network communication protocol used by the fleet. e.g. `TCP` or `UDP`
* `to_port` - (Required) Ending value for a range of allowed port numbers. Port numbers are end-inclusive. This value must be higher than `from_port`.

#### `resource_creation_limit_policy`

* `new_game_sessions_per_creator` - (Optional) Maximum number of game sessions that an individual can create during the policy period.
* `policy_period_in_minutes` - (Optional) Time span used in evaluating the resource creation limit policy.

#### `runtime_configuration`

* `game_session_activation_timeout_seconds` - (Optional) Maximum amount of time (in seconds) that a game session can remain in status `ACTIVATING`.
* `max_concurrent_game_session_activations` - (Optional) Maximum number of game sessions with status `ACTIVATING` to allow on an instance simultaneously. 
* `server_process` - (Optional) Collection of server process configurations that describe which server processes to run on each instance in a fleet. See below.

#### `server_process`

* `concurrent_executions` - (Required) Number of server processes using this configuration to run concurrently on an instance.
* `launch_path` - (Required) Location of the server executable in a game build. All game builds are installed on instances at the root : for Windows instances `C:\game`, and for Linux instances `/local/game`.
* `parameters` - (Optional) Optional list of parameters to pass to the server executable on launch.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - Fleet ID.
* `arn` - Fleet ARN.
* `operating_system` - Operating system of the fleet's computing resources.

## Import

Gamelift Fleets cannot be imported at this time.
