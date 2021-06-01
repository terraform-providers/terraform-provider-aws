---
subcategory: "EC2"
layout: "aws"
page_title: "AWS: aws_ec2_carrier_gateway"
description: |-
  Manages an EC2 Carrier Gateway.
---

# Resource: aws_ec2_carrier_gateway

Manages an EC2 Carrier Gateway. See the AWS [documentation](https://docs.aws.amazon.com/vpc/latest/userguide/Carrier_Gateway.html) for more information.

## Example Usage

```hcl
resource "aws_ec2_carrier_gateway" "example" {
  vpc_id = aws_vpc.example.id

  tags = {
    Name = "example-carrier-gateway"
  }
}
```

## Argument Reference

The following arguments are supported:

* `vpc_id` - (Required) The ID of the VPC to associate with the carrier gateway.
* `tags` - (Optional) A map of tags to assign to the resource.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the carrier gateway.
* `arn` - The ARN of the carrier gateway.
* `owner_id` - The AWS account ID of the owner of the carrier gateway.

## Import

`aws_ec2_carrier_gateway` can be imported using the carrier gateway's ID,
e.g.

```
$ terraform import aws_ec2_carrier_gateway.example cgw-12345
```
