---
subcategory: "VPC"
layout: "aws"
page_title: "AWS: aws_default_subnet"
description: |-
  Manage a default VPC subnet resource.
---

# Resource: aws_default_subnet

Provides a resource to manage a [default AWS VPC subnet](http://docs.aws.amazon.com/AmazonVPC/latest/UserGuide/default-vpc.html#default-vpc-basics) in the current region.

The `aws_default_subnet` behaves differently from normal resources, in that Terraform does not _create_ this resource but instead "adopts" it into management.

The `aws_default_subnet` resource allows you to manage a region's default VPC subnet but Terraform cannot destroy it. Removing this resource from your configuration will remove it from your statefile and Terraform management.

## Example Usage

```hcl
resource "aws_default_subnet" "default_az1" {
  availability_zone = "us-west-2a"

  tags = {
    Name = "Default subnet for us-west-2a"
  }
}
```

## Argument Reference

The following argument is required:

* `availability_zone`- (Required) AZ for the subnet.

The following arguments are optional:

* `map_public_ip_on_launch` - (Optional) Whether instances launched into the subnet should be assigned a public IP address.
* `tags` - (Optional) Map of tags to assign to the resource.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - ARN for the subnet.
* `assign_ipv6_address_on_creation` - Whether IPv6 addresses are assigned on creation.
* `availability_zone_id`- AZ ID of the subnet.
* `cidr_block` - CIDR block for the subnet.
* `id` - ID of the subnet
* `ipv6_association_id` - Association ID for the IPv6 CIDR block.
* `ipv6_cidr_block` - IPv6 CIDR block.
* `owner_id` - ID of the AWS account that owns the subnet.
* `vpc_id` - VPC ID.

## Import

Subnets can be imported using the `subnet id`, e.g.

```
$ terraform import aws_default_subnet.public_subnet subnet-9d4a7b6c
```