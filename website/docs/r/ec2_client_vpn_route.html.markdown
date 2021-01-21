---
subcategory: "EC2"
layout: "aws"
page_title: "AWS: aws_ec2_client_vpn_route"
description: |-
  Provides additional routes for AWS Client VPN endpoints.
---

# Resource: aws_ec2_client_vpn_route

Provides additional routes for AWS Client VPN endpoints. For more information on usage, please see the
[AWS Client VPN Administrator's Guide](https://docs.aws.amazon.com/vpn/latest/clientvpn-admin/what-is.html).

## Example Usage

```hcl
resource "aws_ec2_client_vpn_route" "example" {
  client_vpn_endpoint_id = aws_ec2_client_vpn_endpoint.example.id
  destination_cidr_block = "0.0.0.0/0"
  target_vpc_subnet_id   = aws_ec2_client_vpn_network_association.example.subnet_id
}

resource "aws_ec2_client_vpn_network_association" "example" {
  client_vpn_endpoint_id = aws_ec2_client_vpn_endpoint.example.id
  subnet_id              = aws_subnet.example.id
}

resource "aws_ec2_client_vpn_endpoint" "example" {
  description            = "Example Client VPN endpoint"
  server_certificate_arn = aws_acm_certificate.example.arn
  client_cidr_block      = "10.0.0.0/16"

  authentication_options {
    type                       = "certificate-authentication"
    root_certificate_chain_arn = aws_acm_certificate.example.arn
  }

  connection_log_options {
    enabled = false
  }
}
```

## Argument Reference

The following arguments are supported:

* `client_vpn_endpoint_id` - (Required) The ID of the Client VPN endpoint.
* `destination_cidr_block` - (Required) The IPv4 address range, in CIDR notation, of the route destination.
* `description` - (Optional) A brief description of the authorization rule.
* `target_vpc_subnet_id` - (Required) The ID of the Subnet to route the traffic through. It must already be attached to the Client VPN.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the Client VPN endpoint.
* `origin` - Indicates how the Client VPN route was added. Will be `add-route` for routes created by this resource.
* `type` - The type of the route.

## Import

AWS Client VPN routes can be imported using the endpoint ID, target subnet ID, and destination CIDR block. All values are separated by a `,`.

```
$ terraform import aws_ec2_client_vpn_route.example cvpn-endpoint-1234567890abcdef,subnet-9876543210fedcba,10.1.0.0/24
```
