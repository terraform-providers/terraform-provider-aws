---
subcategory: "EC2"
layout: "aws"
page_title: "AWS: aws_dedicated_host"
description: |-
  Provides an EC2 host resource. This allows hosts to be created, updated, and deleted.
---

# Resource: aws_dedicated_host

Provides an EC2 host resource. This allows hosts to be created, updated,
and deleted. hosts also support [provisioning](/docs/provisioners/index.html).

## Example Usage

```hcl
# Create a new host of the latest Ubuntu 14.04 on an
# t2.micro node with an AWS Tag naming it "HelloWorld"
provider "aws" {
  region = "us-west-2"
}

data "aws_dedicated_host" "host" {
  filter {
    instance_type   = "c5.18xlarge"
  }
}

resource "aws_dedicated_host" "test" {
      + auto_placement    = true
      + availability_zone = "us-west-2a"
      + host_recovery     = true
      + instance_type     = "c5.18xlarge"
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

* `id` - The host ID.


## Import

hosts can be imported using the `id`, e.g.

```
$ terraform import aws_dedicated_host.host h-0385a99d0e4b20cbb
```
