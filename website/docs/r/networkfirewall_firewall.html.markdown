---
subcategory: "Network Firewall"
layout: "aws"
page_title: "AWS: aws_networkfirewall_firewall"
description: |-
  Provides an AWS Network Firewall Firewall resource.
---

# Resource: aws_networkfirewall_firewall

Provides an AWS Network Firewall Firewall Resource

## Example Usage

```hcl
resource "aws_networkfirewall_firewall" "example" {
  name                = "example"
  firewall_policy_arn = aws_networkfirewall_firewall_policy.example.arn
  vpc_id              = aws_vpc.example.id
  subnet_mapping {
    subnet_id = aws_subnet.example.id
  }

  tags = {
    Tag1 = "Value1"
    Tag2 = "Value2"
  }
}
```

## Argument Reference

The following arguments are supported:

* `delete_protection` - (Optional) A boolean flag indicating whether it is possible to delete the firewall. Defaults to `false`.

* `description` - (Optional) A friendly description of the firewall.

* `firewall_policy_arn` - (Required) The Amazon Resource Name (ARN) of the VPC Firewall policy.

* `firewall_policy_change_protection` - (Option) A boolean flag indicating whether it is possible to change the associated firewall policy. Defaults to `false`.

* `name` - (Required, Forces new resource) A friendly name of the firewall.

* `subnet_change_protection` - (Optional) A boolean flag indicating whether it is possible to change the associated subnet(s). Defaults to `false`.

* `subnet_mapping` - (Required) Set of configuration blocks describing the public subnets. Each subnet must belong to a different Availability Zone in the VPC. AWS Network Firewall creates a firewall endpoint in each subnet. See [Subnet Mapping](#subnet-mapping) below for details.

* `tags` - (Optional) The key:value pairs to associate with the resource.

* `vpc_id` - (Required, Forces new resource) The unique identifier of the VPC where AWS Network Firewall should create the firewall.

### Subnet Mapping

The `subnet_mapping` block supports the following arguments:

* `subnet_id` - (Required) The unique identifier for the subnet.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The Amazon Resource Name (ARN) that identifies the firewall.

* `arn` - The Amazon Resource Name (ARN) that identifies the firewall.

* `firewall_status` - Nested list of information about the current status of the firewall.
    * `sync_states` - Set of subnets configured for use by the firewall.
        * `attachment` - Nested list describing the attachment status of the firewall's association with a single VPC subnet.
            * `endpoint_id` - The identifier of the firewall endpoint that AWS Network Firewall has instantiated in the subnet. You use this to identify the firewall endpoint in the VPC route tables, when you redirect the VPC traffic through the endpoint.
            * `subnet_id` - The unique identifier of the subnet that you've specified to be used for a firewall endpoint.
        * `availability_zone` - The Availability Zone where the subnet is configured.

* `update_token` - A string token used when updating a firewall.

## Import

Network Firewall Firewalls can be imported using their `ARN`.

```
$ terraform import aws_networkfirewall_firewall.example arn:aws:network-firewall:us-west-1:123456789012:firewall/example
```
