---
subcategory: "Transit Gateway Network Manager"
layout: "aws"
page_title: "AWS: aws_networkmanager_global_network"
description: |-
  Provides a global network resource.
---

# Resource: aws_networkmanager_global_network

Provides a global network resource.

## Example Usage

```hcl
resource "aws_networkmanager_global_network" "example" {
  description = "example"
}
```

## Argument Reference

The following arguments are supported:

* `description` - (Optional) Description of the Global Network.
* `tags` - (Optional) Key-value tags for the Global Network.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - Global Network Amazon Resource Name (ARN)

## Import

`aws_networkmanager_global_network` can be imported using the global network ID, e.g.

```
$ terraform import aws_networkmanager_global_network.example global-network-0d47f6t230mz46dy4
```
