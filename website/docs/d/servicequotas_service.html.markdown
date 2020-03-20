---
subcategory: "Service Quotas"
layout: "aws"
page_title: "AWS: aws_servicequotas_service"
description: |-
  Retrieve information about a Service Quotas Service
---

# Data Source: aws_servicequotas_service

Retrieve information about a Service Quotas Service.

## Example Usage

```hcl
data "aws_servicequotas_service" "example" {
  service_name = "Amazon Virtual Private Cloud (Amazon VPC)"
}
```

## Argument Reference

* `service_name` - (Required) Service name to lookup within Service Quotas. Available values can be found with the [AWS CLI service-quotas list-services command](https://docs.aws.amazon.com/cli/latest/reference/service-quotas/list-services.html).

## Attributes Reference

* `id` - Code of the service.
* `service_code` - Code of the service.
