---
layout: "aws"
page_title: "AWS: aws_vpc_peering_connection"
sidebar_current: "docs-aws-resource-vpc-peering"
description: |-
  Provides a resource to manage a VPC peering connection.
---

# aws_vpc_peering_connection

Provides a resource to manage a VPC peering connection.

~> **NOTE on VPC Peering Connections and VPC Peering Connection Options:** Terraform provides
both a standalone [VPC Peering Connection Options](vpc_peering_options.html) and a VPC Peering Connection
resource with `accepter` and `requester` attributes. Do not manage options for the same VPC peering
connection in both a VPC Peering Connection resource and a VPC Peering Connection Options resource.
Doing so will cause a conflict of options and will overwrite the options.
Using a VPC Peering Connection Options resource decouples management of the connection options from
management of the VPC Peering Connection and allows options to be set correctly in cross-account scenarios.

-> **Note:** For cross-account (requester's AWS account differs from the accepter's AWS account) or inter-region
VPC Peering Connections use the `aws_vpc_peering_connection` resource to manage the requester's side of the
connection and use the `aws_vpc_peering_connection_accepter` resource to manage the accepter's side of the connection.

## Example Usage

```hcl
resource "aws_vpc_peering_connection" "foo" {
  peer_owner_id = "${var.peer_owner_id}"
  peer_vpc_id   = "${aws_vpc.bar.id}"
  vpc_id        = "${aws_vpc.foo.id}"
}
```

Basic usage with connection options:

```hcl
resource "aws_vpc_peering_connection" "foo" {
  peer_owner_id = "${var.peer_owner_id}"
  peer_vpc_id   = "${aws_vpc.bar.id}"
  vpc_id        = "${aws_vpc.foo.id}"

  accepter {
    allow_remote_vpc_dns_resolution = true
  }

  requester {
    allow_remote_vpc_dns_resolution = true
  }
}
```

Basic usage with tags:

```hcl
resource "aws_vpc_peering_connection" "foo" {
  peer_owner_id = "${var.peer_owner_id}"
  peer_vpc_id   = "${aws_vpc.bar.id}"
  vpc_id        = "${aws_vpc.foo.id}"
  auto_accept   = true

  tags = {
    Name = "VPC Peering between foo and bar"
  }
}

resource "aws_vpc" "foo" {
  cidr_block = "10.1.0.0/16"
}

resource "aws_vpc" "bar" {
  cidr_block = "10.2.0.0/16"
}
```

Basic usage with region:


```hcl
resource "aws_vpc_peering_connection" "foo" {
  peer_owner_id = "${var.peer_owner_id}"
  peer_vpc_id   = "${aws_vpc.bar.id}"
  vpc_id        = "${aws_vpc.foo.id}"
  peer_region   = "us-east-1"
}

resource "aws_vpc" "foo" {
  provider   = "aws.us-west-2"
  cidr_block = "10.1.0.0/16"
}

resource "aws_vpc" "bar" {
  provider   = "aws.us-east-1"
  cidr_block = "10.2.0.0/16"
}
```

## Argument Reference

-> **Note:** Modifying the VPC Peering Connection options requires peering to be active. An automatic activation
can be done using the [`auto_accept`](vpc_peering.html#auto_accept) attribute. Alternatively, the VPC Peering
Connection has to be made active manually using other means. See [notes](vpc_peering.html#notes) below for
more information.

The following arguments are supported:

* `peer_owner_id` - (Optional) The AWS account ID of the owner of the peer VPC.
   Defaults to the account ID the [AWS provider][1] is currently connected to.
* `peer_vpc_id` - (Required) The ID of the VPC with which you are creating the VPC Peering Connection.
* `vpc_id` - (Required) The ID of the requester VPC.
* `auto_accept` - (Optional) Accept the peering (both VPCs need to be in the same AWS account).
* `peer_region` - (Optional) The region of the accepter VPC of the [VPC Peering Connection]. `auto_accept` must be `false`,
and use the `aws_vpc_peering_connection_accepter` to manage the accepter side.
* `accepter` (Optional) - An optional configuration block that allows for [VPC Peering Connection]
(http://docs.aws.amazon.com/AmazonVPC/latest/PeeringGuide) options to be set for the VPC that accepts
the peering connection (a maximum of one).
* `requester` (Optional) - A optional configuration block that allows for [VPC Peering Connection]
(http://docs.aws.amazon.com/AmazonVPC/latest/PeeringGuide) options to be set for the VPC that requests
the peering connection (a maximum of one).
* `tags` - (Optional) A mapping of tags to assign to the resource.

#### Accepter and Requester Arguments

-> **Note:** When enabled, the DNS resolution feature requires that VPCs participating in the peering
must have support for the DNS hostnames enabled. This can be done using the [`enable_dns_hostnames`]
(vpc.html#enable_dns_hostnames) attribute in the [`aws_vpc`](vpc.html) resource. See [Using DNS with Your VPC]
(http://docs.aws.amazon.com/AmazonVPC/latest/UserGuide/vpc-dns.html) user guide for more information.

* `allow_remote_vpc_dns_resolution` - (Optional) Allow a local VPC to resolve public DNS hostnames to
private IP addresses when queried from instances in the peer VPC. This is
[not supported](https://docs.aws.amazon.com/vpc/latest/peering/modify-peering-connections.html) for
inter-region VPC peering.
* `allow_classic_link_to_remote_vpc` - (Optional) Allow a local linked EC2-Classic instance to communicate
with instances in a peer VPC. This enables an outbound communication from the local ClassicLink connection
to the remote VPC.
* `allow_vpc_to_remote_classic_link` - (Optional) Allow a local VPC to communicate with a linked EC2-Classic
instance in a peer VPC. This enables an outbound communication from the local VPC to the remote ClassicLink
connection.

### Timeouts

`aws_vpc_peering_connection` provides the following
[Timeouts](/docs/configuration/resources.html#timeouts) configuration options:

- `create` - (Default `1 minute`) Used for creating a peering connection
- `update` - (Default `1 minute`) Used for peering connection modifications
- `delete` - (Default `1 minute`) Used for destroying peering connections

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the VPC Peering Connection.
* `accept_status` - The status of the VPC Peering Connection request.


## Notes

If both VPCs are not in the same AWS account do not enable the `auto_accept` attribute.
The accepter can manage its side of the connection using the `aws_vpc_peering_connection_accepter` resource
or accept the connection manually using the AWS Management Console, AWS CLI, through SDKs, etc.

## Import

VPC Peering resources can be imported using the `vpc peering id`, e.g.

```sh
$ terraform import aws_vpc_peering_connection.test_connection pcx-111aaa111
```

[1]: /docs/providers/aws/index.html
