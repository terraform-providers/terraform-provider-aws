---
subcategory: "Direct Connect"
layout: "aws"
page_title: "AWS: aws_dx_gateway_association_proposal"
description: |-
  Manages a Direct Connect Gateway Association Proposal.
---

# Resource: aws_dx_gateway_association_proposal

Manages a Direct Connect Gateway Association Proposal, typically for enabling cross-account associations. For single account associations, see the [`aws_dx_gateway_association` resource](/docs/providers/aws/r/dx_gateway_association.html).

## Example Usage

```hcl
resource "aws_dx_gateway_association_proposal" "example" {
  dx_gateway_id               = aws_dx_gateway.example.id
  dx_gateway_owner_account_id = aws_dx_gateway.example.owner_account_id
  associated_gateway_id       = aws_vpn_gateway.example.id
}
```

A full example of how to create a VPN Gateway in one AWS account, create a Direct Connect Gateway in a second AWS account, and associate the VPN Gateway with the Direct Connect Gateway via the `aws_dx_gateway_association_proposal` and `aws_dx_gateway_association` resources can be found in [the `./examples/dx-gateway-cross-account-vgw-association` directory within the Github Repository](https://github.com/hashicorp/terraform-provider-aws/tree/master/examples/dx-gateway-cross-account-vgw-association).

## Argument Reference

The following arguments are supported:

* `associated_gateway_id` - (Required) The ID of the VGW or transit gateway with which to associate the Direct Connect gateway.
* `dx_gateway_id` - (Required) Direct Connect Gateway identifier.
* `dx_gateway_owner_account_id` - (Required) AWS Account identifier of the Direct Connect Gateway's owner.
* `allowed_prefixes` - (Optional) VPC prefixes (CIDRs) to advertise to the Direct Connect gateway. Defaults to the CIDR block of the VPC associated with the Virtual Gateway. To enable drift detection, must be configured.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - Direct Connect Gateway Association Proposal identifier.
* `associated_gateway_owner_account_id` - The ID of the AWS account that owns the VGW or transit gateway with which to associate the Direct Connect gateway.
* `associated_gateway_type` - The type of the associated gateway, `transitGateway` or `virtualPrivateGateway`.

## Import

Direct Connect Gateway Association Proposals can be imported using the proposal ID, e.g.

```
$ terraform import aws_dx_gateway_association_proposal.example ac90e981-b718-4364-872d-65478c84fafe
```
