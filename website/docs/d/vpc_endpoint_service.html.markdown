---
layout: "aws"
page_title: "AWS: aws_vpc_endpoint_service"
sidebar_current: "docs-aws-datasource-vpc-endpoint-service"
description: |-
    Provides details about a specific service that can be specified when creating a VPC endpoint.
---

# Data Source: aws_vpc_endpoint_service

The VPC Endpoint Service data source details about a specific service that
can be specified when creating a VPC endpoint within the region configured in the provider.

## Example Usage

AWS service usage:

```hcl
# Declare the data source
data "aws_vpc_endpoint_service" "s3" {
  service = "s3"
}

# Create a VPC
resource "aws_vpc" "foo" {
  cidr_block = "10.0.0.0/16"
}

# Create a VPC endpoint
resource "aws_vpc_endpoint" "ep" {
  vpc_id       = "${aws_vpc.foo.id}"
  service_name = "${data.aws_vpc_endpoint_service.s3.service_name}"
}
```

Non-AWS service usage:

```hcl
data "aws_vpc_endpoint_service" "custome" {
  service_name = "com.amazonaws.vpce.us-west-2.vpce-svc-0e87519c997c63cd8"
}
```

## Argument Reference

The arguments of this data source act as filters for querying the available VPC endpoint services.
The given filters must match exactly one VPC endpoint service whose data will be exported as attributes.

* `service` - (Optional) The common name of an AWS service (e.g. `s3`).
* `service_name` - (Optional) The service name that can be specified when creating a VPC endpoint.

~> **NOTE:** One of `service` or `service_name` must be specified.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `service_type` - The service type, `Gateway` or `Interface`.
* `owner` - The AWS account ID of the service owner or `amazon`.
* `vpc_endpoint_policy_supported` - Whether or not the service supports endpoint policies - `true` or `false`.
* `acceptance_required` - Whether or not VPC endpoint connection requests to the service must be accepted by the service owner - `true` or `false`.
* `availability_zones` - The Availability Zones in which the service is available.
* `private_dns_name` - The private DNS name for the service.
* `base_endpoint_dns_names` - The DNS names for the service.
