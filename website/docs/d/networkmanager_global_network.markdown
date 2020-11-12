---
subcategory: "Transit Gateway Network Manager"
layout: "aws"
page_title: "AWS: aws_networkmanager_global_network"
description: |-
  Retrieve information about a global network.
---

# Data Source: aws_networkmanager_global_network

Retrieve information about a global network.

## Example Usage

```hcl
data "aws_networkmanager_global_network" "example" {
  id = var.global_network_id
}
```

## Argument Reference

* `id` - (Optional) The id of the specific global network to retrieve.

* `tags` - (Optional) A map of tags, each pair of which must exactly match
  a pair on the desired global network.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The ARN of the global network.
* `description` - The description of the global network.
