---
subcategory: "API Gateway v2 (WebSocket and HTTP APIs)"
layout: "aws"
page_title: "AWS: aws_apigatewayv2_vpc_link"
description: |-
  Manages an Amazon API Gateway Version 2 VPC Link.
---

# Resource: aws_apigatewayv2_vpc_link

Manages an Amazon API Gateway Version 2 VPC Link.

-> **Note:** Amazon API Gateway Version 2 VPC Links enable private integrations that connect HTTP APIs to private resources in a VPC.
To enable private integration for REST APIs, use the Amazon API Gateway Version 1 VPC Link [resource](/docs/providers/aws/r/api_gateway_vpc_link.html).

## Example Usage

```hcl
resource "aws_apigatewayv2_vpc_link" "example" {
  name               = "example"
  security_group_ids = [data.aws_security_group.example.id]
  subnet_ids         = data.aws_subnet_ids.example.ids

  tags = {
    Usage = "example"
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the VPC Link. Must be between 1 and 128 characters in length.
* `security_group_ids` - (Required) Security group IDs for the VPC Link.
* `subnet_ids` - (Required) Subnet IDs for the VPC Link.
* `tags` - (Optional) A map of tags to assign to the VPC Link.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The VPC Link identifier.
* `arn` - The VPC Link ARN.

## Import

`aws_apigatewayv2_vpc_link` can be imported by using the VPC Link identifier, e.g.

```
$ terraform import aws_apigatewayv2_vpc_link.example aabbccddee
```
