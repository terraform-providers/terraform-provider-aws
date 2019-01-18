---
layout: "aws"
page_title: "AWS: aws_instance"
sidebar_current: "docs-aws-resource-instance"
description: |-
  Provides an EC2 instance resource. This allows instances to be created, updated, and deleted. Instances also support provisioning.
---

# aws_instance

Provides an EC2 instance resource. This allows instances to be created, updated,
and deleted. Instances also support [provisioning](/docs/provisioners/index.html).

## Example Usage

```hcl
# Create a new instance of the latest Ubuntu 14.04 on an
# t2.micro node with an AWS Tag naming it "HelloWorld"
provider "aws" {
  region = "us-west-2"
}

data "aws_ami" "ubuntu" {
  most_recent = true

  filter {
    name   = "name"
    values = ["ubuntu/images/hvm-ssd/ubuntu-trusty-14.04-amd64-server-*"]
  }

  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }

  owners = ["099720109477"] # Canonical
}

resource "aws_instance" "web" {
  ami           = "${data.aws_ami.ubuntu.id}"
  instance_type = "t2.micro"

  tags = {
    Name = "HelloWorld"
  }
}
```

## Argument Reference

The following arguments are supported:

* `ami` - (Required) The AMI to use for the instance.
* `availability_zone` - (Optional) The AZ to start the instance in.
* `placement_group` - (Optional) The Placement Group to start the instance in.
* `tenancy` - (Optional) The tenancy of the instance (if the instance is running in a VPC). An instance with a tenancy of dedicated runs on single-tenant hardware. The host tenancy is not supported for the import-instance command.
* `host_id` - (optional) The Id of a dedicated host that the instance will be assigned to. Use when an instance is to be launched on a specific dedicated host.
* `cpu_core_count` - (Optional) Sets the number of CPU cores for an instance. This option is 
  only supported on creation of instance type that support CPU Options 
  [CPU Cores and Threads Per CPU Core Per Instance Type](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/instance-optimize-cpu.html#cpu-options-supported-instances-values) - specifying this option for unsupported instance types will return an error from the EC2 API.
* `cpu_threads_per_core` - (Optional - has no effect unless `cpu_core_count` is also set)  If set to to 1, hyperthreading is disabled on the launched instance. Defaults to 2 if not set. See [Optimizing CPU Options](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/instance-optimize-cpu.html) for more information.

-> **NOTE:** Changing `cpu_core_count` and/or `cpu_threads_per_core` will cause the resource to be destroyed and re-created.

* `ebs_optimized` - (Optional) If true, the launched EC2 instance will be EBS-optimized.
     Note that if this is not set on an instance type that is optimized by default then
     this will show as disabled but if the instance type is optimized by default then
     there is no need to set this and there is no effect to disabling it.
     See the [EBS Optimized section](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/EBSOptimized.html) of the AWS User Guide for more information.
* `disable_api_termination` - (Optional) If true, enables [EC2 Instance
     Termination Protection](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/terminating-instances.html#Using_ChangingDisableAPITermination)
* `instance_initiated_shutdown_behavior` - (Optional) Shutdown behavior for the
instance. Amazon defaults this to `stop` for EBS-backed instances and
`terminate` for instance-store instances. Cannot be set on instance-store
instances. See [Shutdown Behavior](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/terminating-instances.html#Using_ChangingInstanceInitiatedShutdownBehavior) for more information.
* `instance_type` - (Required) The type of instance to start. Updates to this field will trigger a stop/start of the EC2 instance.
* `key_name` - (Optional) The key name of the Key Pair to use for the instance; which can be managed using [the `aws_key_pair` resource](key_pair.html).

* `get_password_data` - (Optional) If true, wait for password data to become available and retrieve it. Useful for getting the administrator password for instances running Microsoft Windows. The password data is exported to the `password_data` attribute. See [GetPasswordData](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_GetPasswordData.html) for more information.
* `monitoring` - (Optional) If true, the launched EC2 instance will have detailed monitoring enabled. (Available since v0.6.0)
* `security_groups` - (Optional, EC2-Classic and default VPC only) A list of security group names (EC2-Classic) or IDs (default VPC) to associate with.

-> **NOTE:** If you are creating Instances in a VPC, use `vpc_security_group_ids` instead.

* `vpc_security_group_ids` - (Optional, VPC only) A list of security group IDs to associate with.
* `subnet_id` - (Optional) The VPC Subnet ID to launch in.
* `associate_public_ip_address` - (Optional) Associate a public ip address with an instance in a VPC.  Boolean value.
* `private_ip` - (Optional) Private IP address to associate with the
     instance in a VPC.
* `source_dest_check` - (Optional) Controls if traffic is routed to the instance when
  the destination address does not match the instance. Used for NAT or VPNs. Defaults true.
* `user_data` - (Optional) The user data to provide when launching the instance. Do not pass gzip-compressed data via this argument; see `user_data_base64` instead.
* `user_data_base64` - (Optional) Can be used instead of `user_data` to pass base64-encoded binary data directly. Use this instead of `user_data` whenever the value is not a valid UTF-8 string. For example, gzip-encoded user data must be base64-encoded and passed via this argument to avoid corruption.
* `iam_instance_profile` - (Optional) The IAM Instance Profile to
  launch the instance with. Specified as the name of the Instance Profile. Ensure your credentials have the correct permission to assign the instance profile according to the [EC2 documentation](http://docs.aws.amazon.com/IAM/latest/UserGuide/id_roles_use_switch-role-ec2.html#roles-usingrole-ec2instance-permissions), notably `iam:PassRole`.
* `ipv6_address_count`- (Optional) A number of IPv6 addresses to associate with the primary network interface. Amazon EC2 chooses the IPv6 addresses from the range of your subnet.
* `ipv6_addresses` - (Optional) Specify one or more IPv6 addresses from the range of the subnet to associate with the primary network interface
* `tags` - (Optional) A mapping of tags to assign to the resource.
* `volume_tags` - (Optional) A mapping of tags to assign to the devices created by the instance at launch time.
* `root_block_device` - (Optional) Customize details about the root block
  device of the instance. See [Block Devices](#block-devices) below for details.
* `ebs_block_device` - (Optional) Additional EBS block devices to attach to the
  instance.  See [Block Devices](#block-devices) below for details.
* `ephemeral_block_device` - (Optional) Customize Ephemeral (also known as
  "Instance Store") volumes on the instance. See [Block Devices](#block-devices) below for details.
* `network_interface` - (Optional) Customize network interfaces to be attached at instance boot time. See [Network Interfaces](#network-interfaces) below for more details.
* `credit_specification` - (Optional) Customize the credit specification of the instance. See [Credit Specification](#credit-specification) below for more details.

### Timeouts

The `timeouts` block allows you to specify [timeouts](https://www.terraform.io/docs/configuration/resources.html#timeouts) for certain actions:

* `create` - (Defaults to 10 mins) Used when launching the instance (until it reaches the initial `running` state)
* `update` - (Defaults to 10 mins) Used when stopping and starting the instance when necessary during update - e.g. when changing instance type
* `delete` - (Defaults to 20 mins) Used when terminating the instance

### Block devices

Each of the `*_block_device` attributes controls a portion of the AWS
Instance's "Block Device Mapping". It's a good idea to familiarize yourself with [AWS's Block Device
Mapping docs](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/block-device-mapping-concepts.html)
to understand the implications of using these attributes.

The `root_block_device` mapping supports the following:

* `volume_type` - (Optional) The type of volume. Can be `"standard"`, `"gp2"`, `"io1"`, `"sc1"`, or `"st1"`. (Default: `"standard"`).
* `volume_size` - (Optional) The size of the volume in gibibytes (GiB).
* `iops` - (Optional) The amount of provisioned
  [IOPS](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ebs-io-characteristics.html).
  This is only valid for `volume_type` of `"io1"`, and must be specified if
  using that type
* `delete_on_termination` - (Optional) Whether the volume should be destroyed
  on instance termination (Default: `true`).

Modifying any of the `root_block_device` settings requires resource
replacement.

Each `ebs_block_device` supports the following:

* `device_name` - The name of the device to mount.
* `snapshot_id` - (Optional) The Snapshot ID to mount.
* `volume_type` - (Optional) The type of volume. Can be `"standard"`, `"gp2"`,
  or `"io1"`. (Default: `"standard"`).
* `volume_size` - (Optional) The size of the volume in gibibytes (GiB).
* `iops` - (Optional) The amount of provisioned
  [IOPS](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ebs-io-characteristics.html).
  This must be set with a `volume_type` of `"io1"`.
* `delete_on_termination` - (Optional) Whether the volume should be destroyed
  on instance termination (Default: `true`).
* `encrypted` - (Optional) Enables [EBS
  encryption](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/EBSEncryption.html)
  on the volume (Default: `false`). Cannot be used with `snapshot_id`.

Modifying any `ebs_block_device` currently requires resource replacement.

~> **NOTE on EBS block devices:** If you use `ebs_block_device` on an `aws_instance`, Terraform will assume management over the full set of non-root EBS block devices for the instance, and treats additional block devices as drift. For this reason, `ebs_block_device` cannot be mixed with external `aws_ebs_volume` + `aws_volume_attachment` resources for a given instance.

Each `ephemeral_block_device` supports the following:

* `device_name` - The name of the block device to mount on the instance.
* `virtual_name` - (Optional) The [Instance Store Device
  Name](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/InstanceStorage.html#InstanceStoreDeviceNames)
  (e.g. `"ephemeral0"`).
* `no_device` - (Optional) Suppresses the specified device included in the AMI's block device mapping.

Each AWS Instance type has a different set of Instance Store block devices
available for attachment. AWS [publishes a
list](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/InstanceStorage.html#StorageOnInstanceTypes)
of which ephemeral devices are available on each type. The devices are always
identified by the `virtual_name` in the format `"ephemeral{0..N}"`.

~> **NOTE:** Currently, changes to `*_block_device` configuration of _existing_
resources cannot be automatically detected by Terraform. After making updates
to block device configuration, resource recreation can be manually triggered by
using the [`taint` command](/docs/commands/taint.html).

### Network Interfaces

Each of the `network_interface` blocks attach a network interface to an EC2 Instance during boot time. However, because
the network interface is attached at boot-time, replacing/modifying the network interface **WILL** trigger a recreation
of the EC2 Instance. If you should need at any point to detach/modify/re-attach a network interface to the instance, use
the `aws_network_interface` or `aws_network_interface_attachment` resources instead.

The `network_interface` configuration block _does_, however, allow users to supply their own network interface to be used
as the default network interface on an EC2 Instance, attached at `eth0`.

Each `network_interface` block supports the following:

* `device_index` - (Required) The integer index of the network interface attachment. Limited by instance type.
* `network_interface_id` - (Required) The ID of the network interface to attach.
* `delete_on_termination` - (Optional) Whether or not to delete the network interface on instance termination. Defaults to `false`. Currently, the only valid value is `false`, as this is only supported when creating new network interfaces when launching an instance.

### Credit Specification

~> **NOTE:** Removing this configuration on existing instances will only stop managing it. It will not change the configuration back to the default for the instance type.

Credit specification can be applied/modified to the EC2 Instance at any time.

The `credit_specification` block supports the following:

* `cpu_credits` - (Optional) The credit option for CPU usage. Can be `"standard"` or `"unlimited"`. T3 instances are launched as unlimited by default. T2 instances are launched as standard by default.

### Example

```hcl
resource "aws_vpc" "my_vpc" {
  cidr_block = "172.16.0.0/16"

  tags = {
    Name = "tf-example"
  }
}

resource "aws_subnet" "my_subnet" {
  vpc_id            = "${aws_vpc.my_vpc.id}"
  cidr_block        = "172.16.10.0/24"
  availability_zone = "us-west-2a"

  tags = {
    Name = "tf-example"
  }
}

resource "aws_network_interface" "foo" {
  subnet_id   = "${aws_subnet.my_subnet.id}"
  private_ips = ["172.16.10.100"]

  tags = {
    Name = "primary_network_interface"
  }
}

resource "aws_instance" "foo" {
  ami           = "ami-22b9a343" # us-west-2
  instance_type = "t2.micro"

  network_interface {
    network_interface_id = "${aws_network_interface.foo.id}"
    device_index         = 0
  }

  credit_specification {
    cpu_credits = "unlimited"
  }
}
```

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The instance ID.
* `arn` - The ARN of the instance.
* `availability_zone` - The availability zone of the instance.
* `placement_group` - The placement group of the instance.
* `key_name` - The key name of the instance
* `password_data` - Base-64 encoded encrypted password data for the instance.
  Useful for getting the administrator password for instances running Microsoft Windows.
  This attribute is only exported if `get_password_data` is true.
  Note that this encrypted value will be stored in the state file, as with all exported attributes.
  See [GetPasswordData](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_GetPasswordData.html) for more information.
* `public_dns` - The public DNS name assigned to the instance. For EC2-VPC, this
  is only available if you've enabled DNS hostnames for your VPC
* `public_ip` - The public IP address assigned to the instance, if applicable. **NOTE**: If you are using an [`aws_eip`](/docs/providers/aws/r/eip.html) with your instance, you should refer to the EIP's address directly and not use `public_ip`, as this field will change after the EIP is attached.
* `ipv6_addresses` - A list of assigned IPv6 addresses, if any
* `network_interface_id` - The ID of the network interface that was created with the instance.
* `primary_network_interface_id` - The ID of the instance's primary network interface.
* `private_dns` - The private DNS name assigned to the instance. Can only be
  used inside the Amazon EC2, and only available if you've enabled DNS hostnames
  for your VPC
* `private_ip` - The private IP address assigned to the instance
* `security_groups` - The associated security groups.
* `vpc_security_group_ids` - The associated security groups in non-default VPC
* `subnet_id` - The VPC subnet ID.
* `credit_specification` - Credit specification of instance.

For any `root_block_device` and `ebs_block_device` the `volume_id` is exported.
e.g. `aws_instance.web.root_block_device.0.volume_id`

## Import

Instances can be imported using the `id`, e.g.

```
$ terraform import aws_instance.web i-12345678
```
