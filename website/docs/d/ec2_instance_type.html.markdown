---
subcategory: "EC2"
layout: "aws"
page_title: "AWS: aws_ec2_instance_type"
description: |-
  Information about single EC2 Instance Type.
---


# Data Source: aws_ec2_instance_type

Get characteristics for a single EC2 Instance Type.

## Example Usage

```hcl
data "aws_ec2_instance_type" "example" {
  instance_type = "t2.micro"
}

```

## Argument Reference

The following argument is supported:

* `instance_type` - (Required) Instance

## Attribute Reference

In addition to the argument above, the following attributes are exported:

~> **NOTE:** Not all attributes are set for every instance type.

* `auto_recovery_supported` - `true` if auto recovery is supported.
* `bare_metal` - `true` if it is a bare metal instance type.
* `burstable_performance_supported` - `true` if the instance type is a burstable performance instance type.
* `current_generation` - `true`  if the instance type is a current generation.
* `dedicated_hosts_supported` - `true` if Dedicated Hosts are supported on the instance type.
* `default_cores` - The default number of cores for the instance type.
* `default_threads_per_core` - The  default  number of threads per core for the instance type.
* `default_vcpus` - The default number of vCPUs for the instance type.
* `ebs_encryption_support` - Indicates whether Amazon EBS encryption is supported.
* `ebs_nvme_support` - Indicates whether non-volatile memory express (NVMe) is supported.
* `ebs_optimized_support` - Indicates that the instance type is Amazon EBS-optimized.
* `ebs_performance_baseline_bandwidth` - The baseline bandwidth performance for an EBS-optimized instance type, in Mbps.
* `ebs_performance_baseline_iops` - The baseline input/output storage operations per seconds for an EBS-optimized instance type.
* `ebs_performance_baseline_throughput` - The baseline throughput performance for an EBS-optimized instance type, in MBps.
* `ebs_performance_maximum_bandwidth` - The maximum bandwidth performance for an EBS-optimized instance type, in Mbps.
* `ebs_performance_maximum_iops` - The maximum input/output storage operations per second for an EBS-optimized instance type.
* `ebs_performance_maximum_throughput` - The maximum throughput performance for an EBS-optimized instance type, in MBps.
* `efa_supported` - Indicates whether Elastic Fabric Adapter (EFA) is supported.
* `ena_support` - Indicates whether Elastic Network Adapter (ENA) is supported.
* `fpgas` - Describes the FPGA accelerator settings for the instance type.
    * `fpgas.#.count` - The count of FPGA accelerators for the instance type.
    * `fpgas.#.manufacturer` - The manufacturer of the FPGA accelerator.
    * `fpgas.#.memory_size` - The size (in MiB) for the memory available to the FPGA accelerator.
    * `fpgas.#.name` - The name of the FPGA accelerator.
* `free_tier_eligible` - `true` if the instance type is eligible for the free tier.
* `gpus` - Describes the GPU accelerators for the instance type.
    * `gpus.#.count` - The number of GPUs for the instance type.
    * `gpus.#.manufacturer` - The manufacturer of the GPU accelerator.
    * `gpus.#.memory_size` - The size (in MiB) for the memory available to the GPU accelerator.
    * `gpus.#.name` - The name of the GPU accelerator.
* `hibernation_supported` - `true` if On-Demand hibernation is supported.
* `hypervisor` - Indicates the hypervisor used for the instance type.
* `inference_accelerators` Describes the Inference accelerators for the instance type.
    * `inference_accelerators.#.count` - The number of Inference accelerators for the instance type.
    * `inference_accelerators.#.manufacturer` - The manufacturer of the Inference accelerator.
    * `inference_accelerators.#.name` - The name of the Inference accelerator.
* `instance_disks` - Describes the disks for the instance type.
    * `instance_disks.#.count` - The number of disks with this configuration.
    * `instance_disks.#.size` - The size of the disk in GB.
    * `instance_disks.#.type` - The type of disk.
* `instance_storage_supported` - `true` if instance storage is supported.
* `ipv6_supported` - `true` if IPv6 is supported.
* `maximum_ipv4_addresses_per_interface` - The maximum number of IPv4 addresses per network interface.
* `maximum_ipv6_addresses_per_interface` - The maximum number of IPv6 addresses per network interface.
* `maximum_network_interfaces` - The maximum number of network interfaces for the instance type.
* `memory_size` - Size of the instance memory, in MiB.
* `network_performance` - Describes the network performance.
* `supported_architectures` - A list of architectures supported by the instance type.
* `supported_placement_strategies` - A list of supported placement groups types.
* `supported_root_device_types` - Indicates the supported root device types.
* `supported_usages_classes` - Indicates whether the instance type is offered for spot or On-Demand.
* `supported_virtualization_types` - The supported virtualization types.
* `sustained_clock_speed` - The speed of the processor, in GHz.
* `total_fpga_memory` - The total memory of all FPGA accelerators for the instance type (in MiB).
* `total_gpu_memory` - The total size of the memory for the GPU accelerators for the instance type (in MiB).
* `total_instance_storage` - The total size of the instance disks, in GB.
* `valid_cores` - List of the valid number of cores that can be configured for the instance type.
* `valid_threads_per_core` - List of the valid number of threads per core that can be configured for the instance type.
