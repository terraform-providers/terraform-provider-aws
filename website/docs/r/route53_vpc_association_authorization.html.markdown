---
subcategory: "Route53"
layout: "aws"
page_title: "AWS: aws_route53_vpc_association_authorization"
description: |-
  Authorizes a VPC in a peer account to be associated with a local Route53 Hosted Zone
---

# Resource: aws_route53_vpc_association_authorization

Authorizes a VPC in a peer account to be associated with a local Route53 Hosted Zone.

## Example Usage

```hcl
provider "aws" {
}

provider "aws" {
  alias = "alternate"
}

resource "aws_vpc" "example" {
  cidr_block           = "10.6.0.0/16"
  enable_dns_hostnames = true
  enable_dns_support   = true
}

resource "aws_route53_zone" "example" {
  name = "example.com"

  vpc {
    vpc_id = aws_vpc.example.id
  }
}

resource "aws_vpc" "alternate" {
  provider = "aws.alternate"

  cidr_block           = "10.7.0.0/16"
  enable_dns_hostnames = true
  enable_dns_support   = true
}

resource "aws_route53_vpc_association_authorization" "example" {
  vpc_id  = aws_vpc.alternate.id
  zone_id = aws_route53_zone.example.id
}

resource "aws_route53_zone_association" "example" {
  provider = "aws.alternate"

  vpc_id  = aws_route53_vpc_association_authorization.example.vpc_id
  zone_id = aws_route53_vpc_association_authorization.example.zone_id
}
```

## Argument Reference

The following arguments are supported:

* `zone_id` - (Required) The ID of the private hosted zone that you want to authorize associating a VPC with.
* `vpc_id` - (Required) The VPC to authorize for association with the private hosted zone.
* `vpc_region` - (Optional) The VPC's region. Defaults to the region of the AWS provider.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The calculated unique identifier for the association.

## Import

Route 53 VPC Association Authorizations can be imported via the Hosted Zone ID and VPC ID, separated by a colon (`:`), e.g.

```
$ terraform import aws_route53_vpc_association_authorization.example Z123456ABCDEFG:vpc-12345678
```
