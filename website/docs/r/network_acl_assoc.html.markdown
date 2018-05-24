---
layout: "aws"
page_title: "AWS: aws_network_acl_association"
sidebar_current: "docs-aws-resource-network-acl-association"
description: |-
  Provides an network ACL association resource.
---

Provides an network ACL association resource. You might set up network ACLs associate to your subnet.

## Example Usage

```hcl
resource "aws_network_acl_association" "main" {
    network_acl_id = "${aws_network_acl.main.id}"
    subnet_id = "${aws_subnet.main.id}"
}
```

## Argument Reference

The following arguments are supported:

* `network_acl_id` - (Required) The ID of the network acl .
* `subnet_id` - (Required) The ID of the associated Subnet.

## Attributes Reference

The following attributes are exported:

* `id` - The ID of the network ACL
* `network_acl_id` - The ID of the network ACL
* `subnet_id` - The ID of the subnet id