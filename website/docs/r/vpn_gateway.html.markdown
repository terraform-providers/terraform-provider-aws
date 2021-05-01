---
subcategory: "VPC"
layout: "aws"
page_title: "AWS: aws_vpn_gateway"
description: |-
  Provides a resource to create a VPC VPN Gateway.
---

# Resource: aws_vpn_gateway

Provides a resource to create a VPC VPN Gateway.

## Example Usage

```terraform
resource "aws_vpn_gateway" "vpn_gw" {
  vpc_id = aws_vpc.main.id

  tags = {
    Name = "main"
  }
}
```

## Argument Reference

The following arguments are supported:

* `vpc_id` - (Optional) The VPC ID to create in.
* `availability_zone` - (Optional) The Availability Zone for the virtual private gateway.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `amazon_side_asn` - (Optional) The Autonomous System Number (ASN) for the Amazon side of the gateway. If you don't specify an ASN, the virtual private gateway is created with the default ASN.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - Amazon Resource Name (ARN) of the VPN Gateway.
* `id` - The ID of the VPN Gateway.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block).

## Import

VPN Gateways can be imported using the `vpn gateway id`, e.g.

```
$ terraform import aws_vpn_gateway.testvpngateway vgw-9a4cacf3
```
