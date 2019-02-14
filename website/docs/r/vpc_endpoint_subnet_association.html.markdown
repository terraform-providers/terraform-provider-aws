---
layout: "aws"
page_title: "AWS: aws_vpc_endpoint_subnet_association"
sidebar_current: "docs-aws-resource-vpc-endpoint-subnet-association"
description: |-
  Provides a resource to create an association between a VPC endpoint and a subnet.
---

# aws_vpc_endpoint_subnet_association

Provides a resource to create an association between a VPC endpoint and a subnet.

~> **NOTE on VPC Endpoints and VPC Endpoint Subnet Associations:** Terraform provides
both a standalone VPC Endpoint Subnet Association (an association between a VPC endpoint
and a single `subnet_id`) and a [VPC Endpoint](vpc_endpoint.html) resource with a `subnet_ids`
attribute. Do not use the same subnet ID in both a VPC Endpoint resource and a VPC Endpoint Subnet
Association resource. Doing so will cause a conflict of associations and will overwrite the association.

## Example Usage

Basic usage:

```hcl
resource "aws_vpc_endpoint_subnet_association" "sn_ec2" {
  vpc_endpoint_id = "${aws_vpc_endpoint.ec2.id}"
  subnet_id       = "${aws_subnet.sn.id}"
}
```

## Argument Reference

The following arguments are supported:

* `vpc_endpoint_id` - (Required) The ID of the VPC endpoint with which the subnet will be associated.
* `subnet_id` - (Required) The ID of the subnet to be associated with the VPC endpoint.

### Timeouts

`aws_vpc_endpoint_subnet_association` provides the following
[Timeouts](/docs/configuration/resources.html#timeouts) configuration options:

- `create` - (Default `10 minutes`) Used for creating the association
- `delete` - (Default `10 minutes`) Used for destroying the association

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the association.
