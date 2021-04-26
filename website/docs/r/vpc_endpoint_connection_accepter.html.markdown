---
subcategory: "VPC"
layout: "aws"
page_title: "AWS: aws_vpc_endpoint_connection_accepter"
description: |-
  Provides a resource to accept a pending VPC Endpoint accept request to VPC Endpoint Service.
---

# Resource: aws_vpc_endpoint_connection_accepter

Provides a resource to accept a pending VPC Endpoint Connection accept request to VPC Endpoint Service.

## Example Usage

### Accept cross-account request

```terraform
resource "aws_vpc_endpoint_service" "example" {
  acceptance_required        = false
  network_load_balancer_arns = [aws_lb.example.arn]
}

resource "aws_vpc_endpoint" "example" {
  provider = "aws.alternate"

  vpc_id              = aws_vpc.test_alternate.id
  service_name        = aws_vpc_endpoint_service.test.service_name
  vpc_endpoint_type   = "Interface"
  private_dns_enabled = false

  security_group_ids = [
    aws_security_group.test.id,
  ]
}

resource "aws_vpc_endpoint_connection_accepter" "example" {
  service_id  = aws_vpc_endpoint_service.example.id
  endpoint_id = aws_vpc_endpoint.example.id
}
```

## Argument Reference

The following arguments are supported:

* `service_id` - (Required) AWS VPC Endpoint Service ID.
* `endpoint_id` - (Required) AWS VPC Endpoint ID.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the VPC Endpoint Connection.
* `state` - State of the VPC Endpoint Connection.

## Import

VPC Endpoint Services can be imported using ID of the connection, which is the `VPC Endpoint Service ID` and `VPC Endpoint ID` separated by underscore (`_`). e.g.

```
$ terraform import aws_vpc_endpoint_connection_accepter.foo vpce-svc-0f97a19d3fa8220bc_vpce-010601a6db371e263
```
