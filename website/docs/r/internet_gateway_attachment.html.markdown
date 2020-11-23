---
subcategory: "VPC"
layout: "aws"
page_title: "AWS: aws_internet_gateway_attachment"
description: |-
  Provides a resource to create a VPC Internet Gateway Attachment.
---

# Resource: aws_internet_gateway_attachment

Provides a resource to create a VPC Internet Gateway Attachment.

## Example Usage

```hcl
resource "aws_internet_gateway_attachment" "example" {
  vpc_id              = aws_vpc.example.id
  internet_gateway_id = aws_internet_gateway.example.id
}

resource "aws_vpc" "example" {
  cidr_block = "10.1.0.0/16"
}

resource "aws_internet_gateway" "example" {
  lifecycle {
    ignore_changes = ["vpc_id"]
  }
}
```

## Argument Reference

The following arguments are supported:

* `vpc_id` - (Required) The ID of the VPC.
* `internet_gateway_id` - (Required) The ID of the internet gateway.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the VPC and Internet Gateway separated by a colon.


## Import

Internet Gateway Attachments can be imported using the `id`, e.g.

```
$ terraform import aws_internet_gateway_attachment.gw vpc-123456:igw-c0a643a9
```
