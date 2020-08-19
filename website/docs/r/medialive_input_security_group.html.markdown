---
subcategory: "MediaLive"
layout: "aws"
page_title: "AWS: aws_medialive_input_security_group"
description: |-
  Provides an AWS Elemental MediaLive Input Security Group.
---

# Resource: aws_medialive_input_security_group

Provides an AWS Elemental MediaLive Input Security Group.

## Example Usage

```hcl
resource "aws_medialive_input_security_group" "test" {
  whitelist_rule {
    cidr = "10.0.0.0/8"
  }
}
```

## Argument Reference

The following arguments are supported:

* `whitelist_rule` - (Optional) A detail whitelist rule. See below.
* `tags` - (Optional) A mapping of tags to assign to the resource.

### Nested Fields

#### `whitelist_rule`

* `cidr` - (Required) The IPv4 CIDR to whitelist.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The Id of the input security group
* `arn` - The Arn of the input security group

## Import

Media Live Input Security Group can be imported via the input security group id, e.g.

```
$ terraform import aws_medialive_input_security_group.test 1234567
```
