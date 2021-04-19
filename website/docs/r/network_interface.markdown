---
subcategory: "VPC"
layout: "aws"
page_title: "AWS: aws_network_interface"
description: |-
  Provides an Elastic network interface (ENI) resource.
---

# Resource: aws_network_interface

Provides an Elastic network interface (ENI) resource.

## Example Usage

```terraform
resource "aws_network_interface" "test" {
  subnet_id       = aws_subnet.public_a.id
  private_ips     = ["10.0.0.50"]
  security_groups = [aws_security_group.web.id]

  attachment {
    instance     = aws_instance.test.id
    device_index = 1
  }
}
```

## Argument Reference

The following arguments are supported:

* `subnet_id` - (Required) Subnet ID to create the ENI in.
* `description` - (Optional) A description for the network interface.
* `private_ips` - (Optional) List of private IPs to assign to the ENI without regard to order.
* `private_ips_count` - (Optional) Number of secondary private IPs to assign to the ENI. The total number of private IPs will be 1 + `private_ips_count`, as a primary private IP will be assiged to an ENI by default.
* `private_ip_list` - (Optional) List of private IPs to assign to the ENI in sequential order. Requires setting `private_ip_list_enable` to `true`.
* `private_ip_list_enable` - (Optional) Whether `private_ip_list` is allowed and controls the IPs to assign to the ENI and `private_ips` and `private_ips_count` become read-only. Default false.
* `ipv6_addresses` - (Optional) One or more specific IPv6 addresses from the IPv6 CIDR block range of your subnet. Addresses are assigned without regard to order. You can't use this option if you're specifying `ipv6_address_count`.
* `ipv6_address_count` - (Optional) The number of IPv6 addresses to assign to a network interface. You can't use this option if specifying specific `ipv6_addresses`. If your subnet has the AssignIpv6AddressOnCreation attribute set to `true`, you can specify `0` to override this setting.
* `ipv6_address_list` - (Optional) List of private IPs to assign to the ENI in sequential order.
* `ipv6_address_list_enable` - (Optional) Whether `ipv6_addreses_list` is allowed and controls the IPs to assign to the ENI and `ipv6_addresses` and `ipv6_addresses_count` become read-only. Default false.
* `security_groups` - (Optional) List of security group IDs to assign to the ENI.
* `attachment` - (Optional) Block to define the attachment of the ENI. Documented below.
* `source_dest_check` - (Optional) Whether to enable source destination checking for the ENI. Default true.
* `tags` - (Optional) A map of tags to assign to the resource.

The `attachment` block supports:

* `instance` - (Required) ID of the instance to attach to.
* `device_index` - (Required) Integer to define the devices index.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the network interface.
* `subnet_id` - Subnet ID the ENI is in.
* `mac_address` - The MAC address of the network interface.
* `private_dns_name` - The private DNS name of the network interface (IPv4).
* `description` - A description for the network interface.
* `private_ips` - Set of private IPs assigned to the ENI.
* `security_groups` - List of security groups attached to the ENI.
* `attachment` - Block defining the attachment of the ENI.
* `source_dest_check` - Whether source destination checking is enabled
* `tags` - Tags assigned to the ENI.

## Managing Multiple IPs on a Network Interface

By default, private IPs are managed through the `private_ips` and `private_ips_count` arguments which manage IPs as a set of IPs that are configured without regard to order. For a new network interface, the same primary IP address is consistently selected from a given a set of addresses, regardless of the order provided. However, modifications of the set of addresses of an existing interface will not alter the current primary IP address unless it has been removed from the set.

In order to manage the private Ips as a sequentially ordered list instead, configure `private_ip_list_enabled` to `true` and use `private_ip_list` to manage the IPs. This will disable the `private_ips` and `private_ips_count` settings, which must be removed from the config file but are still exported. Note that changing the first address of `private_ip_list`, which is the primary, always requires a new interface.

If you are managing a specific set or list of Ips, instead of just using `private_ips_count`, here is a workflow for also leveraging `private_ips_count` to have AWS automatically assign additional IP addresses:

* Comment out these settings:

    * `private_ips`
    * `private_ip_list`
    * `private_ip_list_enabled`

* Set the desired `private_ips_count`. Note that this count is for the number of secondaries. The primary is not included in this count.
* Apply to assign the extra Ips
* Remove `private_ips_count` and restore your settings from the first step
* Add the new Ips to your current settings
* Apply again to update the stored state

This process can also be used to remove IP addresses in addition to the option of manually removing them. Adding IP addresses in a manual fashion is more difficult because it requires knowledge of which addresses are available.

## Import

Network Interfaces can be imported using the `id`, e.g.

```
$ terraform import aws_network_interface.test eni-e5aa89a3
```
