---
subcategory: "Service Quotas"
layout: "aws"
page_title: "AWS: aws_servicequotas_service_quota"
description: |-
  Manages an individual Service Quota
---

# Resource: aws_servicequotas_service_quota

Manages an individual Service Quota.

## Example Usage

```hcl
resource "aws_servicequotas_service_quota" "example" {
  quota_code   = "L-F678F1CE"
  service_code = "vpc"
  value        = 75
}
```

## Argument Reference

The following arguments are supported:

* `quota_code` - (Required) Code of the service quota to track. For example: `L-F678F1CE`. Available values can be found with the [AWS CLI service-quotas list-service-quotas command](https://docs.aws.amazon.com/cli/latest/reference/service-quotas/list-service-quotas.html).
* `service_code` - (Required) Code of the service to track. For example: `vpc`. Available values can be found with the [AWS CLI service-quotas list-services command](https://docs.aws.amazon.com/cli/latest/reference/service-quotas/list-services.html).
* `value` - (Required) Float specifying the desired value for the service quota. If the desired value is higher than the current value, a quota increase request is submitted. When a known request is submitted and pending, the value reflects the desired value of the pending request.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `adjustable` - Whether the service quota can be increased.
* `arn` - Amazon Resource Name (ARN) of the service quota.
* `default_value` - Default value of the service quota.
* `id` - Service code and quota code, separated by a front slash (`/`)
* `quota_name` - Name of the quota.
* `service_name` - Name of the service.

## Import

~> *NOTE* This resource does not require explicit import and will assume management of an existing service quota on Terraform resource creation.

`aws_servicequotas_service_quota` can be imported by using the service code and quota code, separated by a front slash (`/`), e.g.

```
$ terraform import aws_servicequotas_service_quota.example vpc/L-F678F1CE
```
