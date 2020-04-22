---
subcategory: "EC2"
layout: "aws"
page_title: "AWS: aws_dedicated_host"
description: |-
  Get information on a EC2 host.
---

# Data Source: aws_dedicated_host

Use this data source to get information about the host when allocating an EC2 Dedicated Host.

## Example Usage

```hcl
# Create a new host with instance type of c5.18xlarge with Auto Placement 
# and Host Recovery enabled. 
provider "aws" {
  region = "us-west-2"
}

resource "aws_dedicated_host" "test" {
  instance_type     = "c5.18xlarge"
  availability_zone = "us-west-1a"
  host_recovery     = "on"
  auto_placement    = "on"
}

data "aws_dedicated_host" "test_data" {
  host_id = "${aws_dedicated_host.test.id}"
}
```

## Argument Reference

The following arguments are supported:

* `auto_placement` - (Optional) Indicates whether the host accepts any untargeted instance launches that match its instance type configuration, or if it only accepts Host tenancy instance launches that specify its unique host ID.
* `availability_zone` - (Optional) The AZ to start the host in.
* `host_recovery` - (Optional) Indicates whether to enable or disable host recovery for the Dedicated Host. Host recovery is disabled by default.
* `instance_type` - (Optional) Specifies the instance type for which to configure your Dedicated Hosts. When you specify the instance type, that is the only instance type that you can launch onto that host. 





### Timeouts

The `timeouts` block allows you to specify [timeouts](https://www.terraform.io/docs/configuration/resources.html#timeouts) for certain actions:

* `create` - (Defaults to 10 mins) Used when launching the host (until it reaches the initial `running` state)
* `update` - (Defaults to 10 mins) Used when stopping and starting the host when necessary during update - e.g. when changing host type
* `delete` - (Defaults to 20 mins) Used when terminating the host


## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `host_id` - The host ID. 
* `cores` - The number of cores on the Dedicated Host.
* `instance_family` - The instance family supported by the Dedicated Host. For example, m5.
* `instance_type` -The instance type supported by the Dedicated Host. For example, m5.large. If the host supports multiple instance types, no instanceType is returned.
* `sockets` - The instance family supported by the Dedicated Host. For example, m5.
* `total_vcpus` - The total number of vCPUs on the Dedicated Host.



## Import

hosts can be imported using the `host_id`, e.g.

```
$ terraform import aws_dedicated_host.host_id h-0385a99d0e4b20cbb
```
