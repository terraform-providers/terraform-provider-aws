---
layout: "aws"
page_title: "AWS: aws_dx_gateway_association"
sidebar_current: "docs-aws-resource-dx-gateway-association"
description: |-
  Associates a Direct Connect Gateway with a VGW.
---

# aws_dx_gateway_association

Associates a Direct Connect Gateway with a VGW.

## Example Usage

```hcl
resource "aws_dx_gateway" "example" {
  name            = "example"
  amazon_side_asn = "64512"
}

resource "aws_vpc" "example" {
  cidr_block = "10.255.255.0/28"
}

resource "aws_vpn_gateway" "example" {
  vpc_id = "${aws_vpc.test.id}"
}

resource "aws_dx_gateway_association" "example" {
  dx_gateway_id  = "${aws_dx_gateway.example.id}"
  vpn_gateway_id = "${aws_vpn_gateway.example.id}"
}
```

## Argument Reference

The following arguments are supported:

* `dx_gateway_id` - (Required) The ID of the Direct Connect Gateway.
* `vpn_gateway_id` - (Required) The ID of the VGW with which to associate the gateway.

## Timeouts

`aws_dx_gateway_association` provides the following
[Timeouts](/docs/configuration/resources.html#timeouts) configuration options:

- `create` - (Default `15 minutes`) Used for creating the association
- `delete` - (Default `10 minutes`) Used for destroying the association
