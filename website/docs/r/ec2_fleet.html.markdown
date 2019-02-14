---
layout: "aws"
page_title: "AWS: aws_ec2_fleet"
sidebar_current: "docs-aws-resource-ec2-fleet"
description: |-
  Provides a resource to manage EC2 Fleets
---

# aws_ec2_fleet

Provides a resource to manage EC2 Fleets.

## Example Usage

```hcl
resource "aws_ec2_fleet" "example" {
  launch_template_config {
    launch_template_specification {
      launch_template_id = "${aws_launch_template.example.id}"
      version            = "${aws_launch_template.example.latest_version}"
    }
  }

  target_capacity_specification {
    default_target_capacity_type = "spot"
    total_target_capacity        = 5
  }
}
```

## Argument Reference

The following arguments are supported:

* `launch_template_config` - (Required) Nested argument containing EC2 Launch Template configurations. Defined below.
* `target_capacity_specification` - (Required) Nested argument containing target capacity configurations. Defined below.
* `excess_capacity_termination_policy` - (Optional) Whether running instances should be terminated if the total target capacity of the EC2 Fleet is decreased below the current size of the EC2. Valid values: `no-termination`, `termination`. Defaults to `termination`.
* `on_demand_options` - (Optional) Nested argument containing On-Demand configurations. Defined below.
* `replace_unhealthy_instances` - (Optional) Whether EC2 Fleet should replace unhealthy instances. Defaults to `false`.
* `spot_options` - (Optional) Nested argument containing Spot configurations. Defined below.
* `tags` - (Optional) Map of Fleet tags. To tag instances at launch, specify the tags in the Launch Template.
* `terminate_instances` - (Optional) Whether to terminate instances for an EC2 Fleet if it is deleted successfully. Defaults to `false`.
* `terminate_instances_with_expiration` - (Optional) Whether running instances should be terminated when the EC2 Fleet expires. Defaults to `false`.
* `type` - (Optional) The type of request. Indicates whether the EC2 Fleet only requests the target capacity, or also attempts to maintain it. Valid values: `maintain`, `request`. Defaults to `maintain`.

### launch_template_config

* `launch_template_specification` - (Required) Nested argument containing EC2 Launch Template to use. Defined below.
* `override` - (Optional) Nested argument(s) containing parameters to override the same parameters in the Launch Template. Defined below.

#### launch_template_specification

~> *NOTE:* Either `launch_template_id` or `launch_template_name` must be specified.

* `version` - (Required) Version number of the launch template.
* `launch_template_id` - (Optional) ID of the launch template.
* `launch_template_name` - (Optional) Name of the launch template.

#### override

Example:

```hcl
resource "aws_ec2_fleet" "example" {
  # ... other configuration ...

  launch_template_config {
    # ... other configuration ...

    override {
      instance_type     = "m4.xlarge"
      weighted_capacity = 1
    }

    override {
      instance_type     = "m4.2xlarge"
      weighted_capacity = 2
    }
  }
}
```

* `availability_zone` - (Optional) Availability Zone in which to launch the instances.
* `instance_type` - (Optional) Instance type.
* `max_price` - (Optional) Maximum price per unit hour that you are willing to pay for a Spot Instance.
* `priority` - (Optional) Priority for the launch template override. If `on_demand_options` `allocation_strategy` is set to `prioritized`, EC2 Fleet uses priority to determine which launch template override to use first in fulfilling On-Demand capacity. The highest priority is launched first. The lower the number, the higher the priority. If no number is set, the launch template override has the lowest priority. Valid values are whole numbers starting at 0.
* `subnet_id` - (Optional) ID of the subnet in which to launch the instances.
* `weighted_capacity` - (Optional) Number of units provided by the specified instance type.

### on_demand_options

* `allocation_strategy` - (Optional) The order of the launch template overrides to use in fulfilling On-Demand capacity. Valid values: `lowestPrice`, `prioritized`. Default: `lowestPrice`.

### spot_options

* `allocation_strategy` - (Optional) How to allocate the target capacity across the Spot pools. Valid values: `diversified`, `lowestPrice`. Default: `lowestPrice`.
* `instance_interruption_behavior` - (Optional) Behavior when a Spot Instance is interrupted. Valid values: `hibernate`, `stop`, `terminate`. Default: `terminate`.
* `instance_pools_to_use_count` - (Optional) Number of Spot pools across which to allocate your target Spot capacity. Valid only when Spot `allocation_strategy` is set to `lowestPrice`. Default: `1`.

### target_capacity_specification

* `default_target_capacity_type` - (Required) Default target capacity type. Valid values: `on-demand`, `spot`.
* `total_target_capacity` - (Required) The number of units to request, filled using `default_target_capacity_type`.
* `on_demand_target_capacity` - (Optional) The number of On-Demand units to request.
* `spot_target_capacity` - (Optional) The number of Spot units to request.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - Fleet identifier

## Timeouts

`aws_ec2_fleet` provides the following [Timeouts](/docs/configuration/resources.html#timeouts) configuration options:

* `create` - (Default `10m`) How long to wait for a fleet to be active.
* `update` - (Default `10m`) How long to wait for a fleet to be modified.
* `delete` - (Default `10m`) How long to wait for a fleet to be deleted. If `terminate_instances` is `true`, how long to wait for instances to terminate.

## Import

`aws_ec2_fleet` can be imported by using the Fleet identifier, e.g.

```
$ terraform import aws_ec2_fleet.example fleet-b9b55d27-c5fc-41ac-a6f3-48fcc91f080c
```
