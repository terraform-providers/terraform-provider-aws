---
subcategory: "Direct Connect"
layout: "aws"
page_title: "AWS: aws_dx_connection_association"
description: |-
  Associates a Direct Connect Connection with a LAG.
---

# Resource: aws_dx_connection_association

Associates a Direct Connect Connection with a LAG.

## Example Usage

```hcl
resource "aws_dx_connection" "example" {
  name      = "example"
  bandwidth = "1Gbps"
  location  = "EqSe2-EQ"
}

resource "aws_dx_lag" "example" {
  name                  = "example"
  connections_bandwidth = "1Gbps"
  location              = "EqSe2-EQ"
}

resource "aws_dx_connection_association" "example" {
  connection_id = aws_dx_connection.example.id
  lag_id        = aws_dx_lag.example.id
}
```

## Argument Reference

The following arguments are supported:

* `connection_id` - (Required) The ID of the connection.
* `lag_id` - (Required) The ID of the LAG with which to associate the connection.
