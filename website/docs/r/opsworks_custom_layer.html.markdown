---
layout: "aws"
page_title: "AWS: aws_opsworks_custom_layer"
description: |-
  Provides an OpsWorks custom layer resource.
---

# Resource: aws_opsworks_custom_layer

Provides an OpsWorks custom layer resource.

## Example Usage

```hcl
resource "aws_opsworks_custom_layer" "custlayer" {
  name       = "My Awesome Custom Layer"
  short_name = "awesome"
  stack_id   = "${aws_opsworks_stack.main.id}"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) A human-readable name for the layer.
* `short_name` - (Required) A short, machine-readable name for the layer, which will be used to identify it in the Chef node JSON.
* `stack_id` - (Required) The id of the stack the layer will belong to.
* `auto_assign_elastic_ips` - (Optional) Whether to automatically assign an elastic IP address to the layer's instances.
* `auto_assign_public_ips` - (Optional) For stacks belonging to a VPC, whether to automatically assign a public IP address to each of the layer's instances.
* `custom_instance_profile_arn` - (Optional) The ARN of an IAM profile that will be used for the layer's instances.
* `custom_security_group_ids` - (Optional) Ids for a set of security groups to apply to the layer's instances.
* `auto_healing` - (Optional) Whether to enable auto-healing for the layer.
* `install_updates_on_boot` - (Optional) Whether to install OS and package updates on each instance when it boots.
* `instance_shutdown_timeout` - (Optional) The time, in seconds, that OpsWorks will wait for Chef to complete after triggering the Shutdown event.
* `elastic_load_balancer` - (Optional) Name of an Elastic Load Balancer to attach to this layer
* `drain_elb_on_shutdown` - (Optional) Whether to enable Elastic Load Balancing connection draining.
* `system_packages` - (Optional) Names of a set of system packages to install on the layer's instances.
* `use_ebs_optimized_instances` - (Optional) Whether to use EBS-optimized instances.
* `ebs_volume` - (Optional) `ebs_volume` blocks, as described below, will each create an EBS volume and connect it to the layer's instances.
* `custom_json` - (Optional) Custom JSON attributes to apply to the layer.

The following extra optional arguments, all lists of Chef recipe names, allow
custom Chef recipes to be applied to layer instances at the five different
lifecycle events, if custom cookbooks are enabled on the layer's stack:

* `custom_configure_recipes`
* `custom_deploy_recipes`
* `custom_setup_recipes`
* `custom_shutdown_recipes`
* `custom_undeploy_recipes`

An `ebs_volume` block supports the following arguments:

* `mount_point` - (Required) The path to mount the EBS volume on the layer's instances.
* `size` - (Required) The size of the volume in gigabytes.
* `number_of_disks` - (Required) The number of disks to use for the EBS volume.
* `raid_level` - (Required) The RAID level to use for the volume.
* `type` - (Optional) The type of volume to create. This may be `standard` (the default), `io1` or `gp2`.
* `iops` - (Optional) For PIOPS volumes, the IOPS per disk.

For the autoscaling you have to configure at least one of the threshold options (cpu, mem, load or alarms).
If you don't set at least one you'll receive a API error.
The `autoscaling` parameters are:

* `autoscaling` - (Optional) Enable autoscaling configuration for the layer. Default false. 
* `downscaling_alarms`- (Optional) List of alarms to downscale.
* `downscaling_cpu_threshold` - (Optional) CPU threshold to downscale. Default 20.
* `downscaling_ignore_metrics_time`- (Optional) Number of minutes ignore metrics after downscale. Default 10.
* `downscaling_instance_count` - (Optional) Stop servers in batches of this number on downscale. Default 1
* `downscaling_load_threshold` - (Optional) Load threshold to downscale. Default disabled.
* `downscaling_mem_threshold` - (Optional) Mem threshold to downscale. Default disabled.
* `downscaling_threshold_wait_time` - (Optional) Number of minutes undershot the threshold to downscale. Default 10.
* `upscaling_alarms`- (Optional) List of alarms to upscale.
* `upscaling_cpu_threshold` - (Optional) CPU threshold to upscale. Default 80.
* `upscaling_ignore_metrics_time`- (Optional) Number of minutes ignore metrics after upscale. Default 5.
* `upscaling_instance_count` - (Optional) Start servers in batches of this number on upscale. Default 1.
* `upscaling_load_threshold` - (Optional) Load threshold to upscale. Default disabled.
* `upscaling_mem_threshold` - (Optional) Mem threshold to upscale. Default disabled.
* `upscaling_threshold_wait_time` - (Optional) Number of minutes exceed the threshold to upscale. Default 5.


## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The id of the layer.

## Import

OpsWorks Custom Layers can be imported using the `id`, e.g.

```
$ terraform import aws_opsworks_custom_layer.bar 00000000-0000-0000-0000-000000000000
```
