---
layout: "aws"
page_title: "AWS: aws_security_group_rule"
sidebar_current: "docs-aws-resource-security-group-rule"
description: |-
  Provides an security group rule resource.
---

# aws_security_group_rule

Provides a security group rule resource. Represents a single `ingress` or
`egress` group rule, which can be added to external Security Groups.

~> **NOTE on Security Groups and Security Group Rules:** Terraform currently
provides both a standalone Security Group Rule resource (a single `ingress` or
`egress` rule), and a [Security Group resource](security_group.html) with `ingress` and `egress` rules
defined in-line. At this time you cannot use a Security Group with in-line rules
in conjunction with any Security Group Rule resources. Doing so will cause
a conflict of rule settings and will overwrite rules.

~> **NOTE:** Setting `protocol = "all"` or `protocol = -1` with `from_port` and `to_port` will result in the EC2 API creating a security group rule with all ports open. This API behavior cannot be controlled by Terraform and may generate warnings in the future.

~> **NOTE:** Referencing Security Groups across VPC peering has certain restrictions. More information is available in the [VPC Peering User Guide](https://docs.aws.amazon.com/vpc/latest/peering/vpc-peering-security-groups.html).

## Example Usage

Basic usage

```hcl
resource "aws_security_group_rule" "allow_all" {
  type            = "ingress"
  from_port       = 0
  to_port         = 65535
  protocol        = "tcp"
  cidr_blocks     = ["0.0.0.0/0"]
  prefix_list_ids = ["pl-12c4e678"]

  security_group_id = "sg-123456"
}
```

## Argument Reference

The following arguments are supported:

* `type` - (Required) The type of rule being created. Valid options are `ingress` (inbound)
or `egress` (outbound).
* `cidr_blocks` - (Optional) List of CIDR blocks. Cannot be specified with `source_security_group_id`.
* `ipv6_cidr_blocks` - (Optional) List of IPv6 CIDR blocks.
* `prefix_list_ids` - (Optional) List of prefix list IDs (for allowing access to VPC endpoints).
Only valid with `egress`.
* `from_port` - (Required) The start port (or ICMP type number if protocol is "icmp").
* `protocol` - (Required) The protocol. If not icmp, tcp, udp, or all use the [protocol number](https://www.iana.org/assignments/protocol-numbers/protocol-numbers.xhtml)
* `security_group_id` - (Required) The security group to apply this rule to.
* `source_security_group_id` - (Optional) The security group id to allow access to/from,
     depending on the `type`. Cannot be specified with `cidr_blocks`.
* `self` - (Optional) If true, the security group itself will be added as
     a source to this ingress rule.
* `to_port` - (Required) The end port (or ICMP code if protocol is "icmp").
* `description` - (Optional) Description of the rule.

## Usage with prefix list IDs

Prefix list IDs are manged by AWS internally. Prefix list IDs
are associated with a prefix list name, or service name, that is linked to a specific region.
Prefix list IDs are exported on VPC Endpoints, so you can use this format:

```hcl
resource "aws_security_group_rule" "allow_all" {
  type              = "egress"
  to_port           = 0
  protocol          = "-1"
  prefix_list_ids   = ["${aws_vpc_endpoint.my_endpoint.prefix_list_id}"]
  from_port         = 0
  security_group_id = "sg-123456"
}

# ...
resource "aws_vpc_endpoint" "my_endpoint" {
  # ...
}
```

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the security group rule
* `type` - The type of rule, `ingress` or `egress`
* `from_port` - The start port (or ICMP type number if protocol is "icmp")
* `to_port` - The end port (or ICMP code if protocol is "icmp")
* `protocol` – The protocol used
* `description` – Description of the rule

## Import

Security Group Rules can be imported using the `security_group_id`, `type`, `protocol`, `from_port`, `to_port`, and source(s)/destination(s) (e.g. `cidr_block`) separated by underscores (`_`). All parts are required.

Not all rule permissions (e.g., not all of a rule's CIDR blocks) need to be imported for Terraform to manage rule permissions. However, importing some of a rule's permissions but not others, and then making changes to the rule will result in the creation of an additional rule to capture the updated permissions. Rule permissions that were not imported are left intact in the original rule.

### Examples

Import an ingress rule in security group `sg-6e616f6d69` for TCP port 8000 with an IPv4 destination CIDR of `10.0.3.0/24`:

```console
$ terraform import aws_security_group_rule.ingress sg-6e616f6d69_ingress_tcp_8000_8000_10.0.3.0/24
```

Import a rule with various IPv4 and IPv6 source CIDR blocks:

```console
$ terraform import aws_security_group_rule.ingress sg-4973616163_ingress_tcp_100_121_10.1.0.0/16_2001:db8::/48_10.2.0.0/16_2002:db8::/48
```

Import a rule, applicable to all ports, with a protocol other than TCP/UDP/ICMP/ALL, e.g., Multicast Transport Protocol (MTP), using the IANA protocol number, e.g., 92.

```console
$ terraform import aws_security_group_rule.ingress sg-6777656e646f6c796e_ingress_92_0_65536_10.0.3.0/24_10.0.4.0/24
```

Import an egress rule with a prefix list ID destination:

```console
$ terraform import aws_security_group_rule.egress sg-62726f6479_egress_tcp_8000_8000_pl-6469726b
```

Import a rule applicable to all protocols and ports with a security group source:

```console
$ terraform import aws_security_group_rule.ingress_rule sg-7472697374616e_ingress_all_0_65536_sg-6176657279
```

Import a rule that has itself and an IPv6 CIDR block as sources:

```console
$ terraform import aws_security_group_rule.rule_name sg-656c65616e6f72_ingress_tcp_80_80_self_2001:db8::/48
```
