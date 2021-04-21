---
subcategory: "EC2"
layout: "aws"
page_title: "AWS: aws_ec2_transit_gateway_multicast_domain"
description: |-
  Manages an EC2 Transit Gateway Multicast Domain
---

# Resource: aws_ec2_transit_gateway_multicast_domain

Manages an EC2 Transit Gateway Multicast Domain.

## Example Usage

```hcl
resource "aws_ec2_transit_gateway_multicast_domain" "main" {
  transit_gateway_id = aws_ec2_transit_gateway.main.id
  
  association {
    transit_gateway_attachment_id = aws_ec2_transit_gateway_vpc_attachment.main.id
    subnet_ids                    = [aws_subnet.main.id]
  }

  members {
    group_ip_address      = "224.0.0.1"
    network_interface_ids = [aws_instance.main.primary_network_interface_id]
  }

  sources {
    group_ip_address      = "224.0.0.1"
    network_interface_ids = [aws_instance.main.primary_network_interface_id]
  }
}
```

## Argument Reference

The following arguments are supported:

* `transit_gateway_id` - (Required, Forces new resource) EC2 Transit Gateway Identifier. The target resource must have 
    `multicast_support = "enable"`.
* `association` - (Optional) Can be specified multiple times for different EC2 Transit Gateway Attachments. Each 
    association block supports the fields documented below. This argument is processed in 
    [attribute-as-blocks mode](/docs/configuration/attr-as-blocks.html).
* `members` - (Optional) Can be specified multiple times for different Group IP Addresses. Each members block supports 
    the fields documented below. This argument is processed in 
    [attribute-as-blocks mode](/docs/configuration/attr-as-blocks.html).
* `sources` - (Optional) Can be specified multiple times for different Group IP Addresses. Each members block supports 
    the fields documented below. This argument is processed in 
    [attribute-as-blocks mode](/docs/configuration/attr-as-blocks.html).
* `tags` - (Optional) A mapping of tags to assign to the resource.

The `association` block supports:

* `transit_gateway_attachment_id` - (Required) EC2 Transit Gateway Attachment Identifier.
* `subnet_ids` - (Required, Minimum items: 1) List of subnets identifiers to associate. The listed subnets must reside
    within the specified EC2 Transit Gateway Attachment.
    
The `members` and `sources` blocks support:

* `group_ip_address` - (Required) Multicast Group IP address. Must be valid IPv4 or IPv6 IP Address in the 224.0.0.0/4 
    or ff00::/8 CIDR range.
* `network_interface_ids` - (Required, Minimum items: 1) List of Network Interface Identifiers to create 
    Multicast Group for.
