---
subcategory: "EventBridge (CloudWatch Events)"
layout: "aws"
page_title: "AWS: aws_cloudwatch_event_bus_policy"
description: |-
  Provides a resource to create an EventBridge policy to support cross-account events.
---

# Resource: aws_cloudwatch_event_bus_policy

Provides a resource to create an EventBridge resource policy to support cross-account events.

~> **Note:** EventBridge was formerly known as CloudWatch Events. The functionality is identical.

~> **Note:** The cloudwatch eventbus policy resource is incompatible with the cloudwatch event permissions resource and will overwrite them.

## Example Usage

### Account Access

```hcl
data "aws_iam_policy_document" "test" {
  statement {
    sid    = "DevAccountAccess"
    effect = "Allow"
    actions = [
      "events:PutEvents",
    ]
    resources = [
      "arn:aws:events:eu-west-1:123456789012:event-bus/default"
    ]

    principals {
      type        = "AWS"
      identifiers = ["123456789012"]
    }
  }
}

resource "aws_cloudwatch_event_bus_policy" "test" {
  policy         = data.aws_iam_policy_document.test.json
  event_bus_name = aws_cloudwatch_event_bus.test.name
}
```

### Organization Access

```hcl
data "aws_iam_policy_document" "test" {
  statement {
    sid    = "OrganizationAccess"
    effect = "Allow"
    actions = [
      "events:DescribeRule",
      "events:ListRules",
      "events:ListTargetsByRule",
      "events:ListTagsForResource",
    ]
    resources = [
      "arn:aws:events:eu-west-1:123456789012:rule/*",
      "arn:aws:events:eu-west-1:123456789012:event-bus/default"
    ]

    principals {
      type        = "AWS"
      identifiers = ["*"]
    }

    condition {
      test     = "StringEquals"
      variable = "aws:PrincipalOrgID"
      values   = aws_organizations_organization.example.id
    }
  }
}

resource "aws_cloudwatch_event_bus_policy" "test" {
  policy         = data.aws_iam_policy_document.test.json
  event_bus_name = aws_cloudwatch_event_bus.test.name
}
```

### Multiple Statements

```hcl
data "aws_iam_policy_document" "test" {

  statement {
    sid    = "DevAccountAccess"
    effect = "Allow"
    actions = [
      "events:PutEvents",
    ]
    resources = [
      "arn:aws:events:eu-west-1:123456789012:event-bus/default"
    ]

    principals {
      type        = "AWS"
      identifiers = ["123456789012"]
    }
  }

  statement {
    sid    = "OrganizationAccess"
    effect = "Allow"
    actions = [
      "events:DescribeRule",
      "events:ListRules",
      "events:ListTargetsByRule",
      "events:ListTagsForResource",
    ]
    resources = [
      "arn:aws:events:eu-west-1:123456789012:rule/*",
      "arn:aws:events:eu-west-1:123456789012:event-bus/default"
    ]

    principals {
      type        = "AWS"
      identifiers = ["*"]
    }

    condition {
      test     = "StringEquals"
      variable = "aws:PrincipalOrgID"
      values   = aws_organizations_organization.example.id
    }
  }
}

resource "aws_cloudwatch_event_bus_policy" "test" {
  policy         = data.aws_iam_policy_document.test.json
  event_bus_name = aws_cloudwatch_event_bus.test.name
}
```

## Argument Reference

The following arguments are supported:

* `policy` - (Required) The text of the policy. For more information about building AWS IAM policy documents with Terraform, see the [AWS IAM Policy Document Guide](https://learn.hashicorp.com/terraform/aws/iam-policy).
* `event_bus_name` - (Optional) The event bus to set the permissions on. If you omit this, the permissions are set on the `default` event bus.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The name of the EventBridge event bus.

## Import

EventBridge permissions can be imported using the `event_bus_name`, e.g.

```shell
$ terraform import aws_cloudwatch_event_bus_policy.DevAccountAccess example-event-bus
```
