---
layout: "aws"
page_title: "AWS: aws_inspector_assessment_target"
sidebar_current: "docs-aws-resource-inspector-assessment-target"
description: |-
  Provides a Inspector assessment target.
---

# aws_inspector_assessment_target

Provides a Inspector assessment target

## Example Usage

```hcl
resource "aws_inspector_resource_group" "bar" {
  tags = {
    Name = "foo"
    Env  = "bar"
  }
}

resource "aws_inspector_assessment_target" "foo" {
  name               = "assessment target"
  resource_group_arn = "${aws_inspector_resource_group.bar.arn}"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the assessment target.
* `resource_group_arn` (Optional) Inspector Resource Group Amazon Resource Name (ARN) stating tags for instance matching. If not specified, all EC2 instances in the current AWS account and region are included in the assessment target.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The target assessment ARN.

## Import

Inspector Assessment Targets can be imported via their Amazon Resource Name (ARN), e.g.

```sh
$ terraform import aws_inspector_assessment_target.example arn:aws:inspector:us-east-1:123456789012:target/0-xxxxxxx
```
