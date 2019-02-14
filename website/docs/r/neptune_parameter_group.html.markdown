---
layout: "aws"
page_title: "AWS: aws_neptune_parameter_group"
sidebar_current: "docs-aws-resource-aws-neptune-parameter-group"
description: |-
  Manages a Neptune Parameter Group
---

# aws_neptune_parameter_group

Manages a Neptune Parameter Group

## Example Usage

```hcl
resource "aws_neptune_parameter_group" "example" {
  family = "neptune1"
  name   = "example"

  parameter {
    name  = "neptune_query_timeout"
    value = "25"
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required, Forces new resource) The name of the Neptune parameter group.
* `family` - (Required) The family of the Neptune parameter group.
* `description` - (Optional) The description of the Neptune parameter group. Defaults to "Managed by Terraform".
* `parameter` - (Optional) A list of Neptune parameters to apply.
* `tags` - (Optional) A mapping of tags to assign to the resource.

Parameter blocks support the following:

* `name`  - (Required) The name of the Neptune parameter.
* `value` - (Required) The value of the Neptune parameter.
* `apply_method` - (Optional) The apply method of the Neptune parameter. Valid values are `immediate` and `pending-reboot`. Defaults to `pending-reboot`.


## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The Neptune parameter group name.
* `arn` - The Neptune parameter group Amazon Resource Name (ARN).

## Import

Neptune Parameter Groups can be imported using the `name`, e.g.

```
$ terraform import aws_neptune_parameter_group.some_pg some-pg
```
