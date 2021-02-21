---
subcategory: "IoT"
layout: "aws"
page_title: "AWS: aws_iot_policy"
description: |-
  Provides an IoT policy.
---

# Resource: aws_iot_policy

Provides an IoT policy.

## Example Usage

```hcl
resource "aws_iot_policy" "pubsub" {
  name = "PubSubToAnyTopic"

  # Terraform's "jsonencode" function converts a
  # Terraform expression result to valid JSON syntax.
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = [
          "iot:*",
        ]
        Effect   = "Allow"
        Resource = "*"
      },
    ]
  })
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the policy.
* `policy` - (Required) The policy document. This is a JSON formatted string. Use the [IoT Developer Guide](http://docs.aws.amazon.com/iot/latest/developerguide/iot-policies.html) for more information on IoT Policies. For more information about building AWS IAM policy documents with Terraform, see the [AWS IAM Policy Document Guide](https://learn.hashicorp.com/terraform/aws/iam-policy).

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The ARN assigned by AWS to this policy.
* `name` - The name of this policy.
* `default_version_id` - The default version of this policy.
* `policy` - The policy document.

## Import

IoT policies can be imported using the `name`, e.g.

```
$ terraform import aws_iot_policy.pubsub PubSubToAnyTopic
```
