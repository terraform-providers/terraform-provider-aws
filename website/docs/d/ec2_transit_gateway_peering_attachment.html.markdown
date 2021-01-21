---
subcategory: "EC2"
layout: "aws"
page_title: "AWS: aws_ec2_transit_gateway_peering_attachment"
description: |-
  Get information on an EC2 Transit Gateway Peering Attachment
---

# Data Source: aws_ec2_transit_gateway_peering_attachment

Get information on an EC2 Transit Gateway Peering Attachment.

## Example Usage

### By Filter

```hcl
data "aws_ec2_transit_gateway_peering_attachment" "example" {
  filter {
    name   = "transit-gateway-attachment-id"
    values = ["tgw-attach-12345678"]
  }
}
```

### By Identifier

```hcl
data "aws_ec2_transit_gateway_peering_attachment" "attachment" {
  id = "tgw-attach-12345678"
}
```

## Argument Reference

The following arguments are supported:

* `filter` - (Optional) One or more configuration blocks containing name-values filters. Detailed below.
* `id` - (Optional) Identifier of the EC2 Transit Gateway Peering Attachment.
* `tags` - (Optional) A mapping of tags, each pair of which must exactly match
  a pair on the specific EC2 Transit Gateway Peering Attachment to retrieve.

More complex filters can be expressed using one or more `filter` sub-blocks,
which take the following arguments:

* `name` - (Required) The name of the field to filter by, as defined by
  [the underlying AWS API](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DescribeTransitGatewayPeeringAttachments.html).
* `values` - (Required) Set of values that are accepted for the given field.
  An EC2 Transit Gateway Peering Attachment be selected if any one of the given values matches.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `peer_account_id` - Identifier of the peer AWS account
* `peer_region` - Identifier of the peer AWS region
* `peer_transit_gateway_id` - Identifier of the peer EC2 Transit Gateway
* `transit_gateway_id` - Identifier of the local EC2 Transit Gateway
