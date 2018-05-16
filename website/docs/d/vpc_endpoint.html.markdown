---
layout: "aws"
page_title: "AWS: aws_vpc_endpoint"
sidebar_current: "docs-aws-datasource-vpc-endpoint-x"
description: |-
    Provides details about a specific VPC endpoint.
---

# Data Source: aws_vpc_endpoint

The VPC Endpoint data source provides details about
a specific VPC endpoint.

## Example Usage

```hcl
# Declare the data source
data "aws_vpc_endpoint" "s3" {
  vpc_id       = "${aws_vpc.foo.id}"
  service_name = "com.amazonaws.us-west-2.s3"
}

resource "aws_vpc_endpoint_route_table_association" "private_s3" {
  vpc_endpoint_id = "${data.aws_vpc_endpoint.s3.id}"
  route_table_id  = "${aws_route_table.private.id}"
}
```

## Argument Reference

The arguments of this data source act as filters for querying the available VPC endpoints.
The given filters must match exactly one VPC endpoint whose data will be exported as attributes.

* `id` - (Optional) The ID of the specific VPC Endpoint to retrieve.

* `state` - (Optional) The state of the specific VPC Endpoint to retrieve.

* `vpc_id` - (Optional) The ID of the VPC in which the specific VPC Endpoint is used.

* `service_name` - (Optional) The AWS service name of the specific VPC Endpoint to retrieve.

## Attributes Reference

All of the argument attributes are also exported as result attributes.

* `vpc_endpoint_type` - The VPC Endpoint type, `Gateway` or `Interface`.
* `policy` - The policy document associated with the VPC Endpoint. Applicable for endpoints of type `Gateway`.
* `route_table_ids` - One or more route tables associated with the VPC Endpoint. Applicable for endpoints of type `Gateway`.
* `prefix_list_id` - The prefix list ID of the exposed AWS service. Applicable for endpoints of type `Gateway`.
* `cidr_blocks` - The list of CIDR blocks for the exposed AWS service. Applicable for endpoints of type `Gateway`.
* `subnet_ids` - One or more subnets in which the VPC Endpoint is located. Applicable for endpoints of type `Interface`.
* `network_interface_ids` - One or more network interfaces for the VPC Endpoint. Applicable for endpoints of type `Interface`.
* `security_group_ids` - One or more security groups associated with the network interfaces. Applicable for endpoints of type `Interface`.
* `private_dns_enabled` - Whether or not the VPC is associated with a private hosted zone - `true` or `false`. Applicable for endpoints of type `Interface`.
* `dns_entry` - The DNS entries for the VPC Endpoint. Applicable for endpoints of type `Interface`. DNS blocks are documented below.

DNS blocks (for `dns_entry`) support the following attributes:

* `dns_name` - The DNS name.
* `hosted_zone_id` - The ID of the private hosted zone.
