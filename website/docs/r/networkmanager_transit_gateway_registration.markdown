---
subcategory: "Transit Gateway Network Manager"
layout: "aws"
page_title: "AWS: aws_networkmanager_transit_gateway_registration"
description: |-
  Registers a transit gateway to a global network.
---

# Resource: aws_networkmanager_transit_gateway_registration

Registers a transit gateway to a global network. The transit gateway can be in any AWS Region,
but it must be owned by the same AWS account that owns the global network.
You cannot register a transit gateway in more than one global network.

## Example Usage

```hcl
resource "aws_networkmanager_global_network" "example" {
  description = "example"
}

resource "aws_ec2_transit_gateway" "example" {}

resource "aws_networkmanager_transit_gateway_registration" "example" {
  global_network_id   = aws_networkmanager_global_network.example.id
  transit_gateway_arn = aws_ec2_transit_gateway.example.arn
}
```

## Argument Reference

The following arguments are supported:

* `global_network_id` - (Required) The ID of the Global Network to register to.
* `transit_gateway_arn` - (Required) The ARN of the Transit Gateway to register.

## Import

`aws_networkmanager_transit_gateway_registration` can be imported using the transit gateway ARN and the global network ID, e.g.

```
$ terraform import aws_networkmanager_transit_gateway_registration.example arn:aws:ec2:us-west-2:123456789012:transit-gateway/tgw-123abc05e04123abc,global-network-0d47f6t230mz46dy4
```
