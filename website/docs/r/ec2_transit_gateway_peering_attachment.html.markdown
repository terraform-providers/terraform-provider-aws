---
subcategory: "EC2"
layout: "aws"
page_title: "AWS: aws_ec2_transit_gateway_peering_attachment"
description: |-
  Manages an EC2 Transit Gateway Peering Attachment
---

# Resource: aws_ec2_transit_gateway_peering_attachment

Manages an EC2 Transit Gateway Peering Attachment.
For examples of custom route table association and propagation, see the [EC2 Transit Gateway Networking Examples Guide](https://docs.aws.amazon.com/vpc/latest/tgw/TGW_Scenarios.html).

## Example Usage

```hcl
provider "aws" {
  alias  = "local"
  region = "us-east-1"
}

provider "aws" {
  alias  = "peer"
  region = "us-west-2"
}

data "aws_region" "peer" {
  provider = aws.peer
}

resource "aws_ec2_transit_gateway" "local" {
  provider = aws.local

  tags = {
    Name = "Local TGW"
  }
}

resource "aws_ec2_transit_gateway" "peer" {
  provider = aws.peer

  tags = {
    Name = "Peer TGW"
  }
}

resource "aws_ec2_transit_gateway_peering_attachment" "example" {
  peer_account_id         = aws_ec2_transit_gateway.peer.owner_id
  peer_region             = data.aws_region.peer.name
  peer_transit_gateway_id = aws_ec2_transit_gateway.peer.id
  transit_gateway_id      = aws_ec2_transit_gateway.local.id

  tags = {
    Name = "TGW Peering Requestor"
  }
}
```

A full example of how to create a Transit Gateway in one AWS account, share it with a second AWS account, and attach a to a Transit Gateway in the second account via the `aws_ec2_transit_gateway_peering_attachment` resource can be found in [the `./examples/transit-gateway-cross-account-peering-attachment` directory within the Github Repository](https://github.com/hashicorp/terraform-provider-aws/tree/master/examples/transit-gateway-cross-account-peering-attachment).

## Argument Reference

The following arguments are supported:

* `peer_account_id` - (Optional) Account ID of EC2 Transit Gateway to peer with. Defaults to the account ID the [AWS provider][1] is currently connected to.
* `peer_region` - (Required) Region of EC2 Transit Gateway to peer with.
* `peer_transit_gateway_id` - (Required) Identifier of EC2 Transit Gateway to peer with.
* `tags` - (Optional) Key-value tags for the EC2 Transit Gateway Peering Attachment.
* `transit_gateway_id` - (Required) Identifier of EC2 Transit Gateway.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - EC2 Transit Gateway Attachment identifier

## Import

`aws_ec2_transit_gateway_peering_attachment` can be imported by using the EC2 Transit Gateway Attachment identifier, e.g.

```sh
terraform import aws_ec2_transit_gateway_peering_attachment.example tgw-attach-12345678
```

[1]: /docs/providers/aws/index.html
