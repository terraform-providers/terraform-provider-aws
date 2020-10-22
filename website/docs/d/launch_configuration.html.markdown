---
subcategory: "Autoscaling"
layout: "aws"
page_title: "AWS: aws_launch_configuration"
description: |-
  Provides a Launch Configuration data source.
---

# Data Source: aws_launch_configuration

Provides information about a Launch Configuration.

## Example Usage

```hcl
data "aws_launch_configuration" "ubuntu" {
  name = "test-launch-config"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the launch configuration.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the launch configuration.
* `arn` - The Amazon Resource Name of the launch configuration.
* `name` - The Name of the launch configuration.
* `image_id` - The EC2 Image ID of the instance.
* `instance_type` - The Instance Type of the instance to launch.
* `iam_instance_profile` - The IAM Instance Profile to associate with launched instances.
* `key_name` - The Key Name that should be used for the instance.
* `security_groups` - A list of associated Security Group IDS.
* `associate_public_ip_address` - Whether a Public IP address is associated with the instance.
* `vpc_classic_link_id` - The ID of a ClassicLink-enabled VPC.
* `vpc_classic_link_security_groups` - The IDs of one or more Security Groups for the specified ClassicLink-enabled VPC.
* `user_data` - The User Data of the instance.
* `enable_monitoring` - Whether Detailed Monitoring is Enabled.
* `ebs_optimized` - Whether the launched EC2 instance will be EBS-optimized.
* `root_block_device` - The Root Block Device of the instance.
* `ebs_block_device` - The EBS Block Devices attached to the instance.
* `ephemeral_block_device` - The Ephemeral volumes on the instance.
* `spot_price` - The Price to use for reserving Spot instances.
* `placement_tenancy` - The Tenancy of the instance.

`root_block_device` is exported with the following attributes:

* `delete_on_termination` - Whether the EBS Volume will be deleted on instance termination.
* `encrypted` - Whether the volume is Encrypted.
* `iops` - The provisioned IOPs of the volume.
* `volume_size` - The Size of the volume.
* `volume_type` - The Type of the volume.

`ebs_block_device` is exported with the following attributes:

* `delete_on_termination` - Whether the EBS Volume will be deleted on instance termination.
* `device_name` - The Name of the device.
* `no_device` - Whether the device in the block device mapping of the AMI is suppressed.
* `iops` - The provisioned IOPs of the volume.
* `snapshot_id` - The Snapshot ID of the mount.
* `volume_size` - The Size of the volume.
* `volume_type` - The Type of the volume.
* `encrypted` - Whether the volume is Encrypted.

`ephemeral_block_device` is exported with the following attributes:

* `device_name` - The Name of the device.
* `virtual_name` - The Virtual Name of the device.
