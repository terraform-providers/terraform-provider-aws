---
layout: "aws"
page_title: "AWS: aws_ec2_capacity_reservation"
sidebar_current: "docs-aws-resource-ec2-capacity-reservation"
description: |-
  Provides an EC2 Capacity Reservation. This allows you to reserve capacity for your Amazon EC2 instances in a specific Availability Zone for any duration.
---

# aws_ec2_capacity_reservation

Provides an EC2 Capacity Reservation. This allows you to reserve capacity for your Amazon EC2 instances in a specific Availability Zone for any duration.

## Example Usage

```hcl
resource "aws_ec2_capacity_reservation" "default" {
  instance_type     = "t2.micro"
  instance_platform = "Linux/UNIX"
  availability_zone = "eu-west-1a"
  instance_count    = 1
}
```

## Argument Reference

The following arguments are supported:

* `availability_zone` - (Required) The Availability Zone in which to create the Capacity Reservation.
* `ebs_optimized` - (Optional) Indicates whether the Capacity Reservation supports EBS-optimized instances.
* `end_date` - (Optional) The date and time at which the Capacity Reservation expires. When a Capacity Reservation expires, the reserved capacity is released and you can no longer launch instances into it. Valid values: [RFC3339 time string](https://tools.ietf.org/html/rfc3339#section-5.8) (`YYYY-MM-DDTHH:MM:SSZ`)
* `end_date_type` - (Optional) Indicates the way in which the Capacity Reservation ends. Specify either `unlimited` or `limited`.
* `ephemeral_storage` - (Optional) Indicates whether the Capacity Reservation supports instances with temporary, block-level storage.
* `instance_count` - (Required) The number of instances for which to reserve capacity.
* `instance_match_criteria` - (Optional) Indicates the type of instance launches that the Capacity Reservation accepts. Specify either `open` or `targeted`.
* `instance_platform` - (Required) The type of operating system for which to reserve capacity. Valid options are `Linux/UNIX`, `Red Hat Enterprise Linux`, `SUSE Linux`, `Windows`, `Windows with SQL Server`, `Windows with SQL Server Enterprise`, `Windows with SQL Server Standard` or `Windows with SQL Server Web`.
* `instance_type` - (Required) The instance type for which to reserve capacity.
* `tags` - (Optional) A mapping of tags to assign to the resource.
* `tenancy` - (Optional) Indicates the tenancy of the Capacity Reservation. Specify either `default` or `dedicated`.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The Capacity Reservation ID.

## Import

Capacity Reservations can be imported using the `id`, e.g.

```
$ terraform import aws_ec2_capacity_reservation.web cr-0123456789abcdef0
```
