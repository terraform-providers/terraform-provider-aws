---
subcategory: "Elastic Map Reduce (EMR)"
layout: "aws"
page_title: "AWS: aws_emr_instance_fleet"
description: |-
  Provides an Elastic MapReduce Cluster Instance Fleet
---

# aws_emr_instance_fleet

Provides an Elastic MapReduce Cluster Instance Fleet configuration.
See [Amazon Elastic MapReduce Documentation](https://aws.amazon.com/documentation/emr/) for more information.

~> **NOTE:** At this time, Instance Fleets cannot be destroyed through the API nor
web interface. Instance Fleets are destroyed when the EMR Cluster is destroyed.
Terraform will resize any Instance Fleet to zero when destroying the resource.

## Example Usage

```hcl
resource "aws_emr_instance_fleet" "task" {
  cluster_id = aws_emr_cluster.cluster.id
  instance_type_configs {
    bid_price_as_percentage_of_on_demand_price = 100
    ebs_config {
      size                 = 100
      type                 = "gp2"
      volumes_per_instance = 1
    }
    instance_type     = "m4.xlarge"
    weighted_capacity = 1
  }
  instance_type_configs {
    bid_price_as_percentage_of_on_demand_price = 100
    ebs_config {
      size                 = 100
      type                 = "gp2"
      volumes_per_instance = 1
    }
    instance_type     = "m4.2xlarge"
    weighted_capacity = 2
  }
  launch_specifications {
    spot_specification {
      allocation_strategy      = "capacity-optimized"
      block_duration_minutes   = 0
      timeout_action           = "TERMINATE_CLUSTER"
      timeout_duration_minutes = 10
    }
  }
  name                      = "task fleet"
  target_on_demand_capacity = 1
  target_spot_capacity      = 1
}
```

## Argument Reference

The following arguments are supported:

* `cluster_id` - (Required) ID of the EMR Cluster to attach to. Changing this forces a new resource to be created.
* `instance_type_configs` - (Optional) Configuration block for instance fleet
* `launch_specifications` - (Optional) Configuration block for launch specification
* `target_on_demand_capacity` - (Optional)  The target capacity of On-Demand units for the instance fleet, which determines how many On-Demand instances to provision.
* `target_spot_capacity` - (Optional) The target capacity of Spot units for the instance fleet, which determines how many Spot instances to provision.
* `name` - (Optional) Friendly name given to the instance fleet.

## instance_type_configs Configuration Block

* `bid_price` - (Optional) The bid price for each EC2 Spot instance type as defined by `instance_type`. Expressed in USD. If neither `bid_price` nor `bid_price_as_percentage_of_on_demand_price` is provided, `bid_price_as_percentage_of_on_demand_price` defaults to 100%.
* `bid_price_as_percentage_of_on_demand_price` - (Optional) The bid price, as a percentage of On-Demand price, for each EC2 Spot instance as defined by `instance_type`. Expressed as a number (for example, 20 specifies 20%). If neither `bid_price` nor `bid_price_as_percentage_of_on_demand_price` is provided, `bid_price_as_percentage_of_on_demand_price` defaults to 100%.
* `configurations` - (Optional) A configuration classification that applies when provisioning cluster instances, which can include configurations for applications and software that run on the cluster. List of `configuration` blocks.
* `ebs_config` - (Optional) Configuration block(s) for EBS volumes attached to each instance in the instance group. Detailed below.
* `instance_type` - (Required) An EC2 instance type, such as m4.xlarge.
* `weighted_capacity` - (Optional) The number of units that a provisioned instance of this type provides toward fulfilling the target capacities defined in `aws_emr_instance_fleet`.

## configurations Configuration Block

A configuration classification that applies when provisioning cluster instances, which can include configurations for applications and software that run on the cluster. See [Configuring Applications](https://docs.aws.amazon.com/emr/latest/ReleaseGuide/emr-configure-apps.html).

* `classification` - (Optional) The classification within a configuration.
* `properties` - (Optional) A map of properties specified within a configuration classification

## ebs_config

Attributes for the EBS volumes attached to each EC2 instance in the `master_instance_group` and `core_instance_group` configuration blocks:

* `size` - (Required) The volume size, in gibibytes (GiB).
* `type` - (Required) The volume type. Valid options are `gp2`, `io1`, `standard` and `st1`. See [EBS Volume Types](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/EBSVolumeTypes.html).
* `iops` - (Optional) The number of I/O operations per second (IOPS) that the volume supports
* `volumes_per_instance` - (Optional) The number of EBS volumes with this configuration to attach to each EC2 instance in the instance group (default is 1)

## launch_specifications Configuration Block

* `on_demand_specification` - (Optional) Configuration block for on demand instances launch specifications
* `spot_specification` - (Optional) Configuration block for spot instances launch specifications

## on_demand_specification  Configuration Block

The launch specification for On-Demand instances in the instance fleet, which determines the allocation strategy.
The instance fleet configuration is available only in Amazon EMR versions 4.8.0 and later, excluding 5.0.x versions. On-Demand instances allocation strategy is available in Amazon EMR version 5.12.1 and later.

* `allocation_strategy` - (Required) Specifies the strategy to use in launching On-Demand instance fleets. Currently, the only option is `lowest-price` (the default), which launches the lowest price first.

## spot_specification  Configuration Block

The launch specification for Spot instances in the fleet, which determines the defined duration, provisioning timeout behavior, and allocation strategy.

* `allocation_strategy` - (Required) Specifies the strategy to use in launching Spot instance fleets. Currently, the only option is `capacity-optimized` (the default), which launches instances from Spot instance pools with optimal capacity for the number of instances that are launching.
* `block_duration_minutes` - (Optional) The defined duration for Spot instances (also known as Spot blocks) in minutes. When specified, the Spot instance does not terminate before the defined duration expires, and defined duration pricing for Spot instances applies. Valid values are 60, 120, 180, 240, 300, or 360. The duration period starts as soon as a Spot instance receives its instance ID. At the end of the duration, Amazon EC2 marks the Spot instance for termination and provides a Spot instance termination notice, which gives the instance a two-minute warning before it terminates.
* `timeout_action` - (Required) The action to take when TargetSpotCapacity has not been fulfilled when the TimeoutDurationMinutes has expired; that is, when all Spot instances could not be provisioned within the Spot provisioning timeout. Valid values are `TERMINATE_CLUSTER` and `SWITCH_TO_ON_DEMAND`. SWITCH_TO_ON_DEMAND specifies that if no Spot instances are available, On-Demand Instances should be provisioned to fulfill any remaining Spot capacity.
* `timeout_duration_minutes` - (Required) The spot provisioning timeout period in minutes. If Spot instances are not provisioned within this time period, the TimeOutAction is taken. Minimum value is 5 and maximum value is 1440. The timeout applies only during initial provisioning, when the cluster is first created.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The unique identifier of the instance fleet.

* `provisioned_on_demand_capacity` The number of On-Demand units that have been provisioned for the instance
fleet to fulfill TargetOnDemandCapacity. This provisioned capacity might be less than or greater than TargetOnDemandCapacity.

* `provisioned_spot_capacity` The number of Spot units that have been provisioned for this instance fleet
to fulfill TargetSpotCapacity. This provisioned capacity might be less than or greater than TargetSpotCapacity.

* `status` The current status of the instance fleet.

## Import

EMR Instance Fleet can be imported with the EMR Cluster identifier and Instance Fleet identifier separated by a forward slash (`/`), e.g.

```console
$ terraform import aws_emr_instance_fleet.example j-123456ABCDEF/if-15EK4O09RZLNR
```
