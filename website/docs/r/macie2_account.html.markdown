---
subcategory: "Macie"
layout: "aws"
page_title: "AWS: aws_macie2_account"
description: |-
  Provides a resource to manage Amazon Macie on an AWS Account.
---

# Resource: aws_macie2_account

Provides a resource to manage an [AWS Macie Account](https://docs.aws.amazon.com/macie/latest/APIReference/macie.html).

## Example Usage

```terraform
resource "aws_macie2_account" "test" {
  finding_publishing_frequency = "FIFTEEN_MINUTES"
  status                       = "ENABLED"
}
```

## Argument Reference

The following arguments are supported:

* `finding_publishing_frequency` -  (Optional) Specifies how often to publish updates to policy findings for the account. This includes publishing updates to AWS Security Hub and Amazon EventBridge (formerly called Amazon CloudWatch Events). Valid values are `FIFTEEN_MINUTES`, `ONE_HOUR` or `SIX_HOURS`.
* `status` - (Optional) Specifies the status for the account. To enable Amazon Macie and start all Macie activities for the account, set this value to `ENABLED`. Valid values are `ENABLED` or `PAUSED`.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The unique identifier (ID) of the macie account.
* `service_role` - The Amazon Resource Name (ARN) of the service-linked role that allows Macie to monitor and analyze data in AWS resources for the account.
* `created_at` - The date and time, in UTC and extended RFC 3339 format, when the Amazon Macie account was created.
* `updated_at` - The date and time, in UTC and extended RFC 3339 format, of the most recent change to the status of the Macie account.

## Import

`aws_macie2_account` can be imported using the id, e.g.

```
$ terraform import aws_macie2_account.example abcd1
```
