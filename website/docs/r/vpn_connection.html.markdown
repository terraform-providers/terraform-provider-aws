---
subcategory: "VPC"
layout: "aws"
page_title: "AWS: aws_vpn_connection"
description: |-
  Manages an EC2 VPN connection. These objects can be connected to customer gateways, and allow you to establish tunnels between your network and Amazon.
---

# Resource: aws_vpn_connection

Manages an EC2 VPN connection. These objects can be connected to customer gateways, and allow you to establish tunnels between your network and Amazon.

~> **Note:** All arguments including `tunnel1_preshared_key` and `tunnel2_preshared_key` will be stored in the raw state as plain-text.
[Read more about sensitive data in state](https://www.terraform.io/docs/state/sensitive-data.html).

~> **Note:** The CIDR blocks in the arguments `tunnel1_inside_cidr` and `tunnel2_inside_cidr` must have a prefix of /30 and be a part of a specific range.
[Read more about this in the AWS documentation](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_VpnTunnelOptionsSpecification.html).

## Example Usage

### EC2 Transit Gateway

```hcl
resource "aws_ec2_transit_gateway" "example" {}

resource "aws_customer_gateway" "example" {
  bgp_asn    = 65000
  ip_address = "172.0.0.1"
  type       = "ipsec.1"
}

resource "aws_vpn_connection" "example" {
  customer_gateway_id = aws_customer_gateway.example.id
  transit_gateway_id  = aws_ec2_transit_gateway.example.id
  type                = aws_customer_gateway.example.type
}
```

### Virtual Private Gateway

```hcl
resource "aws_vpc" "vpc" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_vpn_gateway" "vpn_gateway" {
  vpc_id = aws_vpc.vpc.id
}

resource "aws_customer_gateway" "customer_gateway" {
  bgp_asn    = 65000
  ip_address = "172.0.0.1"
  type       = "ipsec.1"
}

resource "aws_vpn_connection" "main" {
  vpn_gateway_id      = aws_vpn_gateway.vpn_gateway.id
  customer_gateway_id = aws_customer_gateway.customer_gateway.id
  type                = "ipsec.1"
  static_routes_only  = true
}
```

## Argument Reference

The following arguments are required:

* `customer_gateway_id` - (Required) The ID of the customer gateway.
* `type` - (Required) The type of VPN connection. The only type AWS supports at this time is "ipsec.1".

One of the following arguments is required:

* `transit_gateway_id` - (Optional) The ID of the EC2 Transit Gateway.
* `vpn_gateway_id` - (Optional) The ID of the Virtual Private Gateway.

Other arguments:

* `static_routes_only` - (Optional, Default `false`) Whether the VPN connection uses static routes exclusively. Static routes must be used for devices that don't support BGP.
* `enable_acceleration` - (Optional, Default `false`) Indicate whether to enable acceleration for the VPN connection. Supports only EC2 Transit Gateway.
* `tags` - (Optional) Tags to apply to the connection.
* `local_ipv4_network_cidr` - (Optional, Default `0.0.0.0/0`) The IPv4 CIDR on the customer gateway (on-premises) side of the VPN connection.
* `local_ipv6_network_cidr` - (Optional, Default `::/0`) The IPv6 CIDR on the customer gateway (on-premises) side of the VPN connection.
* `remote_ipv4_network_cidr` - (Optional, Default `0.0.0.0/0`) The IPv4 CIDR on the AWS side of the VPN connection.
* `remote_ipv6_network_cidr` - (Optional, Default `::/0`) The IPv6 CIDR on the customer gateway (on-premises) side of the VPN connection.
* `tunnel_inside_ip_version` - (Optional, Default `ipv4`) Indicate whether the VPN tunnels process IPv4 or IPv6 traffic. Valid values are `ipv4 | ipv6`. `ipv6` Supports only EC2 Transit Gateway.
* `tunnel1_inside_cidr` - (Optional) The CIDR block of the inside IP addresses for the first VPN tunnel. Valid value is a size /30 CIDR block from the 169.254.0.0/16 range.
* `tunnel2_inside_cidr` - (Optional) The CIDR block of the inside IP addresses for the second VPN tunnel. Valid value is a size /30 CIDR block from the 169.254.0.0/16 range.
* `tunnel1_inside_ipv6_cidr` - (Optional) The range of inside IPv6 addresses for the first VPN tunnel. Supports only EC2 Transit Gateway. Valid value is a size /126 CIDR block from the local fd00::/8 range.
* `tunnel2_inside_ipv6_cidr` - (Optional) The range of inside IPv6 addresses for the second VPN tunnel. Supports only EC2 Transit Gateway. Valid value is a size /126 CIDR block from the local fd00::/8 range.
* `tunnel1_preshared_key` - (Optional) The preshared key of the first VPN tunnel. The preshared key must be between 8 and 64 characters in length and cannot start with zero(0). Allowed characters are alphanumeric characters, periods(.) and underscores(_).
* `tunnel2_preshared_key` - (Optional) The preshared key of the second VPN tunnel. The preshared key must be between 8 and 64 characters in length and cannot start with zero(0). Allowed characters are alphanumeric characters, periods(.) and underscores(_).
* `tunnel1_dpd_timeout_action` - (Optional, Default `clear`) The action to take after DPD timeout occurs for the first VPN tunnel. Specify restart to restart the IKE initiation. Specify clear to end the IKE session. Valid values are `clear | none | restart`.
* `tunnel2_dpd_timeout_action` - (Optional, Default `clear`) The action to take after DPD timeout occurs for the second VPN tunnel. Specify restart to restart the IKE initiation. Specify clear to end the IKE session. Valid values are `clear | none | restart`.
* `tunnel1_dpd_timeout_seconds` - (Optional, Default `30`) The number of seconds after which a DPD timeout occurs for the first VPN tunnel. Valid value is equal or higher than `30`.
* `tunnel2_dpd_timeout_seconds` - (Optional, Default `30`) The number of seconds after which a DPD timeout occurs for the second VPN tunnel. Valid value is equal or higher than `30`.
* `tunnel1_ike_versions` - (Optional) The IKE versions that are permitted for the first VPN tunnel. Valid values are `ikev1 | ikev2`.
* `tunnel2_ike_versions` - (Optional) The IKE versions that are permitted for the second VPN tunnel. Valid values are `ikev1 | ikev2`.
* `tunnel1_phase1_dh_group_numbers` - (Optional) List of one or more Diffie-Hellman group numbers that are permitted for the first VPN tunnel for phase 1 IKE negotiations. Valid values are ` 2 | 14 | 15 | 16 | 17 | 18 | 19 | 20 | 21 | 22 | 23 | 24`.
* `tunnel2_phase1_dh_group_numbers` - (Optional) List of one or more Diffie-Hellman group numbers that are permitted for the second VPN tunnel for phase 1 IKE negotiations. Valid values are ` 2 | 14 | 15 | 16 | 17 | 18 | 19 | 20 | 21 | 22 | 23 | 24`.
* `tunnel1_phase1_encryption_algorithms` - (Optional) List of one or more encryption algorithms that are permitted for the first VPN tunnel for phase 1 IKE negotiations. Valid values are `AES128 | AES256 | AES128-GCM-16 | AES256-GCM-16`.
* `tunnel2_phase1_encryption_algorithms` - (Optional) List of one or more encryption algorithms that are permitted for the second VPN tunnel for phase 1 IKE negotiations. Valid values are `AES128 | AES256 | AES128-GCM-16 | AES256-GCM-16`.
* `tunnel1_phase1_integrity_algorithms` - (Optional) One or more integrity algorithms that are permitted for the first VPN tunnel for phase 1 IKE negotiations. Valid values are `SHA1 | SHA2-256 | SHA2-384 | SHA2-512`.
* `tunnel2_phase1_integrity_algorithms` - (Optional) One or more integrity algorithms that are permitted for the second VPN tunnel for phase 1 IKE negotiations. Valid values are `SHA1 | SHA2-256 | SHA2-384 | SHA2-512`.
* `tunnel1_phase1_lifetime_seconds` - (Optional, Default `28800`) The lifetime for phase 1 of the IKE negotiation for the first VPN tunnel, in seconds. Valid value is between `900` and `28800`.
* `tunnel2_phase1_lifetime_seconds` - (Optional, Default `28800`) The lifetime for phase 1 of the IKE negotiation for the second VPN tunnel, in seconds. Valid value is between `900` and `28800`.
* `tunnel1_phase2_dh_group_numbers` - (Optional) List of one or more Diffie-Hellman group numbers that are permitted for the first VPN tunnel for phase 2 IKE negotiations. Valid values are `2 | 5 | 14 | 15 | 16 | 17 | 18 | 19 | 20 | 21 | 22 | 23 | 24`.
* `tunnel2_phase2_dh_group_numbers` - (Optional) List of one or more Diffie-Hellman group numbers that are permitted for the second VPN tunnel for phase 2 IKE negotiations. Valid values are `2 | 5 | 14 | 15 | 16 | 17 | 18 | 19 | 20 | 21 | 22 | 23 | 24`.
* `tunnel1_phase2_encryption_algorithms` - (Optional) List of one or more encryption algorithms that are permitted for the first VPN tunnel for phase 2 IKE negotiations. Valid values are `AES128 | AES256 | AES128-GCM-16 | AES256-GCM-16`.
* `tunnel2_phase2_encryption_algorithms` - (Optional) List of one or more encryption algorithms that are permitted for the second VPN tunnel for phase 2 IKE negotiations. Valid values are `AES128 | AES256 | AES128-GCM-16 | AES256-GCM-16`.
* `tunnel1_phase2_integrity_algorithms` - (Optional) List of one or more integrity algorithms that are permitted for the first VPN tunnel for phase 2 IKE negotiations. Valid values are `SHA1 | SHA2-256 | SHA2-384 | SHA2-512`.
* `tunnel2_phase2_integrity_algorithms` - (Optional) List of one or more integrity algorithms that are permitted for the second VPN tunnel for phase 2 IKE negotiations. Valid values are `SHA1 | SHA2-256 | SHA2-384 | SHA2-512`.
* `tunnel1_phase2_lifetime_seconds` - (Optional, Default `3600`) The lifetime for phase 2 of the IKE negotiation for the first VPN tunnel, in seconds. Valid value is between `900` and `3600`.
* `tunnel2_phase2_lifetime_seconds` - (Optional, Default `3600`) The lifetime for phase 2 of the IKE negotiation for the second VPN tunnel, in seconds. Valid value is between `900` and `3600`.
* `tunnel1_rekey_fuzz_percentage` - (Optional, Default `100`) The percentage of the rekey window for the first VPN tunnel (determined by `tunnel1_rekey_margin_time_seconds`) during which the rekey time is randomly selected. Valid value is between `0` and `100`.
* `tunnel2_rekey_fuzz_percentage` - (Optional, Default `100`) The percentage of the rekey window for the second VPN tunnel (determined by `tunnel2_rekey_margin_time_seconds`) during which the rekey time is randomly selected. Valid value is between `0` and `100`.
* `tunnel1_rekey_margin_time_seconds` - (Optional, Default `540`) The margin time, in seconds, before the phase 2 lifetime expires, during which the AWS side of the first VPN connection performs an IKE rekey. The exact time of the rekey is randomly selected based on the value for `tunnel1_rekey_fuzz_percentage`. Valid value is between `60` and half of `tunnel1_phase2_lifetime_seconds`.
* `tunnel2_rekey_margin_time_seconds` - (Optional, Default `540`) The margin time, in seconds, before the phase 2 lifetime expires, during which the AWS side of the second VPN connection performs an IKE rekey. The exact time of the rekey is randomly selected based on the value for `tunnel2_rekey_fuzz_percentage`. Valid value is between `60` and half of `tunnel2_phase2_lifetime_seconds`.
* `tunnel1_replay_window_size` - (Optional, Default `1024`) The number of packets in an IKE replay window for the first VPN tunnel. Valid value is between `64` and `2048`.
* `tunnel2_replay_window_size` - (Optional, Default `1024`) The number of packets in an IKE replay window for the second VPN tunnel. Valid value is between `64` and `2048`.
* `tunnel1_startup_action` - (Optional, Default `add`) The action to take when the establishing the tunnel for the first VPN connection. By default, your customer gateway device must initiate the IKE negotiation and bring up the tunnel. Specify start for AWS to initiate the IKE negotiation. Valid values are `add | start`.
* `tunnel2_startup_action` - (Optional, Default `add`) The action to take when the establishing the tunnel for the second VPN connection. By default, your customer gateway device must initiate the IKE negotiation and bring up the tunnel. Specify start for AWS to initiate the IKE negotiation. Valid values are `add | start`.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - Amazon Resource Name (ARN) of the VPN Connection.
* `id` - The amazon-assigned ID of the VPN connection.
* `customer_gateway_configuration` - The configuration information for the VPN connection's customer gateway (in the native XML format).
* `customer_gateway_id` - The ID of the customer gateway to which the connection is attached.
* `static_routes_only` - Whether the VPN connection uses static routes exclusively.
* `transit_gateway_attachment_id` - When associated with an EC2 Transit Gateway (`transit_gateway_id` argument), the attachment ID. See also the [`aws_ec2_tag` resource](/docs/providers/aws/r/ec2_tag.html) for tagging the EC2 Transit Gateway VPN Attachment.
* `tunnel1_address` - The public IP address of the first VPN tunnel.
* `tunnel1_cgw_inside_address` - The RFC 6890 link-local address of the first VPN tunnel (Customer Gateway Side).
* `tunnel1_vgw_inside_address` - The RFC 6890 link-local address of the first VPN tunnel (VPN Gateway Side).
* `tunnel1_preshared_key` - The preshared key of the first VPN tunnel.
* `tunnel1_bgp_asn` - The bgp asn number of the first VPN tunnel.
* `tunnel1_bgp_holdtime` - The bgp holdtime of the first VPN tunnel.
* `tunnel2_address` - The public IP address of the second VPN tunnel.
* `tunnel2_cgw_inside_address` - The RFC 6890 link-local address of the second VPN tunnel (Customer Gateway Side).
* `tunnel2_vgw_inside_address` - The RFC 6890 link-local address of the second VPN tunnel (VPN Gateway Side).
* `tunnel2_preshared_key` - The preshared key of the second VPN tunnel.
* `tunnel2_bgp_asn` - The bgp asn number of the second VPN tunnel.
* `tunnel2_bgp_holdtime` - The bgp holdtime of the second VPN tunnel.
* `vpn_gateway_id` - The ID of the virtual private gateway to which the connection is attached.


## Import

VPN Connections can be imported using the `vpn connection id`, e.g.

```
$ terraform import aws_vpn_connection.testvpnconnection vpn-40f41529
```
