---
subcategory: "EC2"
layout: "aws"
page_title: "AWS: aws_ec2_tag"
description: |-
  Manages an individual EC2 resource tag
---

# Resource: aws_ec2_tag

Manages an individual EC2 resource tag. This resource should only be used in cases where EC2 resources are created outside Terraform (e.g. AMIs), being shared via Resource Access Manager (RAM), or implicitly created by other means (e.g. Transit Gateway VPN Attachments).

~> **NOTE:** This tagging resource should not be combined with the Terraform resource for managing the parent resource. For example, using `aws_vpc` and `aws_ec2_tag` to manage tags of the same VPC will cause a perpetual difference where the `aws_vpc` resource will try to remove the tag being added by the `aws_ec2_tag` resource.

~> **NOTE:** This tagging resource does not use the [provider `ignore_tags` configuration](/docs/providers/aws/index.html#ignore_tags).

## Example Usage

```hcl
resource "aws_ec2_transit_gateway" "example" {}

resource "aws_customer_gateway" "example" {
  bgp_asn    = 65000
  ip_address = "172.0.0.1"
  type       = "ipsec.1"
}

resource "aws_vpn_connection" "example" {
  customer_gateway_id = aws_customer_gateway.example.id
  transit_gateway_id  = aws_ec2_transit_gateway.example.id
  type                = aws_customer_gateway.example.type
}

resource "aws_ec2_tag" "example" {
  resource_id = aws_vpn_connection.example.transit_gateway_attachment_id
  key         = "Name"
  value       = "Hello World"
}
```

## Argument Reference

The following arguments are supported:

* `resource_id` - (Required) The ID of the EC2 resource to manage the tag for.
* `key` - (Required) The tag name.
* `value` - (Required) The value of the tag.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - EC2 resource identifier and key, separated by a comma (`,`)

## Import

`aws_ec2_tag` can be imported by using the EC2 resource identifier and key, separated by a comma (`,`), e.g.

```
$ terraform import aws_ec2_tag.example tgw-attach-1234567890abcdef,Name
```
