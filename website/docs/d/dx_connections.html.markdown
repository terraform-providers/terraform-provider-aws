---
subcategory: "Direct Connect"
layout: "aws"
page_title: "AWS: aws_dx_connections"
description: |-
  Provides details about multiple DirectConnect connections.
---

# Data Source: aws_dx_connections

Provides details about multiple DirectConnect connections.

## Example Usage

```hcl
data "aws_dx_connections" "example" {
  name = "My-DX-Connection"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Optional) The connection name.
* `tags` - (Optional) A map of tags, each pair of which must exactly match
  a pair on the desired connections.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `ids` - Set of connection identifiers.

