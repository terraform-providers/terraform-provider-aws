---
subcategory: "SNS"
layout: "aws"
page_title: "AWS: aws_sns_topic_subscription"
description: |-
  Provides a resource for subscribing to SNS topics.
---

# Resource: aws_sns_topic_subscription

Provides a resource for subscribing to SNS topics. Requires that an SNS topic exist for the subscription to attach to. This resource allows you to automatically place messages sent to SNS topics in SQS queues, send them as HTTP(S) POST requests to a given endpoint, send SMS messages, or notify devices / applications. The most likely use case for Terraform users will probably be SQS queues.

~> **NOTE:** If the SNS topic and SQS queue are in different AWS regions, the `aws_sns_topic_subscription` must use an AWS provider that is in the same region as the SNS topic. If the `aws_sns_topic_subscription` uses a provider with a different region than the SNS topic, Terraform will fail to create the subscription.

~> **NOTE:** Setup of cross-account subscriptions from SNS topics to SQS queues requires Terraform to have access to BOTH accounts.

~> **NOTE:** If an SNS topic and SQS queue are in different AWS accounts but the same region, the `aws_sns_topic_subscription` must use the AWS provider for the account with the SQS queue. If `aws_sns_topic_subscription` uses a Provider with a different account than the SQS queue, Terraform creates the subscription but does not keep state and tries to re-create the subscription at every `apply`.

~> **NOTE:** If an SNS topic and SQS queue are in different AWS accounts and different AWS regions, the subscription needs to be initiated from the account with the SQS queue but in the region of the SNS topic.

~> **NOTE:** You cannot unsubscribe to a subscription that is pending confirmation. If you use `email`, `email-json`, or `http`/`https` (without auto-confirmation enabled), until the subscription is confirmed (e.g., outside of Terraform), AWS does not allow Terraform to delete / unsubscribe the subscription. If you `destroy` an unconfirmed subscription, Terraform will remove the subscription from its state but the subscription will still exist in AWS. However, if you delete an SNS topic, SNS [deletes all the subscriptions](https://docs.aws.amazon.com/sns/latest/dg/sns-delete-subscription-topic.html) associated with the topic. Also, you can import a subscription after confirmation and then have the capability to delete it.

## Example Usage

You can directly supply a topic and ARN by hand in the `topic_arn` property along with the queue ARN:

```terraform
resource "aws_sns_topic_subscription" "user_updates_sqs_target" {
  topic_arn = "arn:aws:sns:us-west-2:432981146916:user-updates-topic"
  protocol  = "sqs"
  endpoint  = "arn:aws:sqs:us-west-2:432981146916:terraform-queue-too"
}
```

Alternatively you can use the ARN properties of a managed SNS topic and SQS queue:

```terraform
resource "aws_sns_topic" "user_updates" {
  name = "user-updates-topic"
}

resource "aws_sqs_queue" "user_updates_queue" {
  name = "user-updates-queue"
}

resource "aws_sns_topic_subscription" "user_updates_sqs_target" {
  topic_arn = aws_sns_topic.user_updates.arn
  protocol  = "sqs"
  endpoint  = aws_sqs_queue.user_updates_queue.arn
}
```

You can subscribe SNS topics to SQS queues in different Amazon accounts and regions:

```terraform
variable "sns" {
  default = {
    account-id   = "111111111111"
    role-name    = "service/service-hashicorp-terraform"
    name         = "example-sns-topic"
    display_name = "example"
    region       = "us-west-1"
  }
}

variable "sqs" {
  default = {
    account-id = "222222222222"
    role-name  = "service/service-hashicorp-terraform"
    name       = "example-sqs-queue"
    region     = "us-east-1"
  }
}

data "aws_iam_policy_document" "sns-topic-policy" {
  policy_id = "__default_policy_ID"

  statement {
    actions = [
      "SNS:Subscribe",
      "SNS:SetTopicAttributes",
      "SNS:RemovePermission",
      "SNS:Publish",
      "SNS:ListSubscriptionsByTopic",
      "SNS:GetTopicAttributes",
      "SNS:DeleteTopic",
      "SNS:AddPermission",
    ]

    condition {
      test     = "StringEquals"
      variable = "AWS:SourceOwner"

      values = [
        var.sns["account-id"],
      ]
    }

    effect = "Allow"

    principals {
      type        = "AWS"
      identifiers = ["*"]
    }

    resources = [
      "arn:aws:sns:${var.sns["region"]}:${var.sns["account-id"]}:${var.sns["name"]}",
    ]

    sid = "__default_statement_ID"
  }

  statement {
    actions = [
      "SNS:Subscribe",
      "SNS:Receive",
    ]

    condition {
      test     = "StringLike"
      variable = "SNS:Endpoint"

      values = [
        "arn:aws:sqs:${var.sqs["region"]}:${var.sqs["account-id"]}:${var.sqs["name"]}",
      ]
    }

    effect = "Allow"

    principals {
      type        = "AWS"
      identifiers = ["*"]
    }

    resources = [
      "arn:aws:sns:${var.sns["region"]}:${var.sns["account-id"]}:${var.sns["name"]}",
    ]

    sid = "__console_sub_0"
  }
}

data "aws_iam_policy_document" "sqs-queue-policy" {
  policy_id = "arn:aws:sqs:${var.sqs["region"]}:${var.sqs["account-id"]}:${var.sqs["name"]}/SQSDefaultPolicy"

  statement {
    sid    = "example-sns-topic"
    effect = "Allow"

    principals {
      type        = "AWS"
      identifiers = ["*"]
    }

    actions = [
      "SQS:SendMessage",
    ]

    resources = [
      "arn:aws:sqs:${var.sqs["region"]}:${var.sqs["account-id"]}:${var.sqs["name"]}",
    ]

    condition {
      test     = "ArnEquals"
      variable = "aws:SourceArn"

      values = [
        "arn:aws:sns:${var.sns["region"]}:${var.sns["account-id"]}:${var.sns["name"]}",
      ]
    }
  }
}

# provider to manage SNS topics
provider "aws" {
  alias  = "sns"
  region = var.sns["region"]

  assume_role {
    role_arn     = "arn:aws:iam::${var.sns["account-id"]}:role/${var.sns["role-name"]}"
    session_name = "sns-${var.sns["region"]}"
  }
}

# provider to manage SQS queues
provider "aws" {
  alias  = "sqs"
  region = var.sqs["region"]

  assume_role {
    role_arn     = "arn:aws:iam::${var.sqs["account-id"]}:role/${var.sqs["role-name"]}"
    session_name = "sqs-${var.sqs["region"]}"
  }
}

# provider to subscribe SQS to SNS (using the SQS account but the SNS region)
provider "aws" {
  alias  = "sns2sqs"
  region = var.sns["region"]

  assume_role {
    role_arn     = "arn:aws:iam::${var.sqs["account-id"]}:role/${var.sqs["role-name"]}"
    session_name = "sns2sqs-${var.sns["region"]}"
  }
}

resource "aws_sns_topic" "sns-topic" {
  provider     = "aws.sns"
  name         = var.sns["name"]
  display_name = var.sns["display_name"]
  policy       = data.aws_iam_policy_document.sns-topic-policy.json
}

resource "aws_sqs_queue" "sqs-queue" {
  provider = "aws.sqs"
  name     = var.sqs["name"]
  policy   = data.aws_iam_policy_document.sqs-queue-policy.json
}

resource "aws_sns_topic_subscription" "sns-topic" {
  provider  = "aws.sns2sqs"
  topic_arn = aws_sns_topic.sns-topic.arn
  protocol  = "sqs"
  endpoint  = aws_sqs_queue.sqs-queue.arn
}
```

## Argument Reference

The following arguments are required:

* `endpoint` - (Required) Endpoint to send data to. The contents vary with the protocol. See details below.
* `protocol` - (Required) Protocol to use. Valid values are: `sqs`, `sms`, `lambda`, `firehose`, and `application`. Protocols `email`, `email-json`, `http` and `https` are also valid but partially supported. See details below.
* `subscription_role_arn` - (Required if `protocol` is `firehose`) ARN of the IAM role to publish to Kinesis Data Firehose delivery stream. Refer to [SNS docs](https://docs.aws.amazon.com/sns/latest/dg/sns-firehose-as-subscriber.html).
* `topic_arn` - (Required) ARN of the SNS topic to subscribe to.

The following arguments are optional:

* `confirmation_timeout_in_minutes` - (Optional) Integer indicating number of minutes to wait in retrying mode for fetching subscription arn before marking it as failure. Only applicable for http and https protocols. Default is `1`.
* `delivery_policy` - (Optional) JSON String with the delivery policy (retries, backoff, etc.) that will be used in the subscription - this only applies to HTTP/S subscriptions. Refer to the [SNS docs](https://docs.aws.amazon.com/sns/latest/dg/DeliveryPolicies.html) for more details.
* `endpoint_auto_confirms` - (Optional) Whether the endpoint is capable of [auto confirming subscription](http://docs.aws.amazon.com/sns/latest/dg/SendMessageToHttp.html#SendMessageToHttp.prepare) (e.g., PagerDuty). Default is `false`.
* `filter_policy` - (Optional) JSON String with the filter policy that will be used in the subscription to filter messages seen by the target resource. Refer to the [SNS docs](https://docs.aws.amazon.com/sns/latest/dg/message-filtering.html) for more details.
* `raw_message_delivery` - (Optional) Whether to enable raw message delivery (the original message is directly passed, not wrapped in JSON with the original message in the message property). Default is `false`.
* `redrive_policy` - (Optional) JSON String with the redrive policy that will be used in the subscription. Refer to the [SNS docs](https://docs.aws.amazon.com/sns/latest/dg/sns-dead-letter-queues.html#how-messages-moved-into-dead-letter-queue) for more details.

### Protocol support

Supported values for `protocol` include:

* `application` - Delivers JSON-encoded messages. `endpoint` is the endpoint ARN of a mobile app and device.
* `firehose` - Delivers JSON-encoded messages. `endpoint` is the ARN of an Amazon Kinesis Data Firehose delivery stream (e.g.,
`arn:aws:firehose:us-east-1:123456789012:deliverystream/ticketUploadStream`).
* `lambda` - Delivers JSON-encoded messages. `endpoint` is the ARN of an AWS Lambda function.
* `sms` - Delivers text messages via SMS. `endpoint` is the phone number of an SMS-enabled device.
* `sqs` - Delivers JSON-encoded messages. `endpoint` is the ARN of an Amazon SQS queue (e.g., `arn:aws:sqs:us-west-2:123456789012:terraform-queue-too`).

Partially supported values for `protocol` include:

~> **NOTE:** If an `aws_sns_topic_subscription` uses a partially-supported protocol and the subscription is not confirmed, either through automatic confirmation or means outside of Terraform (e.g., clicking on a "Confirm Subscription" link in an email), Terraform cannot delete / unsubscribe the subscription. Attempting to `destroy` an unconfirmed subscription will remove the `aws_sns_topic_subscription` from Terraform's state but **_will not_** remove the subscription from AWS. The `pending_confirmation` attribute provides confirmation status.

* `email` - Delivers messages via SMTP. `endpoint` is an email address.
* `email-json` - Delivers JSON-encoded messages via SMTP. `endpoint` is an email address.
* `http` -- Delivers JSON-encoded messages via HTTP POST. `endpoint` is a URL beginning with `http://`.
* `https` -- Delivers JSON-encoded messages via HTTPS POST. `endpoint` is a URL beginning with `https://`.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - ARN of the subscription.
* `confirmation_was_authenticated` - Whether the subscription confirmation request was authenticated.
* `id` - ARN of the subscription.
* `owner_id` - AWS account ID of the subscription's owner.
* `pending_confirmation` - Whether the subscription has not been confirmed.

## Import

SNS Topic Subscriptions can be imported using the `subscription arn`, e.g.

```
$ terraform import aws_sns_topic_subscription.user_updates_sqs_target arn:aws:sns:us-west-2:0123456789012:my-topic:8a21d249-4329-4871-acc6-7be709c6ea7f
```
