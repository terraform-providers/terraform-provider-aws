---
subcategory: "Direct Connect"
layout: "aws"
page_title: "AWS: aws_dx_gateway_association"
description: |-
  Associates a Direct Connect Gateway with a VGW or transit gateway.
---

# Resource: aws_dx_gateway_association

Associates a Direct Connect Gateway with a VGW or transit gateway.

To create a cross-account association, create an [`aws_dx_gateway_association_proposal` resource](/docs/providers/aws/r/dx_gateway_association_proposal.html)
in the AWS account that owns the VGW or transit gateway and then accept the proposal in the AWS account that owns the Direct Connect Gateway
by creating an `aws_dx_gateway_association` resource with the `proposal_id` and `associated_gateway_owner_account_id` attributes set.

## Example Usage

### VPN Gateway Association

```hcl
resource "aws_dx_gateway" "example" {
  name            = "example"
  amazon_side_asn = "64512"
}

resource "aws_vpc" "example" {
  cidr_block = "10.255.255.0/28"
}

resource "aws_vpn_gateway" "example" {
  vpc_id = aws_vpc.example.id
}

resource "aws_dx_gateway_association" "example" {
  dx_gateway_id         = aws_dx_gateway.example.id
  associated_gateway_id = aws_vpn_gateway.example.id
}
```

### Transit Gateway Association

```hcl
resource "aws_dx_gateway" "example" {
  name            = "example"
  amazon_side_asn = "64512"
}

resource "aws_ec2_transit_gateway" "example" {
}

resource "aws_dx_gateway_association" "example" {
  dx_gateway_id         = aws_dx_gateway.example.id
  associated_gateway_id = aws_ec2_transit_gateway.example.id

  allowed_prefixes = [
    "10.255.255.0/30",
    "10.255.255.8/30",
  ]
}
```

### Allowed Prefixes

```hcl
resource "aws_dx_gateway" "example" {
  name            = "example"
  amazon_side_asn = "64512"
}

resource "aws_vpc" "example" {
  cidr_block = "10.255.255.0/28"
}

resource "aws_vpn_gateway" "example" {
  vpc_id = aws_vpc.example.id
}

resource "aws_dx_gateway_association" "example" {
  dx_gateway_id         = aws_dx_gateway.example.id
  associated_gateway_id = aws_vpn_gateway.example.id

  allowed_prefixes = [
    "210.52.109.0/24",
    "175.45.176.0/22",
  ]
}
```

A full example of how to create a VPN Gateway in one AWS account, create a Direct Connect Gateway in a second AWS account, and associate the VPN Gateway with the Direct Connect Gateway via the `aws_dx_gateway_association_proposal` and `aws_dx_gateway_association` resources can be found in [the `./examples/dx-gateway-cross-account-vgw-association` directory within the Github Repository](https://github.com/hashicorp/terraform-provider-aws/tree/master/examples/dx-gateway-cross-account-vgw-association).

## Argument Reference

~> **NOTE:** `dx_gateway_id` and `associated_gateway_id` must be specified for single account Direct Connect gateway associations.

The following arguments are supported:

* `dx_gateway_id` - (Required) The ID of the Direct Connect gateway.
* `associated_gateway_id` - (Optional) The ID of the VGW or transit gateway with which to associate the Direct Connect gateway.
Used for single account Direct Connect gateway associations.
* `associated_gateway_owner_account_id` - (Optional) The ID of the AWS account that owns the VGW or transit gateway with which to associate the Direct Connect gateway.
Used for cross-account Direct Connect gateway associations.
* `proposal_id` - (Optional) The ID of the Direct Connect gateway association proposal.
Used for cross-account Direct Connect gateway associations.
* `allowed_prefixes` - (Optional) VPC prefixes (CIDRs) to advertise to the Direct Connect gateway. Defaults to the CIDR block of the VPC associated with the Virtual Gateway. To enable drift detection, must be configured.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the Direct Connect gateway association resource.
* `associated_gateway_type` - The type of the associated gateway, `transitGateway` or `virtualPrivateGateway`.
* `dx_gateway_association_id` - The ID of the Direct Connect gateway association.
* `dx_gateway_owner_account_id` - The ID of the AWS account that owns the Direct Connect gateway.

## Timeouts

`aws_dx_gateway_association` provides the following
[Timeouts](https://www.terraform.io/docs/configuration/blocks/resources/syntax.html#operation-timeouts) configuration options:

- `create` - (Default `30 minutes`) Used for creating the association
- `update` - (Default `30 minutes`) Used for updating the association
- `delete` - (Default `30 minutes`) Used for destroying the association

## Import

Direct Connect gateway associations can be imported using `dx_gateway_id` together with `associated_gateway_id`,
e.g.

```
$ terraform import aws_dx_gateway_association.example dxgw-12345678/vgw-98765432
```
