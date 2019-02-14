---
layout: "aws"
page_title: "AWS: aws_dax_parameter_group"
sidebar_current: "docs-aws-resource-dax-parameter-group"
description: |-
  Provides an DAX Parameter Group resource.
---

# aws_dax_parameter_group

Provides a DAX Parameter Group resource.

## Example Usage

```hcl
resource "aws_dax_parameter_group" "example" {
  name = "example"

  parameters {
    name  = "query-ttl-millis"
    value = "100000"
  }

  parameters {
    name  = "record-ttl-millis"
    value = "100000"
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` – (Required) The name of the parameter group.

* `description` - (Optional, ForceNew) A description of the parameter group.

* `parameters` – (Optional) The parameters of the parameter group.

## parameters

`parameters` supports the following:

* `name` - (Required) The name of the parameter.
* `value` - (Required) The value for the parameter.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The name of the parameter group.

## Import

DAX Parameter Group can be imported using the `name`, e.g.

```
$ terraform import aws_dax_parameter_group.example my_dax_pg
```
