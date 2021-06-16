---
subcategory: "Service Quotas"
layout: "aws"
page_title: "AWS: aws_servicequotas_service"
description: |-
  Retrieve information about a Service Quotas Service
---

# Data Source: aws_servicequotas_service

Retrieve information about a Service Quotas Service.

~> **NOTE:** Global quotas apply to all AWS regions, but can only be accessed in `us-east-1` in the Commercial partition or `us-gov-west-1` in the GovCloud partition. In other regions, the AWS API will return the error `The request failed because the specified service does not exist.`

## Example Usage

```terraform
data "aws_servicequotas_service" "example" {
  service_name = "Amazon Virtual Private Cloud (Amazon VPC)"
}
```

## Argument Reference

* `service_name` - (Required) Service name to lookup within Service Quotas. Available values can be found with the [AWS CLI service-quotas list-services command](https://docs.aws.amazon.com/cli/latest/reference/service-quotas/list-services.html).

## Attributes Reference

* `id` - Code of the service.
* `service_code` - Code of the service.
