---
layout: "aws"
page_title: "AWS: sns_sms_preferences"
sidebar_current: "docs-aws-resource-sns-sms-preferences"
description: |-
  Provides a way to set SNS SMS preferences.
---

# aws_sns_sms_preferences

Provides a way to set SNS SMS preferences.

## Example Usage

```hcl
resource "aws_sns_sms_preferences" "update_sms_prefs" {}
```

## Argument Reference

The following arguments are supported:

* `monthly_spend_limit` - (Optional) The maximum amount in USD that you are willing to spend each month to send SMS messages.
* `delivery_status_iam_role_arn` - (Optional) The ARN of the IAM role that allows Amazon SNS to write logs about SMS deliveries in CloudWatch Logs.
* `delivery_status_success_sampling_rate` - (Optional) The percentage of successful SMS deliveries for which Amazon SNS will write logs in CloudWatch Logs. The value must be between 0 and 100.
* `default_sender_id` - (Optional) A string, such as your business brand, that is displayed as the sender on the receiving device.
* `default_sms_type` - (Optional) The type of SMS message that you will send by default. Possible values are: Promotional, Transactional
* `usage_report_s3_bucket` - (Optional) The name of the Amazon S3 bucket to receive daily SMS usage reports from Amazon SNS.
