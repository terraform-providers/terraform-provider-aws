---
subcategory: "API Gateway (REST APIs)"
layout: "aws"
page_title: "AWS: aws_api_gateway_vpc_link"
description: |-
  Get information on a API Gateway VPC Link
---

# Data Source: aws_api_gateway_vpc_link

Use this data source to get the id of a VPC Link in
API Gateway. To fetch the VPC Link you must provide a name to match against.
As there is no unique name constraint on API Gateway VPC Links this data source will
error if there is more than one match.

## Example Usage

```hcl
data "aws_api_gateway_vpc_link" "my_api_gateway_vpc_link" {
  name = "my-vpc-link"
}
```

## Argument Reference

* `name` - (Required) The name of the API Gateway VPC Link to look up. If no API Gateway VPC Link is found with this name, an error will be returned.
 If multiple API Gateway VPC Links are found with this name, an error will be returned.

## Attributes Reference

* `id` - Set to the ID of the found API Gateway VPC Link.
* `description` - The description of the VPC link.
* `status` - The status of the VPC link.
* `status_message` - The status message of the VPC link.
* `target_arns` - The list of network load balancer arns in the VPC targeted by the VPC link. Currently AWS only supports 1 target.
* `tags` - Key-value map of resource tags
