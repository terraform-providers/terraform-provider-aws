---
subcategory: "Inspector"
layout: "aws"
page_title: "AWS: aws_inspector_resource_group"
description: |-
  Provides an Amazon Inspector resource group resource.
---

# Resource: aws_inspector_resource_group

Provides an Amazon Inspector resource group resource.

## Example Usage

```hcl
resource "aws_inspector_resource_group" "example" {
  tags = {
    Name = "foo"
    Env  = "bar"
  }
}
```

## Argument Reference

The following arguments are supported:

* `tags` - (Required) Key-value map of tags that are used to select the EC2 instances to be included in an [Amazon Inspector assessment target](/docs/providers/aws/r/inspector_assessment_target.html).

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The resource group ARN.
