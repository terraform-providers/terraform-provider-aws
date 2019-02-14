---
layout: "aws"
page_title: "AWS: aws_dax_subnet_group"
sidebar_current: "docs-aws-resource-dax-subnet-group"
description: |-
  Provides an DAX Subnet Group resource.
---

# aws_dax_subnet_group

Provides a DAX Subnet Group resource.

## Example Usage

```hcl
resource "aws_dax_subnet_group" "example" {
  name       = "example"
  subnet_ids = ["${aws_subnet.example1.id}", "${aws_subnet.example2.id}"]
}
```

## Argument Reference

The following arguments are supported:

* `name` – (Required) The name of the subnet group.
* `description` - (Optional) A description of the subnet group.
* `subnet_ids` – (Required) A list of VPC subnet IDs for the subnet group.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The name of the subnet group.
* `vpc_id` – VPC ID of the subnet group.

## Import

DAX Subnet Group can be imported using the `name`, e.g.

```
$ terraform import aws_dax_subnet_group.example my_dax_sg
```
