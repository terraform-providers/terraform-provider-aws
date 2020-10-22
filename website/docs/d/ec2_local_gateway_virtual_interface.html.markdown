---
subcategory: "EC2"
layout: "aws"
page_title: "AWS: aws_ec2_local_gateway_virtual_interface"
description: |-
    Provides details about an EC2 Local Gateway Virtual Interface
---

# Data Source: aws_ec2_local_gateway_virtual_interface

Provides details about an EC2 Local Gateway Virtual Interface. More information can be found in the [Outposts User Guide](https://docs.aws.amazon.com/outposts/latest/userguide/outposts-networking-components.html#routing).

## Example Usage

```hcl
data "aws_ec2_local_gateway_virtual_interface" "example" {
  for_each = data.aws_ec2_local_gateway_virtual_interface_group.example.local_gateway_virtual_interface_ids

  id = each.value
}
```

## Argument Reference

The following arguments are optional:

* `filter` - (Optional) One or more configuration blocks containing name-values filters. See the [EC2 API Reference](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DescribeLocalGatewayVirtualInterfaces.html) for supported filters. Detailed below.
* `id` - (Optional) Identifier of EC2 Local Gateway Virtual Interface.
* `tags` - (Optional) Key-value map of resource tags, each pair of which must exactly match a pair on the desired local gateway route table.

### filter Argument Reference

The `filter` configuration block supports the following arguments:

* `name` - (Required) Name of the filter.
* `values` - (Required) List of one or more values for the filter.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `local_address` - Local address.
* `local_bgp_asn` - Border Gateway Protocol (BGP) Autonomous System Number (ASN) of the EC2 Local Gateway.
* `local_gateway_id` - Identifier of the EC2 Local Gateway.
* `peer_address` - Peer address.
* `peer_bgp_asn` - Border Gateway Protocol (BGP) Autonomous System Number (ASN) of the peer.
* `vlan` - Virtual Local Area Network.
