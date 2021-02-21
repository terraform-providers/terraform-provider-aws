---
subcategory: "VPC"
layout: "aws"
page_title: "AWS: aws_default_route_table"
description: |-
  Provides a resource to manage a default route table of a VPC.
---

# Resource: aws_default_route_table

Provides a resource to manage a default route table of a VPC. This resource can manage the default route table of the default or a non-default VPC.

~> **NOTE:** This is an advanced resource with special caveats. Please read this document in its entirety before using this resource. The `aws_default_route_table` resource behaves differently from normal resources. Terraform does not _create_ this resource but instead attempts to "adopt" it into management. **Do not** use both `aws_default_route_table` to manage a default route table **and** `aws_main_route_table_association` with the same VPC due to possible route conflicts.

Every VPC has a default route table that can be managed but not destroyed. When Terraform first adopts a default route table, it **immediately removes all defined routes**. It then proceeds to create any routes specified in the configuration. This step is required so that only the routes specified in the configuration exist in the default route table.

For more information, see the Amazon VPC User Guide on [Route Tables](https://docs.aws.amazon.com/vpc/latest/userguide/VPC_Route_Tables.html). For information about managing normal route tables in Terraform, see [`aws_route_table`](/docs/providers/aws/r/route_table.html).

## Example Usage

```hcl
resource "aws_default_route_table" "r" {
  default_route_table_id = aws_vpc.foo.default_route_table_id

  route {
    # ...
  }

  tags = {
    Name = "default table"
  }
}
```

## Argument Reference

The following arguments are required:

* `default_route_table_id` - (Required) ID of the default route table.

The following arguments are optional:

* `propagating_vgws` - (Optional) List of virtual gateways for propagation.
* `route` - (Optional) Configuration block of routes. Detailed below.
* `tags` - (Optional) Map of tags to assign to the resource.

### route

This argument is processed in [attribute-as-blocks mode](https://www.terraform.io/docs/configuration/attr-as-blocks.html).

One of the following destination arguments must be supplied:

* `cidr_block` - (Required) The CIDR block of the route.
* `ipv6_cidr_block` - (Optional) The Ipv6 CIDR block of the route

One of the following target arguments must be supplied:

* `egress_only_gateway_id` - (Optional) Identifier of a VPC Egress Only Internet Gateway.
* `gateway_id` - (Optional) Identifier of a VPC internet gateway or a virtual private gateway.
* `instance_id` - (Optional) Identifier of an EC2 instance.
* `nat_gateway_id` - (Optional) Identifier of a VPC NAT gateway.
* `network_interface_id` - (Optional) Identifier of an EC2 network interface.
* `transit_gateway_id` - (Optional) Identifier of an EC2 Transit Gateway.
* `vpc_endpoint_id` - (Optional) Identifier of a VPC Endpoint. This route must be removed prior to VPC Endpoint deletion.
* `vpc_peering_connection_id` - (Optional) Identifier of a VPC peering connection.

Note that the default route, mapping the VPC's CIDR block to "local", is created implicitly and cannot be specified.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - ID of the route table.
* `owner_id` - ID of the AWS account that owns the route table.
* `vpc_id` - ID of the VPC.

## Import

Default VPC route tables can be imported using the `vpc_id`, e.g.

```
$ terraform import aws_default_route_table.example vpc-33cc44dd
```

[aws-route-tables]: http://docs.aws.amazon.com/AmazonVPC/latest/UserGuide/VPC_Route_Tables.html#Route_Replacing_Main_Table
[tf-route-tables]: /docs/providers/aws/r/route_table.html
