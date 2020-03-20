---
subcategory: "RAM"
layout: "aws"
page_title: "AWS: aws_ram_resource_share"
description: |-
  Manages a Resource Access Manager (RAM) Resource Share.
---

# Resource: aws_ram_resource_share

Manages a Resource Access Manager (RAM) Resource Share. To association principals with the share, see the [`aws_ram_principal_association` resource](/docs/providers/aws/r/ram_principal_association.html). To associate resources with the share, see the [`aws_ram_resource_association` resource](/docs/providers/aws/r/ram_resource_association.html).

## Example Usage

```hcl
resource "aws_ram_resource_share" "example" {
  name                      = "example"
  allow_external_principals = true

  tags = {
    Environment = "Production"
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the resource share.
* `allow_external_principals` - (Optional) Indicates whether principals outside your organization can be associated with a resource share.
* `tags` - (Optional) A mapping of tags to assign to the resource share.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The Amazon Resource Name (ARN) of the resource share.
* `id` - The Amazon Resource Name (ARN) of the resource share.

## Import

Resource shares can be imported using the `id`, e.g.

```
$ terraform import aws_ram_resource_share.example arn:aws:ram:eu-west-1:123456789012:resource-share/73da1ab9-b94a-4ba3-8eb4-45917f7f4b12
```
