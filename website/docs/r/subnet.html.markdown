---
subcategory: "VPC"
layout: "aws"
page_title: "AWS: aws_subnet"
description: |-
  Provides an VPC subnet resource.
---

# Resource: aws_subnet

Provides an VPC subnet resource.

~> **NOTE:** Due to [AWS Lambda improved VPC networking changes that began deploying in September 2019](https://aws.amazon.com/blogs/compute/announcing-improved-vpc-networking-for-aws-lambda-functions/), subnets associated with Lambda Functions can take up to 45 minutes to successfully delete. Terraform AWS Provider version 2.31.0 and later automatically handles this increased timeout, however prior versions require setting the [customizable deletion timeout](#timeouts) to 45 minutes (`delete = "45m"`). AWS and HashiCorp are working together to reduce the amount of time required for resource deletion and updates can be tracked in this [GitHub issue](https://github.com/hashicorp/terraform-provider-aws/issues/10329).

## Example Usage

### Basic Usage

```hcl
resource "aws_subnet" "main" {
  vpc_id     = aws_vpc.main.id
  cidr_block = "10.0.1.0/24"

  tags = {
    Name = "Main"
  }
}
```

### Subnets In Secondary VPC CIDR Blocks

When managing subnets in one of a VPC's secondary CIDR blocks created using a [`aws_vpc_ipv4_cidr_block_association`](vpc_ipv4_cidr_block_association.html)
resource, it is recommended to reference that resource's `vpc_id` attribute to ensure correct dependency ordering.

```hcl
resource "aws_vpc_ipv4_cidr_block_association" "secondary_cidr" {
  vpc_id     = aws_vpc.main.id
  cidr_block = "172.2.0.0/16"
}

resource "aws_subnet" "in_secondary_cidr" {
  vpc_id     = aws_vpc_ipv4_cidr_block_association.secondary_cidr.vpc_id
  cidr_block = "172.2.0.0/24"
}
```

## Argument Reference

The following arguments are supported:

* `availability_zone` - (Optional) The AZ for the subnet.
* `availability_zone_id` - (Optional) The AZ ID of the subnet.
* `cidr_block` - (Required) The CIDR block for the subnet.
* `customer_owned_ipv4_pool` - (Optional) The customer owned IPv4 address pool. Typically used with the `map_customer_owned_ip_on_launch` argument. The `outpost_arn` argument must be specified when configured.
* `ipv6_cidr_block` - (Optional) The IPv6 network range for the subnet,
    in CIDR notation. The subnet size must use a /64 prefix length.
* `map_customer_owned_ip_on_launch` -  (Optional) Specify `true` to indicate that network interfaces created in the subnet should be assigned a customer owned IP address. The `customer_owned_ipv4_pool` and `outpost_arn` arguments must be specified when set to `true`. Default is `false`.
* `map_public_ip_on_launch` -  (Optional) Specify true to indicate
    that instances launched into the subnet should be assigned
    a public IP address. Default is `false`.
* `outpost_arn` - (Optional) The Amazon Resource Name (ARN) of the Outpost.
* `assign_ipv6_address_on_creation` - (Optional) Specify true to indicate
    that network interfaces created in the specified subnet should be
    assigned an IPv6 address. Default is `false`
* `vpc_id` - (Required) The VPC ID.
* `tags` - (Optional) A map of tags to assign to the resource.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the subnet
* `arn` - The ARN of the subnet.
* `ipv6_cidr_block_association_id` - The association ID for the IPv6 CIDR block.
* `owner_id` - The ID of the AWS account that owns the subnet.

## Timeouts

`aws_subnet` provides the following [Timeouts](https://www.terraform.io/docs/configuration/blocks/resources/syntax.html#operation-timeouts)
configuration options:

- `create` - (Default `10m`) How long to wait for a subnet to be created.
- `delete` - (Default `20m`) How long to retry on `DependencyViolation` errors during subnet deletion from lingering ENIs left by certain AWS services such as Elastic Load Balancing. NOTE: Lambda ENIs can take up to 45 minutes to delete, which is not affected by changing this customizable timeout (in version 2.31.0 and later of the Terraform AWS Provider) unless it is increased above 45 minutes.

## Import

Subnets can be imported using the `subnet id`, e.g.

```
$ terraform import aws_subnet.public_subnet subnet-9d4a7b6c
```
