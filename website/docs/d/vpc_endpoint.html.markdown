---
subcategory: "VPC"
layout: "aws"
page_title: "AWS: aws_vpc_endpoint"
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
  vpc_id       = aws_vpc.foo.id
  service_name = "com.amazonaws.us-west-2.s3"
}

resource "aws_vpc_endpoint_route_table_association" "private_s3" {
  vpc_endpoint_id = data.aws_vpc_endpoint.s3.id
  route_table_id  = aws_route_table.private.id
}
```

## Argument Reference

The arguments of this data source act as filters for querying the available VPC endpoints.
The given filters must match exactly one VPC endpoint whose data will be exported as attributes.

* `filter` - (Optional) Custom filter block as described below.
* `id` - (Optional) The ID of the specific VPC Endpoint to retrieve.
* `service_name` - (Optional) The service name of the specific VPC Endpoint to retrieve. For AWS services the service name is usually in the form `com.amazonaws.<region>.<service>` (the SageMaker Notebook service is an exception to this rule, the service name is in the form `aws.sagemaker.<region>.notebook`).
* `state` - (Optional) The state of the specific VPC Endpoint to retrieve.
* `tags` - (Optional) A map of tags, each pair of which must exactly match
  a pair on the specific VPC Endpoint to retrieve.
* `vpc_id` - (Optional) The ID of the VPC in which the specific VPC Endpoint is used.

More complex filters can be expressed using one or more `filter` sub-blocks,
which take the following arguments:

* `name` - (Required) The name of the field to filter by, as defined by
  [the underlying AWS API](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DescribeVpcEndpoints.html).
* `values` - (Required) Set of values that are accepted for the given field.
  A VPC Endpoint will be selected if any one of the given values matches.

## Attributes Reference

In addition to all arguments above except `filter`, the following attributes are exported:

* `arn` - The Amazon Resource Name (ARN) of the VPC endpoint.
* `cidr_blocks` - The list of CIDR blocks for the exposed AWS service. Applicable for endpoints of type `Gateway`.
* `dns_entry` - The DNS entries for the VPC Endpoint. Applicable for endpoints of type `Interface`. DNS blocks are documented below.
* `network_interface_ids` - One or more network interfaces for the VPC Endpoint. Applicable for endpoints of type `Interface`.
* `owner_id` - The ID of the AWS account that owns the VPC endpoint.
* `policy` - The policy document associated with the VPC Endpoint. Applicable for endpoints of type `Gateway`.
* `prefix_list_id` - The prefix list ID of the exposed AWS service. Applicable for endpoints of type `Gateway`.
* `private_dns_enabled` - Whether or not the VPC is associated with a private hosted zone - `true` or `false`. Applicable for endpoints of type `Interface`.
* `requester_managed` -  Whether or not the VPC Endpoint is being managed by its service - `true` or `false`.
* `route_table_ids` - One or more route tables associated with the VPC Endpoint. Applicable for endpoints of type `Gateway`.
* `security_group_ids` - One or more security groups associated with the network interfaces. Applicable for endpoints of type `Interface`.
* `subnet_ids` - One or more subnets in which the VPC Endpoint is located. Applicable for endpoints of type `Interface`.
* `vpc_endpoint_type` - The VPC Endpoint type, `Gateway` or `Interface`.

DNS blocks (for `dns_entry`) support the following attributes:

* `dns_name` - The DNS name.
* `hosted_zone_id` - The ID of the private hosted zone.
