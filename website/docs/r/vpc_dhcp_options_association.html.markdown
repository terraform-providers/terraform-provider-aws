---
subcategory: "VPC"
layout: "aws"
page_title: "AWS: aws_vpc_dhcp_options_association"
description: |-
  Provides a VPC DHCP Options Association resource.
---

# Resource: aws_vpc_dhcp_options_association

Provides a VPC DHCP Options Association resource.

## Example Usage

```hcl
resource "aws_vpc_dhcp_options_association" "dns_resolver" {
  vpc_id          = aws_vpc.foo.id
  dhcp_options_id = aws_vpc_dhcp_options.foo.id
}
```

## Argument Reference

The following arguments are supported:

* `vpc_id` - (Required) The ID of the VPC to which we would like to associate a DHCP Options Set.
* `dhcp_options_id` - (Required) The ID of the DHCP Options Set to associate to the VPC.

## Remarks

* You can only associate one DHCP Options Set to a given VPC ID.
* Removing the DHCP Options Association automatically sets AWS's `default` DHCP Options Set to the VPC.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the DHCP Options Set Association.

## Import

DHCP associations can be imported by providing the VPC ID associated with the options:

```
$ terraform import aws_vpc_dhcp_options_association.imported vpc-0f001273ec18911b1
```
