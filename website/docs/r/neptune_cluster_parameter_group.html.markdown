---
layout: "aws"
page_title: "AWS: aws_neptune_cluster_parameter_group"
sidebar_current: "docs-aws-resource-aws-neptune-cluster-parameter-group"
description: |-
  Manages a Neptune Cluster Parameter Group
---

# aws_neptune_cluster_parameter_group

Manages a Neptune Cluster Parameter Group

## Example Usage

```hcl
resource "aws_neptune_cluster_parameter_group" "example" {
  family      = "neptune1"
  name        = "example"
  description = "neptune cluster parameter group"

  parameter {
    name  = "neptune_enable_audit_log"
    value = 1
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Optional, Forces new resource) The name of the neptune cluster parameter group. If omitted, Terraform will assign a random, unique name.
* `name_prefix` - (Optional, Forces new resource) Creates a unique name beginning with the specified prefix. Conflicts with `name`.
* `family` - (Required) The family of the neptune cluster parameter group.
* `description` - (Optional) The description of the neptune cluster parameter group. Defaults to "Managed by Terraform".
* `parameter` - (Optional) A list of neptune parameters to apply.
* `tags` - (Optional) A mapping of tags to assign to the resource.

Parameter blocks support the following:

* `name` - (Required) The name of the neptune parameter.
* `value` - (Required) The value of the neptune parameter.
* `apply_method` - (Optional) Valid values are `immediate` and `pending-reboot`. Defaults to `pending-reboot`.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The neptune cluster parameter group name.
* `arn` - The ARN of the neptune cluster parameter group.


## Import

Neptune Cluster Parameter Groups can be imported using the `name`, e.g.

```
$ terraform import aws_neptune_cluster_parameter_group.cluster_pg production-pg-1
```
