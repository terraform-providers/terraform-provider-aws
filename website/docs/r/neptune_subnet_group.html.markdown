---
layout: "aws"
page_title: "AWS: aws_neptune_subnet_group"
sidebar_current: "docs-aws-resource-neptune-subnet-group"
description: |-
  Provides an Neptune subnet group resource.
---

# aws_neptune_subnet_group

Provides an Neptune subnet group resource.

## Example Usage

```hcl
resource "aws_neptune_subnet_group" "default" {
  name       = "main"
  subnet_ids = ["${aws_subnet.frontend.id}", "${aws_subnet.backend.id}"]

  tags = {
    Name = "My neptune subnet group"
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Optional, Forces new resource) The name of the neptune subnet group. If omitted, Terraform will assign a random, unique name.
* `name_prefix` - (Optional, Forces new resource) Creates a unique name beginning with the specified prefix. Conflicts with `name`.
* `description` - (Optional) The description of the neptune subnet group. Defaults to "Managed by Terraform".
* `subnet_ids` - (Required) A list of VPC subnet IDs.
* `tags` - (Optional) A mapping of tags to assign to the resource.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The neptune subnet group name.
* `arn` - The ARN of the neptune subnet group.


## Import

Neptune Subnet groups can be imported using the `name`, e.g.

```
$ terraform import aws_neptune_subnet_group.default production-subnet-group
```
