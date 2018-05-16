---
layout: "aws"
page_title: "AWS: sns_platform_application"
sidebar_current: "docs-aws-resource-sns-platform-application"
description: |-
  Provides an SNS platform application resource.
---

# aws_sns_platform_application

Provides an SNS platform application resource

## Example Usage

### Apple Push Notification Service (APNS)

```hcl
resource "aws_sns_platform_application" "apns_application" {
  name                = "apns_application"
  platform            = "APNS"
  platform_credential = "<APNS PRIVATE KEY>"
  platform_principal  = "<APNS CERTIFICATE>"
}
```

### Google Cloud Messaging (GCM)

```hcl
resource "aws_sns_platform_application" "gcm_application" {
  name                = "gcm_application"
  platform            = "GCM"
  platform_credential = "<GCM API KEY>"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The friendly name for the SNS platform application
* `platform` - (Required) The platform that the app is registered with. See [Platform][1] for supported platforms.
* `platform_credential` - (Required) Application Platform credential. See [Credential][1] for type of credential required for platform. The value of this attribute when stored into the Terraform state is only a hash of the real value, so therefore it is not practical to use this as an attribute for other resources.
* `event_delivery_failure_topic_arn` - (Optional) SNS Topic triggered when a delivery to any of the platform endpoints associated with your platform application encounters a permanent failure.
* `event_endpoint_created_topic_arn` - (Optional) SNS Topic triggered when a new platform endpoint is added to your platform application.
* `event_endpoint_deleted_topic_arn` - (Optional) SNS Topic triggered when an existing platform endpoint is deleted from your platform application.
* `event_endpoint_updated_topic_arn` - (Optional) SNS Topic triggered when an existing platform endpoint is changed from your platform application.
* `failure_feedback_role_arn` - (Optional) The IAM role permitted to receive failure feedback for this application.
* `platform_principal` - (Optional) Application Platform principal. See [Principal][2] for type of principal required for platform. The value of this attribute when stored into the Terraform state is only a hash of the real value, so therefore it is not practical to use this as an attribute for other resources.
* `success_feedback_role_arn` - (Optional) The IAM role permitted to receive success feedback for this application.
* `success_feedback_sample_rate` - (Optional) The percentage of success to sample (0-100)

## Attributes Reference

The following additional attributes are exported:

* `id` - The ARN of the SNS platform application
* `arn` - The ARN of the SNS platform application

[1]: http://docs.aws.amazon.com/sns/latest/dg/mobile-push-send-register.html
[2]: http://docs.aws.amazon.com/sns/latest/api/API_CreatePlatformApplication.html

## Import

SNS platform applications can be imported using the ARN, e.g.

```
$ terraform import aws_sns_platform_application.gcm_application arn:aws:sns:us-west-2:0123456789012:app/GCM/gcm_application
```
