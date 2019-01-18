---
layout: "aws"
page_title: "AWS: aws_resourcegroups_group"
sidebar_current: "docs-aws-resource-resourcegroups-group"
description: |-
  Provides a Resource Group.
---

# aws_resourcegroups_group

Provides a Resource Group.

## Example Usage

```hcl
resource "aws_resourcegroups_group" "test" {
  name        = "test-group"

  resource_query {
    query = <<JSON
{
  "ResourceTypeFilters": [
    "AWS::EC2::Instance"
  ],
  "TagFilters": [
    {
      "Key": "Stage",
      "Values": ["Test"]
    }
  ]
}
JSON
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The resource group's name. A resource group name can have a maximum of 127 characters, including letters, numbers, hyphens, dots, and underscores. The name cannot start with `AWS` or `aws`.
* `description` - (Optional) A description of the resource group.
* `resource_query` - (Required) A `resource_query` block. Resource queries are documented below.

An `resource_query` block supports the following arguments:

* `query` - (Required) The resource query as a JSON string.
* `type` - (Required) The type of the resource query. Defaults to `TAG_FILTERS_1_0`. 

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The ARN assigned by AWS for this resource group.

## Import

Resource groups can be imported using the `name`, e.g.

```
$ terraform import aws_resourcegroups_group.foo resource-group-name
```
