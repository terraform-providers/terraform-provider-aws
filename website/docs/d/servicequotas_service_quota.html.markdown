---
subcategory: "Service Quotas"
layout: "aws"
page_title: "AWS: aws_servicequotas_service_quota"
description: |-
  Retrieve information about a Service Quota
---

# Data Source: aws_servicequotas_service_quota

Retrieve information about a Service Quota.

## Example Usage

```hcl
data "aws_servicequotas_service_quota" "by_quota_code" {
  quota_code   = "L-F678F1CE"
  service_code = "vpc"
}

data "aws_servicequotas_service_quota" "by_quota_name" {
  quota_name   = "VPCs per Region"
  service_code = "vpc"
}
```

## Argument Reference

~> *NOTE:* Either `quota_code` or `quota_name` must be configured.

* `service_code` - (Required) Service code for the quota. Available values can be found with the [`aws_servicequotas_service` data source](/docs/providers/aws/d/servicequotas_service.html) or [AWS CLI service-quotas list-services command](https://docs.aws.amazon.com/cli/latest/reference/service-quotas/list-services.html).
* `quota_code` - (Optional) Quota code within the service. When configured, the data source directly looks up the service quota. Available values can be found with the [AWS CLI service-quotas list-service-quotas command](https://docs.aws.amazon.com/cli/latest/reference/service-quotas/list-service-quotas.html).
* `quota_name` - (Optional) Quota name within the service. When configured, the data source searches through all service quotas to find the matching quota name. Available values can be found with the [AWS CLI service-quotas list-service-quotas command](https://docs.aws.amazon.com/cli/latest/reference/service-quotas/list-service-quotas.html).

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `adjustable` - Whether the service quota is adjustable.
* `arn` - Amazon Resource Name (ARN) of the service quota.
* `default_value` - Default value of the service quota.
* `global_quota` - Whether the service quota is global for the AWS account.
* `id` - Amazon Resource Name (ARN) of the service quota.
* `service_name` - Name of the service.
* `value` - Current value of the service quota.
