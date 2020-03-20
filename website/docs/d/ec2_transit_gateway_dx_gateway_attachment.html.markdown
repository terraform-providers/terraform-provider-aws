---
subcategory: "EC2"
layout: "aws"
page_title: "AWS: aws_ec2_transit_gateway_dx_gateway_attachment"
description: |-
  Get information on an EC2 Transit Gateway's attachment to a Direct Connect Gateway
---

# Data Source: aws_ec2_transit_gateway_dx_gateway_attachment

Get information on an EC2 Transit Gateway's attachment to a Direct Connect Gateway.

## Example Usage

### By Transit Gateway and Direct Connect Gateway Identifiers

```hcl
data "aws_ec2_transit_gateway_dx_gateway_attachment" "example" {
  transit_gateway_id = "${aws_ec2_transit_gateway.example.id}"
  dx_gateway_id      = "${aws_dx_gateway.example.id}"
}
```

## Argument Reference

The following arguments are supported:

* `transit_gateway_id` - (Required) Identifier of the EC2 Transit Gateway.
* `dx_gateway_id` - (Required) Identifier of the Direct Connect Gateway.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - EC2 Transit Gateway Attachment identifier
* `tags` - Key-value tags for the EC2 Transit Gateway Attachment
