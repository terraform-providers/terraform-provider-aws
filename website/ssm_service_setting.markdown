---
subcategory: "SSM"
layout: "aws"
page_title: "AWS: aws_ssm_service_setting"
description: |-
  Defines how a user interacts with or uses a service or a feature of a SSM service.
---

# Resource: aws_ssm_service_setting

Defines how a user interacts with or uses a service or a feature of a SSM service.

## Example Usage

```hcl
resource "aws_ssm_service_setting" "test_setting" {
  service_id    = "arn:aws:ssm:us-east-1:123456789012:servicesetting/ssm/parameter-store/high-throughput-enabled"
  service_value = "true"
}
```

## Argument Reference

The following arguments are supported:

* `service_id` - (Required) The ID of the service setting.
* `service_value` - (Required) The value of the service setting.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The ARN of the service setting.
* `last_modified_date` - The last time the service setting was modified.
* `last_modified_user` - The ARN of the last modified user. This field is populated only if the setting value was overwritten.
* `status` - The status of the service setting. The value can be Default, Customized or PendingUpdate.

## Import

AWS SSM Service Setting can be imported using the `setting_id`, e.g.

```sh
$ terraform import aws_ssm_service_setting.example arn:aws:ssm:us-east-1:123456789012:servicesetting/ssm/parameter-store/high-throughput-enabled
```
