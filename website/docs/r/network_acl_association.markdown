---
subcategory: "VPC"
layout: "aws"
page_title: "AWS: aws_network_acl_association"
description: |-
  Provides an network ACL association resource.
---

# Resource: aws_network_acl_association

Provides an network ACL association resource which allows you to associate your network ACL with any subnet(s).

## Example Usage

```terraform
resource "aws_network_acl_association" "main" {
  network_acl_id = aws_network_acl.main.id
  subnet_id      = aws_subnet.main.id
}
```

## Argument Reference

The following arguments are supported:

* `network_acl_id` - (Required) The ID of the network ACL.
* `subnet_id` - (Required) The ID of the associated Subnet.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the network ACL association