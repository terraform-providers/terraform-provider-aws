---
subcategory: "API Gateway (REST APIs)"
layout: "aws"
page_title: "AWS: aws_api_gateway_vpc_link"
description: |-
  Provides an API Gateway VPC Link.
---

# Resource: aws_api_gateway_vpc_link

Provides an API Gateway VPC Link.

-> **Note:** Amazon API Gateway Version 1 VPC Links enable private integrations that connect REST APIs to private resources in a VPC.
To enable private integration for HTTP APIs, use the Amazon API Gateway Version 2 VPC Link [resource](/docs/providers/aws/r/apigatewayv2_vpc_link.html).

## Example Usage

```hcl
resource "aws_lb" "example" {
  name               = "example"
  internal           = true
  load_balancer_type = "network"

  subnet_mapping {
    subnet_id = "12345"
  }
}

resource "aws_api_gateway_vpc_link" "example" {
  name        = "example"
  description = "example description"
  target_arns = [aws_lb.example.arn]
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name used to label and identify the VPC link.
* `description` - (Optional) The description of the VPC link.
* `target_arns` - (Required, ForceNew) The list of network load balancer arns in the VPC targeted by the VPC link. Currently AWS only supports 1 target.
* `tags` - (Optional) Key-value map of resource tags

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The identifier of the VpcLink.

## Import

API Gateway VPC Link can be imported using the `id`, e.g.

```
$ terraform import aws_api_gateway_vpc_link.example <vpc_link_id>
```
