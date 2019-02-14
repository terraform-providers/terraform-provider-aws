---
layout: "aws"
page_title: "AWS: aws_network_interface"
sidebar_current: "docs-aws-datasource-network-interface-x"
description: |-  
  Get information on a Network Interface resource.
---

# aws_network_interface

Use this data source to get information about a Network Interface.

## Example Usage

```hcl
data "aws_network_interface" "bar" {
  id = "eni-01234567"
}
```

## Argument Reference

The following arguments are supported:

* `id` – (Optional) The identifier for the network interface.
* `filter` – (Optional) One or more name/value pairs to filter off of. There are several valid keys, for a full reference, check out [describe-network-interfaces](https://docs.aws.amazon.com/cli/latest/reference/ec2/describe-network-interfaces.html) in the AWS CLI reference.

## Attributes Reference

See the [Network Interface](/docs/providers/aws/r/network_interface.html) for details on the returned attributes.

Additionally, the following attributes are exported:

* `association` - The association information for an Elastic IP address (IPv4) associated with the network interface. See supported fields below.
* `availability_zone` - The Availability Zone.
* `description` - Description of the network interface.
* `interface_type` - The type of interface.
* `ipv6_addresses` - List of IPv6 addresses to assign to the ENI.
* `mac_address` - The MAC address.
* `owner_id` - The AWS account ID of the owner of the network interface.
* `private_dns_name` - The private DNS name.
* `private_ip` - The private IPv4 address of the network interface within the subnet.
* `private_ips` - The private IPv4 addresses associated with the network interface.
* `requester_id` - The ID of the entity that launched the instance on your behalf.
* `security_groups` - The list of security groups for the network interface.
* `subnet_id` - The ID of the subnet.
* `tags` - Any tags assigned to the network interface.
* `vpc_id` - The ID of the VPC.

### `association`

* `allocation_id` - The allocation ID.
* `association_id` - The association ID.
* `ip_owner_id` - The ID of the Elastic IP address owner.
* `public_dns_name` - The public DNS name.
* `public_ip` - The address of the Elastic IP address bound to the network interface.

## Import

Elastic Network Interfaces can be imported using the `id`, e.g.

```
$ terraform import aws_network_interface.test eni-12345
```
