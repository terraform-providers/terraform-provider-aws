---
subcategory: "EC2"
layout: "aws"
page_title: "AWS: aws_ec2_client_vpn_endpoint"
description: |-
  Get information on an EC2 Client Vpn Endpoint
---

# Data Source: aws_ec2_client_vpn_endpoint

Get information on an EC2 Client VPN Endpoint.

## Example Usage

### By Filter

```hcl
data "aws_ec2_client_vpn_endpoint" "example" {
  filter {
    name   = "tag:Name"
    values = ["ExampleVpn"]
  }
}
```

### By Identifier

```hcl
data "aws_ec2_client_vpn_endpoint" "example" {
  client_vpn_endpoint_id = "cvpn-endpoint-083cf50d6eb314f21"
}
```

## Argument Reference

The following arguments are supported:

* `filter` - (Optional) One or more configuration blocks containing name-values filters. Detailed below.
* `id` - (Optional) The ID of the Client VPN endpoint.

### filter Argument Reference

* `name` - (Required) Name of the filter.
* `values` - (Required) List of one or more values for the filter.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `arn` -  The ARN of the Client VPN endpoint.
* `authentication_options` - Information about the authentication method to be used to authenticate clients.
* `client_cidr_block` - The IPv4 address range, in CIDR notation, from which to assign client IP addresses.
* `connection_log_options` - Information about the client connection logging options.
* `description` - Description of de Client Vpn Endpoint.
* `dns_name` - The DNS name to be used by clients when establishing their VPN session.
* `dns_servers` - Information about the DNS servers to be used for DNS resolution.
* `security_group_ids` - List VPC security groups associated with the endpoint.
* `server_certificate_arn`- The ARN of the ACM server certificate.
* `split_tunnel` - Indicates whether split-tunnel is enabled on VPN endpoint.
* `tags` - A mapping of tags of the resource.
* `transport_protocol` - The transport protocol to be used by the VPN session.
* `vpc_id` - The VPC Id associated with the endpoint.
* `vpn_port` - The vpn port.
* `vpn_protocol` - The vpn protocol.
