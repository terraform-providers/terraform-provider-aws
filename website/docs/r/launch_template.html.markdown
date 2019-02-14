---
layout: "aws"
page_title: "AWS: aws_launch_template"
sidebar_current: "docs-aws-resource-launch-template"
description: |-
  Provides an EC2 launch template resource. Can be used to create instances or auto scaling groups.
---

# aws_launch_template

Provides an EC2 launch template resource. Can be used to create instances or auto scaling groups.

## Example Usage

```hcl
resource "aws_launch_template" "foo" {
  name = "foo"

  block_device_mappings {
    device_name = "/dev/sda1"

    ebs {
      volume_size = 20
    }
  }

  capacity_reservation_specification {
    capacity_reservation_preference = "open"
  }

  credit_specification {
    cpu_credits = "standard"
  }

  disable_api_termination = true

  ebs_optimized = true

  elastic_gpu_specifications {
    type = "test"
  }

  iam_instance_profile {
    name = "test"
  }

  image_id = "ami-test"

  instance_initiated_shutdown_behavior = "terminate"

  instance_market_options {
    market_type = "spot"
  }

  instance_type = "t2.micro"

  kernel_id = "test"

  key_name = "test"

  license_specification {
    license_configuration_arn = "arn:aws:license-manager:eu-west-1:123456789012:license-configuration:lic-0123456789abcdef0123456789abcdef"
  }

  monitoring {
    enabled = true
  }

  network_interfaces {
    associate_public_ip_address = true
  }

  placement {
    availability_zone = "us-west-2a"
  }

  ram_disk_id = "test"

  vpc_security_group_ids = ["sg-12345678"]

  tag_specifications {
    resource_type = "instance"

  tags = {
      Name = "test"
    }
  }

  user_data = "${base64encode(...)}"
}
```

## Argument Reference

The following arguments are supported:

* `name` - The name of the launch template. If you leave this blank, Terraform will auto-generate a unique name.
* `name_prefix` - Creates a unique name beginning with the specified prefix. Conflicts with `name`.
* `description` - Description of the launch template.
* `block_device_mappings` - Specify volumes to attach to the instance besides the volumes specified by the AMI.
  See [Block Devices](#block-devices) below for details.
* `capacity_reservation_specification` - Targeting for EC2 capacity reservations. See [Capacity Reservation Specification](#capacity-reservation-specification) below for more details.
* `credit_specification` - Customize the credit specification of the instance. See [Credit
  Specification](#credit-specification) below for more details.
* `disable_api_termination` - If `true`, enables [EC2 Instance
  Termination Protection](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/terminating-instances.html#Using_ChangingDisableAPITermination)
* `ebs_optimized` - If `true`, the launched EC2 instance will be EBS-optimized.
* `elastic_gpu_specifications` - The elastic GPU to attach to the instance. See [Elastic GPU](#elastic-gpu)
  below for more details.
* `iam_instance_profile` - The IAM Instance Profile to launch the instance with. See [Instance Profile](#instance-profile)
  below for more details.
* `image_id` - The AMI from which to launch the instance.
* `instance_initiated_shutdown_behavior` - Shutdown behavior for the instance. Can be `stop` or `terminate`.
  (Default: `stop`).
* `instance_market_options` - The market (purchasing) option for the instance. See [Market Options](#market-options)
  below for details.
* `instance_type` - The type of the instance.
* `kernel_id` - The kernel ID.
* `key_name` - The key name to use for the instance.
* `license_specification` - A list of license specifications to associate with. See [License Specification](#license-specification) below for more details.
* `monitoring` - The monitoring option for the instance. See [Monitoring](#monitoring) below for more details.
* `network_interfaces` - Customize network interfaces to be attached at instance boot time. See [Network
  Interfaces](#network-interfaces) below for more details.
* `placement` - The placement of the instance. See [Placement](#placement) below for more details.
* `ram_disk_id` - The ID of the RAM disk.
* `security_group_names` - A list of security group names to associate with. If you are creating Instances in a VPC, use
  `vpc_security_group_ids` instead.
* `vpc_security_group_ids` - A list of security group IDs to associate with.
* `tag_specifications` - The tags to apply to the resources during launch. See [Tag Specifications](#tag-specifications) below for more details.
* `tags` - (Optional) A mapping of tags to assign to the launch template.
* `user_data` - The Base64-encoded user data to provide when launching the instance.

### Block devices

Configure additional volumes of the instance besides specified by the AMI. It's a good idea to familiarize yourself with
  [AWS's Block Device Mapping docs](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/block-device-mapping-concepts.html)
  to understand the implications of using these attributes.

To find out more information for an existing AMI to override the configuration, such as `device_name`, you can use the [AWS CLI ec2 describe-images command](https://docs.aws.amazon.com/cli/latest/reference/ec2/describe-images.html).

Each `block_device_mappings` supports the following:

* `device_name` - The name of the device to mount.
* `ebs` - Configure EBS volume properties.
* `no_device` - Suppresses the specified device included in the AMI's block device mapping.
* `virtual_name` - The [Instance Store Device
  Name](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/InstanceStorage.html#InstanceStoreDeviceNames)
  (e.g. `"ephemeral0"`).

The `ebs` block supports the following:

* `delete_on_termination` - Whether the volume should be destroyed on instance termination (Default: `true`).
* `encrypted` - Enables [EBS encryption](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/EBSEncryption.html)
  on the volume (Default: `false`). Cannot be used with `snapshot_id`.
* `iops` - The amount of provisioned
  [IOPS](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ebs-io-characteristics.html).
  This must be set with a `volume_type` of `"io1"`.
* `kms_key_id` - AWS Key Management Service (AWS KMS) customer master key (CMK) to use when creating the encrypted volume.
 `encrypted` must be set to `true` when this is set.
* `snapshot_id` - The Snapshot ID to mount.
* `volume_size` - The size of the volume in gigabytes.
* `volume_type` - The type of volume. Can be `"standard"`, `"gp2"`, or `"io1"`. (Default: `"standard"`).

### Capacity Reservation Specification

The `capacity_reservation_specification` block supports the following:

* `capacity_reservation_preference` - Indicates the instance's Capacity Reservation preferences. Can be `open` or `none`. (Default `none`).
* `capacity_reservation_target` - Used to target a specific Capacity Reservation:

The `capacity_reservation_target` block supports the following:

* `capacity_reservation_id` - The ID of the Capacity Reservation to target.

### Credit Specification

Credit specification can be applied/modified to the EC2 Instance at any time.

The `credit_specification` block supports the following:

* `cpu_credits` - The credit option for CPU usage. Can be `"standard"` or `"unlimited"`. T3 instances are launched as unlimited by default. T2 instances are launched as standard by default.

### Elastic GPU

Attach an elastic GPU the instance.

The `elastic_gpu_specifications` block supports the following:

* `type` - The [Elastic GPU Type](https://docs.aws.amazon.com/AWSEC2/latest/WindowsGuide/elastic-gpus.html#elastic-gpus-basics)

### Instance Profile

The [IAM Instance Profile](https://docs.aws.amazon.com/IAM/latest/UserGuide/id_roles_use_switch-role-ec2_instance-profiles.html)
to attach.

The `iam_instance_profile` block supports the following:

* `arn` - The Amazon Resource Name (ARN) of the instance profile.
* `name` - The name of the instance profile.

### License Specification

Associate one of more license configurations.

The `license_specification` block supports the following:

* `license_configuration_arn` - (Required) ARN of the license configuration.

### Market Options

The market (purchasing) option for the instances.

The `instance_market_options` block supports the following:

* `market_type` - The market type. Can be `spot`.
* `spot_options` - The options for [Spot Instance](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/using-spot-instances.html)

The `spot_options` block supports the following:

* `block_duration_minutes` - The required duration in minutes. This value must be a multiple of 60.
* `instance_interruption_behavior` - The behavior when a Spot Instance is interrupted. Can be `hibernate`,
  `stop`, or `terminate`. (Default: `terminate`).
* `max_price` - The maximum hourly price you're willing to pay for the Spot Instances.
* `spot_instance_type` - The Spot Instance request type. Can be `one-time`, or `persistent`.
* `valid_until` - The end date of the request.

### Monitoring

The `monitoring` block supports the following:

* `enabled` - If `true`, the launched EC2 instance will have detailed monitoring enabled.

### Network Interfaces

Attaches one or more [Network Interfaces](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/using-eni.html) to the instance.

Check limitations for autoscaling group in [Creating an Auto Scaling Group Using a Launch Template Guide](https://docs.aws.amazon.com/autoscaling/ec2/userguide/create-asg-launch-template.html#limitations)

Each `network_interfaces` block supports the following:

* `associate_public_ip_address` - Associate a public ip address with the network interface.  Boolean value.
* `delete_on_termination` - Whether the network interface should be destroyed on instance termination.
* `description` - Description of the network interface.
* `device_index` - The integer index of the network interface attachment.
* `ipv6_addresses` - One or more specific IPv6 addresses from the IPv6 CIDR block range of your subnet. Conflicts with `ipv6_address_count`
* `ipv6_address_count` - The number of IPv6 addresses to assign to a network interface. Conflicts with `ipv6_addresses`
* `network_interface_id` - The ID of the network interface to attach.
* `private_ip_address` - The primary private IPv4 address.
* `ipv4_address_count` - The number of secondary private IPv4 addresses to assign to a network interface. Conflicts with `ipv4_address_count`
* `ipv4_addresses` - One or more private IPv4 addresses to associate. Conflicts with `ipv4_addresses`
* `security_groups` - A list of security group IDs to associate.
* `subnet_id` - The VPC Subnet ID to associate.

### Placement

The [Placement Group](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/placement-groups.html) of the instance.

The `placement` block supports the following:

* `affinity` - The affinity setting for an instance on a Dedicated Host.
* `availability_zone` - The Availability Zone for the instance.
* `group_name` - The name of the placement group for the instance.
* `host_id` - The ID of the Dedicated Host for the instance.
* `spread_domain` - Reserved for future use.
* `tenancy` - The tenancy of the instance (if the instance is running in a VPC). Can be `default`, `dedicated`, or `host`.

### Tag Specifications

The tags to apply to the resources during launch. You can tag instances and volumes. More information can be found in the [EC2 API documentation](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_LaunchTemplateTagSpecificationRequest.html).

Each `tag_specifications` block supports the following:

* `resource_type` - The type of resource to tag. Valid values are `instance` and `volume`.
* `tags` - A mapping of tags to assign to the resource.


## Attributes Reference

The following attributes are exported along with all argument references:

* `arn` - Amazon Resource Name (ARN) of the launch template.
* `id` - The ID of the launch template.
* `default_version` - The default version of the launch template.
* `latest_version` - The latest version of the launch template.

## Import

Launch Templates can be imported using the `id`, e.g.

```
$ terraform import aws_launch_template.web lt-12345678
```
