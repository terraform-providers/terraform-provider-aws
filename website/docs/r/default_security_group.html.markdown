---
subcategory: "VPC"
layout: "aws"
page_title: "AWS: aws_default_security_group"
description: |-
  Manage the default Security Group resource.
---

# Resource: aws_default_security_group

Provides a resource to manage the default AWS Security Group.

For EC2 Classic accounts, each region comes with a Default Security Group.
Additionally, each VPC created in AWS comes with a Default Security Group that can be managed, but not
destroyed. **This is an advanced resource**, and has special caveats to be aware
of when using it. Please read this document in its entirety before using this
resource.

The `aws_default_security_group` behaves differently from normal resources, in that
Terraform does not _create_ this resource, but instead "adopts" it
into management. We can do this because these default security groups cannot be
destroyed, and are created with a known set of default ingress/egress rules.

When Terraform first adopts the Default Security Group, it **immediately removes all
ingress and egress rules in the Security Group**. It then proceeds to create any rules specified in the
configuration. This step is required so that only the rules specified in the
configuration are created.

This resource treats its inline rules as absolute; only the rules defined
inline are created, and any additions/removals external to this resource will
result in diff shown. For these reasons, this resource is incompatible with the
`aws_security_group_rule` resource.

For more information about Default Security Groups, see the AWS Documentation on
[Default Security Groups][aws-default-security-groups].

## Basic Example Usage, with default rules

The following config gives the Default Security Group the same rules that AWS
provides by default, but pulls the resource under management by Terraform. This means that
any ingress or egress rules added or changed will be detected as drift.

```hcl
resource "aws_vpc" "mainvpc" {
  cidr_block = "10.1.0.0/16"
}

resource "aws_default_security_group" "default" {
  vpc_id = aws_vpc.mainvpc.id

  ingress {
    protocol  = -1
    self      = true
    from_port = 0
    to_port   = 0
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}
```

## Example config to deny all Egress traffic, allowing Ingress

The following denies all Egress traffic by omitting any `egress` rules, while
including the default `ingress` rule to allow all traffic.

```hcl
resource "aws_vpc" "mainvpc" {
  cidr_block = "10.1.0.0/16"
}

resource "aws_default_security_group" "default" {
  vpc_id = aws_vpc.mainvpc.id

  ingress {
    protocol  = -1
    self      = true
    from_port = 0
    to_port   = 0
  }
}
```

## Argument Reference

The arguments of an `aws_default_security_group` differ slightly from `aws_security_group`
resources. Namely, the `name` argument is computed, and the `name_prefix` attribute
removed. The following arguments are still supported:

* `ingress` - (Optional) Can be specified multiple times for each ingress rule. Each ingress block supports fields documented [below](#ingress-blocks).
* `egress` - (Optional, VPC only) Can be specified multiple times for each egress rule. Each egress block supports fields documented [below](#egress-blocks).
* `vpc_id` - (Optional, Forces new resource) The VPC ID. **Note that changing the `vpc_id` will _not_ restore any default security group rules that were modified, added, or removed.** It will be left in its current state
* `tags` - (Optional) A map of tags to assign to the resource.

### `ingress` Block

* `cidr_blocks` - (Optional) List of CIDR blocks.
* `description` - (Optional) Description of this ingress rule.
* `from_port` - (Required) The start port (or ICMP type number if protocol is "icmp" or "icmpv6")
* `ipv6_cidr_blocks` - (Optional) List of IPv6 CIDR blocks.
* `prefix_list_ids` - (Optional) List of prefix list IDs.
* `protocol` - (Required) The protocol. If you select a protocol of "-1" (semantically equivalent to `"all"`, which is not a valid value here), you must specify a "from_port" and "to_port" equal to 0. If not icmp, icmpv6, tcp, udp, or "-1" use the [protocol number](https://www.iana.org/assignments/protocol-numbers/protocol-numbers.xhtml)
* `security_groups` - (Optional) List of security group Group Names if using EC2-Classic, or Group IDs if using a VPC.
* `self` - (Optional) If true, the security group itself will be added as a source to this ingress rule.
* `to_port` - (Required) The end range port (or ICMP code if protocol is "icmp").

### `egress` Block

* `cidr_blocks` - (Optional) List of CIDR blocks.
* `description` - (Optional) Description of this egress rule.
* `from_port` - (Required) The start port (or ICMP type number if protocol is "icmp")
* `ipv6_cidr_blocks` - (Optional) List of IPv6 CIDR blocks.
* `prefix_list_ids` - (Optional) List of prefix list IDs (for allowing access to VPC endpoints)
* `protocol` - (Required) The protocol. If you select a protocol of "-1" (semantically equivalent to `"all"`, which is not a valid value here), you must specify a "from_port" and "to_port" equal to 0. If not icmp, tcp, udp, or "-1" use the [protocol number](https://www.iana.org/assignments/protocol-numbers/protocol-numbers.xhtml)
* `security_groups` - (Optional) List of security group Group Names if using EC2-Classic, or Group IDs if using a VPC.
* `self` - (Optional) If true, the security group itself will be added as a source to this egress rule.
* `to_port` - (Required) The end range port (or ICMP code if protocol is "icmp").


## Usage

With the exceptions mentioned above, `aws_default_security_group` should
identical behavior to `aws_security_group`. Please consult [AWS_SECURITY_GROUP](/docs/providers/aws/r/security_group.html)
for further usage documentation.

### Removing `aws_default_security_group` from your configuration

Each AWS VPC (or region, if using EC2 Classic) comes with a Default Security
Group that cannot be deleted. The `aws_default_security_group` allows you to
manage this Security Group, but Terraform cannot destroy it. Removing this resource
from your configuration will remove it from your statefile and management, but
will not destroy the Security Group. All ingress or egress rules will be left as
they are at the time of removal. You can resume managing them via the AWS Console.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the security group
* `arn` - The ARN of the security group
* `vpc_id` - The VPC ID.
* `owner_id` - The owner ID.
* `name` - The name of the security group
* `description` - The description of the security group

[aws-default-security-groups]: http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/using-network-security.html#default-security-group
