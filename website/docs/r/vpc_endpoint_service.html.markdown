---
subcategory: "VPC"
layout: "aws"
page_title: "AWS: aws_vpc_endpoint_service"
description: |-
  Provides a VPC Endpoint Service resource.
---

# Resource: aws_vpc_endpoint_service

Provides a VPC Endpoint Service resource.
Service consumers can create an _Interface_ [VPC Endpoint](vpc_endpoint.html) to connect to the service.

~> **NOTE on VPC Endpoint Services and VPC Endpoint Service Allowed Principals:** Terraform provides
both a standalone [VPC Endpoint Service Allowed Principal](vpc_endpoint_service_allowed_principal.html) resource
and a VPC Endpoint Service resource with an `allowed_principals` attribute. Do not use the same principal ARN in both
a VPC Endpoint Service resource and a VPC Endpoint Service Allowed Principal resource. Doing so will cause a conflict
and will overwrite the association.

## Example Usage

### Network Load Balancers

```hcl
resource "aws_vpc_endpoint_service" "example" {
  acceptance_required        = false
  network_load_balancer_arns = [aws_lb.example.arn]
}
```

### Gateway Load Balancers

```hcl
resource "aws_vpc_endpoint_service" "example" {
  acceptance_required        = false
  gateway_load_balancer_arns = [aws_lb.example.arn]
}
```

## Argument Reference

The following arguments are supported:

* `acceptance_required` - (Required) Whether or not VPC endpoint connection requests to the service must be accepted by the service owner - `true` or `false`.
* `allowed_principals` - (Optional) The ARNs of one or more principals allowed to discover the endpoint service.
* `gateway_load_balancer_arns` - (Optional) Amazon Resource Names (ARNs) of one or more Gateway Load Balancers for the endpoint service.
* `network_load_balancer_arns` - (Optional) Amazon Resource Names (ARNs) of one or more Network Load Balancers for the endpoint service.
* `tags` - (Optional) A map of tags to assign to the resource.
* `private_dns_name` - (Optional) The private DNS name for the service.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the VPC endpoint service.
* `availability_zones` - The Availability Zones in which the service is available.
* `arn` - The Amazon Resource Name (ARN) of the VPC endpoint service.
* `base_endpoint_dns_names` - The DNS names for the service.
* `manages_vpc_endpoints` - Whether or not the service manages its VPC endpoints - `true` or `false`.
* `service_name` - The service name.
* `service_type` - The service type, `Gateway` or `Interface`.
* `state` - The state of the VPC endpoint service.
* `private_dns_name_configuration` - List of objects containing information about the endpoint service private DNS name configuration.
    * `name` - Name of the record subdomain the service provider needs to create.
    * `state` - Verification state of the VPC endpoint service. Consumers of the endpoint service can use the private name only when the state is `verified`.
    * `type` - Endpoint service verification type, for example `TXT`.
    * `value` - Value the service provider adds to the private DNS name domain record before verification.

## Import

VPC Endpoint Services can be imported using the `VPC endpoint service id`, e.g.

```
$ terraform import aws_vpc_endpoint_service.foo vpce-svc-0f97a19d3fa8220bc
```
