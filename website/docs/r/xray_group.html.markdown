---
subcategory: "XRay"
layout: "aws"
page_title: "AWS: aws_xray_group"
description: |-
    Creates and manages an AWS XRay Group.
---

# Resource: aws_xray_group

Creates and manages an AWS XRay Group.

## Example Usage

```hcl
resource "aws_xray_group" "example" {
  group_name        = "example"
  filter_expression = "responsetime > 5"
}
```

## Argument Reference

* `group_name` - (Required) The name of the group.
* `filter_expression` - (Required) The filter expression defining criteria by which to group traces. more info can be found in official [docs](https://docs.aws.amazon.com/xray/latest/devguide/xray-console-filters.html).
* `tags` - (Optional) Key-value mapping of resource tags

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ARN of the Group.
* `arn` - The ARN of the Group.

## Import

XRay Groups can be imported using the ARN, e.g.

```
$ terraform import aws_xray_group.example arn:aws:xray:us-west-2:1234567890:group/example-group/TNGX7SW5U6QY36T4ZMOUA3HVLBYCZTWDIOOXY3CJAXTHSS3YCWUA
```
