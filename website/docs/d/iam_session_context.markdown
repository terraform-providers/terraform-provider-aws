---
subcategory: "IAM"
layout: "aws"
page_title: "AWS: aws_iam_session_context"
description: |-
  Get information on the IAM source role of an STS assumed role
---

# Data Source: aws_iam_session_context

This data source provides information on the IAM source role of an STS assumed role. For non-role ARNs, this data source simply passes the ARN through in `issuer_arn`.

For some AWS resources, multiple types of principals are allowed in the same argument (e.g., IAM users and IAM roles). However, these arguments often do not allow assumed-role (i.e., STS, temporary credential) principals. Given an STS ARN, this data source provides the ARN for the source IAM role.

## Example Usage

### Basic Example

```terraform
data "aws_iam_session_context" "example" {
  arn = "arn:aws:sts::123456789012:assumed-role/Audien-Heaven/MatyNoyes"
}
```

### Find the Terraform Runner's Source Role

Combined with `aws_caller_identity`, you can get the current user's source IAM role ARN (`issuer_arn`) if you're using an assumed role. If you're not using an assumed role, the caller's (e.g., an IAM user's) ARN will simply be passed through. In environments where both IAM users and individuals using assumed roles need to apply the same configurations, this data source enables seamless use.

```terraform
data "aws_caller_identity" "current" {}

data "aws_iam_session_context" "example" {
  arn = data.aws_caller_identity.current.arn
}
```

## Argument Reference

* `arn` - (Required) ARN for an assumed role.

~> If `arn` is a non-role ARN, Terraform gives no error and `issuer_arn` will be equal to the `arn` value. For STS assumed-role ARNs, Terraform gives an error if the identified IAM role does not exist.

## Attributes Reference

~> With the exception of `issuer_arn`, the attributes will not be populated unless the `arn` corresponds to an STS assumed role.

* `issuer_arn` - IAM source role ARN if `arn` corresponds to an STS assumed role. Otherwise, `issuer_arn` is equal to `arn`.
* `issuer_id` - Unique identifier of the IAM role that issues the STS assumed role.
* `issuer_name` - Name of the source role. Only available if `arn` corresponds to an STS assumed role.
* `session_name` - Name of the STS session. Only available if `arn` corresponds to an STS assumed role.
