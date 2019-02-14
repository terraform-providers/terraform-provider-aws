---
layout: "aws"
page_title: "AWS: aws_dx_gateway"
sidebar_current: "docs-aws-resource-dx-gateway"
description: |-
  Provides a Direct Connect Gateway.
---

# aws_dx_gateway

Provides a Direct Connect Gateway.

## Example Usage

```hcl
resource "aws_dx_gateway" "example" {
  name            = "tf-dxg-example"
  amazon_side_asn = "64512"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the connection.
* `amazon_side_asn` - (Required) The ASN to be configured on the Amazon side of the connection. The ASN must be in the private range of 64,512 to 65,534 or 4,200,000,000 to 4,294,967,294.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the gateway.

## Timeouts

`aws_dx_gateway` provides the following
[Timeouts](/docs/configuration/resources.html#timeouts) configuration options:

- `create` - (Default `10 minutes`) Used for creating the gateway
- `delete` - (Default `10 minutes`) Used for destroying the gateway

## Import

Direct Connect Gateways can be imported using the `gateway id`, e.g.

```
$ terraform import aws_dx_gateway.test abcd1234-dcba-5678-be23-cdef9876ab45
```
