---
subcategory: "EC2"
layout: "aws"
page_title: "AWS: aws_launch_template"
description: |-
  Provides an EC2 launch template resource. Can be used to create instances or auto scaling groups.
---

# Resource: aws_launch_template

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

  cpu_options {
    core_count       = 4
    threads_per_core = 2
  }

  credit_specification {
    cpu_credits = "standard"
  }

  disable_api_termination = true

  ebs_optimized = true

  elastic_gpu_specifications {
    type = "test"
  }

  elastic_inference_accelerator {
    type = "eia1.medium"
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

  metadata_options {
    http_endpoint               = "enabled"
    http_tokens                 = "required"
    http_put_response_hop_limit = 1
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

  user_data = filebase64("${path.module}/example.sh")
}
```

## Argument Reference

The following arguments are supported:

* `name` - The name of the launch template. If you leave this blank, Terraform will auto-generate a unique name.
* `name_prefix` - Creates a unique name beginning with the specified prefix. Conflicts with `name`.
* `description` - Description of the launch template.
* `default_version` - Default Version of the launch template.
* `update_default_version` - Whether to update Default Version each update. Conflicts with `default_version`.
* `block_device_mappings` - Specify volumes to attach to the instance besides the volumes specified by the AMI.
  See [Block Devices](#block-devices) below for details.
* `capacity_reservation_specification` - Targeting for EC2 capacity reservations. See [Capacity Reservation Specification](#capacity-reservation-specification) below for more details.
* `cpu_options` - The CPU options for the instance. See [CPU Options](#cpu-options) below for more details.
* `credit_specification` - Customize the credit specification of the instance. See [Credit
  Specification](#credit-specification) below for more details.
* `disable_api_termination` - If `true`, enables [EC2 Instance
  Termination Protection](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/terminating-instances.html#Using_ChangingDisableAPITermination)
* `ebs_optimized` - If `true`, the launched EC2 instance will be EBS-optimized.
* `elastic_gpu_specifications` - The elastic GPU to attach to the instance. See [Elastic GPU](#elastic-gpu)
  below for more details.
* `elastic_inference_accelerator` - (Optional) Configuration block containing an Elastic Inference Accelerator to attach to the instance. See [Elastic Inference Accelerator](#elastic-inference-accelerator) below for more details.
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
* `metadata_options` - (Optional) Customize the metadata options for the instance. See [Metadata Options](#metadata-options) below for more details.
* `monitoring` - The monitoring option for the instance. See [Monitoring](#monitoring) below for more details.
* `network_interfaces` - Customize network interfaces to be attached at instance boot time. See [Network
  Interfaces](#network-interfaces) below for more details.
* `placement` - The placement of the instance. See [Placement](#placement) below for more details.
* `ram_disk_id` - The ID of the RAM disk.
* `security_group_names` - A list of security group names to associate with. If you are creating Instances in a VPC, use
  `vpc_security_group_ids` instead.
* `vpc_security_group_ids` - A list of security group IDs to associate with.
* `tag_specifications` - The tags to apply to the resources during launch. See [Tag Specifications](#tag-specifications) below for more details.
* `tags` - (Optional) A map of tags to assign to the launch template.
* `user_data` - The Base64-encoded user data to provide when launching the instance.
* `hibernation_options` - The hibernation options for the instance. See [Hibernation Options](#hibernation-options) below for more details.
* `enclave_options` - (Optional) Enable Nitro Enclaves on launched instances. See [Enclave Options](#enclave-options) below for more details.

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

* `delete_on_termination` - Whether the volume should be destroyed on instance termination. Defaults to `false` if not set. See [Preserving Amazon EBS Volumes on Instance Termination](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/terminating-instances.html#preserving-volumes-on-termination) for more information.
* `encrypted` - Enables [EBS encryption](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/EBSEncryption.html)
  on the volume (Default: `false`). Cannot be used with `snapshot_id`.
* `iops` - The amount of provisioned
  [IOPS](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ebs-io-characteristics.html).
  This must be set with a `volume_type` of `"io1/io2"`.
* `kms_key_id` - The ARN of the AWS Key Management Service (AWS KMS) customer master key (CMK) to use when creating the encrypted volume.
 `encrypted` must be set to `true` when this is set.
* `snapshot_id` - The Snapshot ID to mount.
* `throughput` - The throughput to provision for a `gp3` volume, with a maximum of 1,000 MiB/s.
* `volume_size` - The size of the volume in gigabytes.
* `volume_type` - The volume type. Can be `standard`, `gp2`, `gp3`, `io1`, `io2`, `sc1` or `st1` (Default: `gp2`).

### Capacity Reservation Specification

The `capacity_reservation_specification` block supports the following:

* `capacity_reservation_preference` - Indicates the instance's Capacity Reservation preferences. Can be `open` or `none`. (Default `none`).
* `capacity_reservation_target` - Used to target a specific Capacity Reservation:

The `capacity_reservation_target` block supports the following:

* `capacity_reservation_id` - The ID of the Capacity Reservation to target.

### CPU Options

The `cpu_options` block supports the following:

* `core_count` - The number of CPU cores for the instance.
* `threads_per_core` - The number of threads per CPU core. To disable Intel Hyper-Threading Technology for the instance, specify a value of 1.
Otherwise, specify the default value of 2.

Both number of CPU cores and threads per core must be specified. Valid number of CPU cores and threads per core for the instance type can be found in the [CPU Options Documentation](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/instance-optimize-cpu.html?shortFooter=true#cpu-options-supported-instances-values)

### Credit Specification

Credit specification can be applied/modified to the EC2 Instance at any time.

The `credit_specification` block supports the following:

* `cpu_credits` - The credit option for CPU usage. Can be `"standard"` or `"unlimited"`. T3 instances are launched as unlimited by default. T2 instances are launched as standard by default.

### Elastic GPU

Attach an elastic GPU the instance.

The `elastic_gpu_specifications` block supports the following:

* `type` - The [Elastic GPU Type](https://docs.aws.amazon.com/AWSEC2/latest/WindowsGuide/elastic-gpus.html#elastic-gpus-basics)

### Elastic Inference Accelerator

Attach an Elastic Inference Accelerator to the instance. Additional information about Elastic Inference in EC2 can be found in the [EC2 User Guide](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/elastic-inference.html).

The `elastic_inference_accelerator` configuration block supports the following:

* `type` - (Required) Accelerator type.

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

### Metadata Options

The metadata options for the instances.

The `metadata_options` block supports the following:

* `http_endpoint` - (Optional) Whether the metadata service is available. Can be `"enabled"` or `"disabled"`. (Default: `"enabled"`).
* `http_tokens` - (Optional) Whether or not the metadata service requires session tokens, also referred to as _Instance Metadata Service Version 2 (IMDSv2)_. Can be `"optional"` or `"required"`. (Default: `"optional"`).
* `http_put_response_hop_limit` - (Optional) The desired HTTP PUT response hop limit for instance metadata requests. The larger the number, the further instance metadata requests can travel. Can be an integer from `1` to `64`. (Default: `1`).

For more information, see the documentation on the [Instance Metadata Service](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ec2-instance-metadata.html).

### Monitoring

The `monitoring` block supports the following:

* `enabled` - If `true`, the launched EC2 instance will have detailed monitoring enabled.

### Network Interfaces

Attaches one or more [Network Interfaces](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/using-eni.html) to the instance.

Check limitations for autoscaling group in [Creating an Auto Scaling Group Using a Launch Template Guide](https://docs.aws.amazon.com/autoscaling/ec2/userguide/create-asg-launch-template.html#limitations)

Each `network_interfaces` block supports the following:

* `associate_carrier_ip_address` - Associate a Carrier IP address with `eth0` for a new network interface. Use this option when you launch an instance in a Wavelength Zone and want to associate a Carrier IP address with the network interface. Boolean value.
* `associate_public_ip_address` - Associate a public ip address with the network interface.  Boolean value.
* `delete_on_termination` - Whether the network interface should be destroyed on instance termination. Defaults to `false` if not set.
* `description` - Description of the network interface.
* `device_index` - The integer index of the network interface attachment.
* `ipv6_addresses` - One or more specific IPv6 addresses from the IPv6 CIDR block range of your subnet. Conflicts with `ipv6_address_count`
* `ipv6_address_count` - The number of IPv6 addresses to assign to a network interface. Conflicts with `ipv6_addresses`
* `network_interface_id` - The ID of the network interface to attach.
* `private_ip_address` - The primary private IPv4 address.
* `ipv4_address_count` - The number of secondary private IPv4 addresses to assign to a network interface. Conflicts with `ipv4_addresses`
* `ipv4_addresses` - One or more private IPv4 addresses to associate. Conflicts with `ipv4_address_count`
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
* `partition_number` - The number of the partition the instance should launch in. Valid only if the placement group strategy is set to partition.

### Hibernation Options

The `hibernation_options` block supports the following:

* `configured` - If set to `true`, the launched EC2 instance will hibernation enabled.

### Enclave Options

The `enclave_options` block supports the following:

* `enabled` - If set to `true`, Nitro Enclaves will be enabled on the instance.

For more information, see the documentation on [Nitro Enclaves](https://docs.aws.amazon.com/enclaves/latest/user/nitro-enclave.html).

### Tag Specifications

The tags to apply to the resources during launch. You can tag instances, volumes, elastic GPUs and spot instance requests. More information can be found in the [EC2 API documentation](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_LaunchTemplateTagSpecificationRequest.html).

Each `tag_specifications` block supports the following:

* `resource_type` - The type of resource to tag. Valid values are `instance`, `volume`, `elastic-gpu` and `spot-instances-request`.
* `tags` - A map of tags to assign to the resource.


## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - Amazon Resource Name (ARN) of the launch template.
* `id` - The ID of the launch template.
* `latest_version` - The latest version of the launch template.

## Import

Launch Templates can be imported using the `id`, e.g.

```
$ terraform import aws_launch_template.web lt-12345678
```
