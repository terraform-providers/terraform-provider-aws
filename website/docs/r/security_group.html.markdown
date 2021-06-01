---
subcategory: "VPC"
layout: "aws"
page_title: "AWS: aws_security_group"
description: |-
  Provides a security group resource.
---

# Resource: aws_security_group

Provides a security group resource.

~> **NOTE on Security Groups and Security Group Rules:** Terraform currently
provides both a standalone [Security Group Rule resource](security_group_rule.html) (a single `ingress` or
`egress` rule), and a Security Group resource with `ingress` and `egress` rules
defined in-line. At this time you cannot use a Security Group with in-line rules
in conjunction with any Security Group Rule resources. Doing so will cause
a conflict of rule settings and will overwrite rules.

~> **NOTE:** Referencing Security Groups across VPC peering has certain restrictions. More information is available in the [VPC Peering User Guide](https://docs.aws.amazon.com/vpc/latest/peering/vpc-peering-security-groups.html).

~> **NOTE:** Due to [AWS Lambda improved VPC networking changes that began deploying in September 2019](https://aws.amazon.com/blogs/compute/announcing-improved-vpc-networking-for-aws-lambda-functions/), security groups associated with Lambda Functions can take up to 45 minutes to successfully delete. Terraform AWS Provider version 2.31.0 and later automatically handles this increased timeout, however prior versions require setting the [customizable deletion timeout](#timeouts) to 45 minutes (`delete = "45m"`). AWS and HashiCorp are working together to reduce the amount of time required for resource deletion and updates can be tracked in this [GitHub issue](https://github.com/hashicorp/terraform-provider-aws/issues/10329).

## Example Usage

Basic usage

```hcl
resource "aws_security_group" "allow_tls" {
  name        = "allow_tls"
  description = "Allow TLS inbound traffic"
  vpc_id      = aws_vpc.main.id

  ingress {
    description = "TLS from VPC"
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = [aws_vpc.main.cidr_block]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = "allow_tls"
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Optional, Forces new resource) The name of the security group. If omitted, Terraform will
assign a random, unique name
* `name_prefix` - (Optional, Forces new resource) Creates a unique name beginning with the specified
  prefix. Conflicts with `name`.
* `description` - (Optional, Forces new resource) The security group description. Defaults to
  "Managed by Terraform". Cannot be "". __NOTE__: This field maps to the AWS
  `GroupDescription` attribute, for which there is no Update API. If you'd like
  to classify your security groups in a way that can be updated, use `tags`.
* `ingress` - (Optional) Can be specified multiple times for each
   ingress rule. Each ingress block supports fields documented below.
   This argument is processed in [attribute-as-blocks mode](https://www.terraform.io/docs/configuration/attr-as-blocks.html).
* `egress` - (Optional, VPC only) Can be specified multiple times for each
      egress rule. Each egress block supports fields documented below.
      This argument is processed in [attribute-as-blocks mode](https://www.terraform.io/docs/configuration/attr-as-blocks.html).
* `revoke_rules_on_delete` - (Optional) Instruct Terraform to revoke all of the
Security Groups attached ingress and egress rules before deleting the rule
itself. This is normally not needed, however certain AWS services such as
Elastic Map Reduce may automatically add required rules to security groups used
with the service, and those rules may contain a cyclic dependency that prevent
the security groups from being destroyed without removing the dependency first.
Default `false`
* `vpc_id` - (Optional, Forces new resource) The VPC ID.
* `tags` - (Optional) A map of tags to assign to the resource.

The `ingress` block supports:

* `cidr_blocks` - (Optional) List of CIDR blocks.
* `ipv6_cidr_blocks` - (Optional) List of IPv6 CIDR blocks.
* `prefix_list_ids` - (Optional) List of Prefix List IDs.
* `from_port` - (Required) The start port (or ICMP type number if protocol is "icmp" or "icmpv6")
* `protocol` - (Required) The protocol. If you select a protocol of "-1" (semantically equivalent to `"all"`, which is not a valid value here), you must specify a "from_port" and "to_port" equal to 0.  The supported values are defined in the "IpProtocol" argument on the [IpPermission](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_IpPermission.html) API reference. This argument is normalized to a lowercase value to match the AWS API requirement when using with Terraform 0.12.x and above, please make sure that the value of the protocol is specified as lowercase when using with older version of Terraform to avoid an issue during upgrade.
* `security_groups` - (Optional) List of security group Group Names if using
    EC2-Classic, or Group IDs if using a VPC.
* `self` - (Optional) If true, the security group itself will be added as
     a source to this ingress rule.
* `to_port` - (Required) The end range port (or ICMP code if protocol is "icmp").
* `description` - (Optional) Description of this ingress rule.

The `egress` block supports:

* `cidr_blocks` - (Optional) List of CIDR blocks.
* `ipv6_cidr_blocks` - (Optional) List of IPv6 CIDR blocks.
* `prefix_list_ids` - (Optional) List of Prefix List IDs.
* `from_port` - (Required) The start port (or ICMP type number if protocol is "icmp")
* `protocol` - (Required) The protocol. If you select a protocol of
"-1" (semantically equivalent to `"all"`, which is not a valid value here), you must specify a "from_port" and "to_port" equal to 0.  The supported values are defined in the "IpProtocol" argument in the [IpPermission](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_IpPermission.html) API reference. This argument is normalized to a lowercase value to match the AWS API requirement when using Terraform 0.12.x and above. Please make sure that the value of the protocol is specified as lowercase when used with older version of Terraform to avoid issues during upgrade.
* `security_groups` - (Optional) List of security group Group Names if using
    EC2-Classic, or Group IDs if using a VPC.
* `self` - (Optional) If true, the security group itself will be added as
     a source to this egress rule.
* `to_port` - (Required) The end range port (or ICMP code if protocol is "icmp").
* `description` - (Optional) Description of this egress rule.

~> **NOTE on Egress rules:** By default, AWS creates an `ALLOW ALL` egress rule when creating a
new Security Group inside of a VPC. When creating a new Security
Group inside a VPC, **Terraform will remove this default rule**, and require you
specifically re-create it if you desire that rule. We feel this leads to fewer
surprises in terms of controlling your egress rules. If you desire this rule to
be in place, you can use this `egress` block:

```hcl
resource "aws_security_group" "example" {
  # ... other configuration ...

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}
```

## Usage with prefix list IDs

Prefix Lists are either managed by AWS internally, or created by the customer using a
[Prefix List resource](ec2_managed_prefix_list.html). Prefix Lists provided by
AWS are associated with a prefix list name, or service name, that is linked to a specific region.
Prefix list IDs are exported on VPC Endpoints, so you can use this format:

```hcl
resource "aws_security_group" "example" {
  # ... other configuration ...

  egress {
    from_port       = 0
    to_port         = 0
    protocol        = "-1"
    prefix_list_ids = [aws_vpc_endpoint.my_endpoint.prefix_list_id]
  }
}

resource "aws_vpc_endpoint" "my_endpoint" {
  # ... other configuration ...
}
```

You can also find a specific Prefix List using the `aws_prefix_list` data source.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the security group
* `arn` - The ARN of the security group
* `vpc_id` - The VPC ID.
* `owner_id` - The owner ID.
* `name` - The name of the security group
* `description` - The description of the security group
* `ingress` - The ingress rules. See above for more.
* `egress` - The egress rules. See above for more.

## Timeouts

`aws_security_group` provides the following [Timeouts](https://www.terraform.io/docs/configuration/blocks/resources/syntax.html#operation-timeouts)
configuration options:

- `create` - (Default `10m`) How long to wait for a security group to be created.
- `delete` - (Default `10m`) How long to retry on `DependencyViolation` errors during security group deletion from lingering ENIs left by certain AWS services such as Elastic Load Balancing. NOTE: Lambda ENIs can take up to 45 minutes to delete, which is not affected by changing this customizable timeout (in version 2.31.0 and later of the Terraform AWS Provider) unless it is increased above 45 minutes.

## Import

Security Groups can be imported using the `security group id`, e.g.

```
$ terraform import aws_security_group.elb_sg sg-903004f8
```
