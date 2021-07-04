---
subcategory: "VPC"
layout: "aws"
page_title: "AWS: aws_vpc_endpoint_policy"
description: |-
  Provides a VPC Endpoint Policy resource.
---

# Resource: aws_vpc_endpoint_policy

Provides a VPC Endpoint Policy resource.

## Example Usage

```terraform
data "aws_vpc_endpoint_service" "example" {
  service = "dynamodb"
}

resource "aws_vpc" "example" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_vpc_endpoint" "example" {
  service_name = data.aws_vpc_endpoint_service.example.service_name
  vpc_id       = aws_vpc.example.id
}

resource "aws_vpc_endpoint_policy" "example" {
  vpc_endpoint_id = aws_vpc_endpoint.example.id
  policy = jsonencode({
    "Version" : "2012-10-17",
    "Statement" : [
      {
        "Sid" : "AllowAll",
        "Effect" : "Allow",
        "Principal" : {
          "AWS" : "*"
        },
        "Action" : [
          "dynamodb:*"
        ],
        "Resource" : "*"
      }
    ]
  })
}
```

## Argument Reference

The following arguments are supported:

* `vpc_endpoint_id` - (Required) The VPC Endpoint ID.
* `policy` - (Optional) A policy to attach to the endpoint that controls access to the service. Defaults to full access. All `Gateway` and some `Interface` endpoints support policies - see the [relevant AWS documentation](https://docs.aws.amazon.com/vpc/latest/userguide/vpc-endpoints-access.html) for more details. For more information about building AWS IAM policy documents with Terraform, see the [AWS IAM Policy Document Guide](https://learn.hashicorp.com/terraform/aws/iam-policy).

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the VPC endpoint.

## Import

VPC Endpoint Policies can be imported using the `vpc endpoint id`, e.g.

```
$ terraform import aws_vpc_endpoint_policy.example vpce-3ecf2a57
```
