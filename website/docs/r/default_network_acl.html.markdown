---
subcategory: "VPC"
layout: "aws"
page_title: "AWS: aws_default_network_acl"
description: |-
  Manage a default network ACL.
---

# Resource: aws_default_network_acl

Provides a resource to manage a VPC's default network ACL. This resource can manage the default network ACL of the default or a non-default VPC.

~> **NOTE:** This is an advanced resource with special caveats. Please read this document in its entirety before using this resource. The `aws_default_network_acl` behaves differently from normal resources. Terraform does not _create_ this resource but instead attempts to "adopt" it into management.

Every VPC has a default network ACL that can be managed but not destroyed. When Terraform first adopts the Default Network ACL, it **immediately removes all rules in the ACL**. It then proceeds to create any rules specified in the configuration. This step is required so that only the rules specified in the configuration are created.

This resource treats its inline rules as absolute; only the rules defined inline are created, and any additions/removals external to this resource will result in diffs being shown. For these reasons, this resource is incompatible with the `aws_network_acl_rule` resource.

For more information about Network ACLs, see the AWS Documentation on [Network ACLs][aws-network-acls].

## Example Usage

### Basic Example

The following config gives the Default Network ACL the same rules that AWS includes but pulls the resource under management by Terraform. This means that any ACL rules added or changed will be detected as drift.

```hcl
resource "aws_vpc" "mainvpc" {
  cidr_block = "10.1.0.0/16"
}

resource "aws_default_network_acl" "default" {
  default_network_acl_id = aws_vpc.mainvpc.default_network_acl_id

  ingress {
    protocol   = -1
    rule_no    = 100
    action     = "allow"
    cidr_block = aws_vpc.mainvpc.cidr_block
    from_port  = 0
    to_port    = 0
  }

  egress {
    protocol   = -1
    rule_no    = 100
    action     = "allow"
    cidr_block = "0.0.0.0/0"
    from_port  = 0
    to_port    = 0
  }
}
```

### Example: Deny All Egress Traffic, Allow Ingress

The following denies all Egress traffic by omitting any `egress` rules, while including the default `ingress` rule to allow all traffic.

```hcl
resource "aws_vpc" "mainvpc" {
  cidr_block = "10.1.0.0/16"
}

resource "aws_default_network_acl" "default" {
  default_network_acl_id = aws_vpc.mainvpc.default_network_acl_id

  ingress {
    protocol   = -1
    rule_no    = 100
    action     = "allow"
    cidr_block = aws_default_vpc.mainvpc.cidr_block
    from_port  = 0
    to_port    = 0
  }
}
```

### Example: Deny All Traffic To Any Subnet In The Default Network ACL

This config denies all traffic in the Default ACL. This can be useful if you want to lock down the VPC to force all resources to assign a non-default ACL.

```hcl
resource "aws_vpc" "mainvpc" {
  cidr_block = "10.1.0.0/16"
}

resource "aws_default_network_acl" "default" {
  default_network_acl_id = aws_vpc.mainvpc.default_network_acl_id

  # no rules defined, deny all traffic in this ACL
}
```

### Managing Subnets In A Default Network ACL

Within a VPC, all Subnets must be associated with a Network ACL. In order to "delete" the association between a Subnet and a non-default Network ACL, the association is destroyed by replacing it with an association between the Subnet and the Default ACL instead.

When managing the Default Network ACL, you cannot "remove" Subnets. Instead, they must be reassigned to another Network ACL, or the Subnet itself must be destroyed. Because of these requirements, removing the `subnet_ids` attribute from the configuration of a `aws_default_network_acl` resource may result in a reoccurring plan, until the Subnets are reassigned to another Network ACL or are destroyed.

Because Subnets are by default associated with the Default Network ACL, any non-explicit association will show up as a plan to remove the Subnet. For example: if you have a custom `aws_network_acl` with two subnets attached, and you remove the `aws_network_acl` resource, after successfully destroying this resource future plans will show a diff on the managed `aws_default_network_acl`, as those two Subnets have been orphaned by the now destroyed network acl and thus adopted by the Default Network ACL. In order to avoid a reoccurring plan, they will need to be reassigned, destroyed, or added to the `subnet_ids` attribute of the `aws_default_network_acl` entry.

As an alternative to the above, you can also specify the following lifecycle configuration in your `aws_default_network_acl` resource:

```hcl
resource "aws_default_network_acl" "default" {
  # ... other configuration ...

  lifecycle {
    ignore_changes = [subnet_ids]
  }
}
```

### Removing `aws_default_network_acl` From Your Configuration

Each AWS VPC comes with a Default Network ACL that cannot be deleted. The `aws_default_network_acl` allows you to manage this Network ACL, but Terraform cannot destroy it. Removing this resource from your configuration will remove it from your statefile and management, **but will not destroy the Network ACL.** All Subnets associations and ingress or egress rules will be left as they are at the time of removal. You can resume managing them via the AWS Console.

## Argument Reference

The following arguments are required:

* `default_network_acl_id` - (Required) Network ACL ID to manage. This attribute is exported from `aws_vpc`, or manually found via the AWS Console.

The following arguments are optional:

* `egress` - (Optional) Configuration block for an egress rule. Detailed below.
* `ingress` - (Optional) Configuration block for an ingress rule. Detailed below.
* `subnet_ids` - (Optional) List of Subnet IDs to apply the ACL to. See the notes below on managing Subnets in the Default Network ACL
* `tags` - (Optional) Map of tags to assign to the resource.

### egress and ingress

Both the `egress` and `ingress` configuration blocks have the same arguments.

The following arguments are required:

* `action` - (Required) The action to take.
* `from_port` - (Required) The from port to match.
* `protocol` - (Required) The protocol to match. If using the -1 'all' protocol, you must specify a from and to port of 0.
* `rule_no` - (Required) The rule number. Used for ordering.
* `to_port` - (Required) The to port to match.

The following arguments are optional:

* `cidr_block` - (Optional) The CIDR block to match. This must be a valid network mask.
* `icmp_code` - (Optional) The ICMP type code to be used. Default 0.
* `icmp_type` - (Optional) The ICMP type to be used. Default 0.
* `ipv6_cidr_block` - (Optional) The IPv6 CIDR block.

-> For more information on ICMP types and codes, see [Internet Control Message Protocol (ICMP) Parameters](https://www.iana.org/assignments/icmp-parameters/icmp-parameters.xhtml).

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - ARN of the Default Network ACL
* `id` - ID of the Default Network ACL
* `owner_id` - ID of the AWS account that owns the Default Network ACL
* `vpc_id` -  ID of the associated VPC

[aws-network-acls]: http://docs.aws.amazon.com/AmazonVPC/latest/UserGuide/VPC_ACLs.html

## Import

Default Network ACLs can be imported using the `id`, e.g.

```
$ terraform import aws_default_network_acl.sample acl-7aaabd18
```
