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

### Basic

```hcl
resource "aws_vpc_endpoint_service" "example" {
  acceptance_required        = false
  network_load_balancer_arns = ["${aws_lb.example.arn}"]
}
```

### Basic w/ Tags

```hcl
resource "aws_vpc_endpoint_service" "example" {
  acceptance_required        = false
  network_load_balancer_arns = ["${aws_lb.example.arn}"]

  tags = {
    Environment = "test"
  }
}
```

## Argument Reference

The following arguments are supported:

* `acceptance_required` - (Required) Whether or not VPC endpoint connection requests to the service must be accepted by the service owner - `true` or `false`.
* `network_load_balancer_arns` - (Required) The ARNs of one or more Network Load Balancers for the endpoint service.
* `allowed_principals` - (Optional) The ARNs of one or more principals allowed to discover the endpoint service.
* `tags` - (Optional) A map of tags to assign to the resource.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the VPC endpoint service.
* `availability_zones` - The Availability Zones in which the service is available.
* `base_endpoint_dns_names` - The DNS names for the service.
* `manages_vpc_endpoints` - Whether or not the service manages its VPC endpoints - `true` or `false`.
* `private_dns_name` - The private DNS name for the service.
* `service_name` - The service name.
* `service_type` - The service type, `Gateway` or `Interface`.
* `state` - The state of the VPC endpoint service.

## Import

VPC Endpoint Services can be imported using the `VPC endpoint service id`, e.g.

```
$ terraform import aws_vpc_endpoint_service.foo vpce-svc-0f97a19d3fa8220bc
```
