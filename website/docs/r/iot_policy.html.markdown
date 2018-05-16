---
layout: "aws"
page_title: "AWS: aws_iot_policy"
sidebar_current: "docs-aws-resource-iot-policy"
description: |-
  Provides an IoT policy.
---

# aws_iot_policy

Provides an IoT policy.

## Example Usage

```hcl
resource "aws_iot_policy" "pubsub" {
  name        = "PubSubToAnyTopic"
  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "iot:*"
      ],
      "Effect": "Allow",
      "Resource": "*"
    }
  ]
}
EOF
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the policy.
* `policy` - (Required) The policy document. This is a JSON formatted string.
  The heredoc syntax or `file` function is helpful here. Use the [IoT Developer Guide]
  (http://docs.aws.amazon.com/iot/latest/developerguide/iot-policies.html) for more information on IoT Policies

## Attributes Reference

The following attributes are exported:

* `arn` - The ARN assigned by AWS to this policy.
* `name` - The name of this policy.
* `default_version_id` - The default version of this policy.
* `policy` - The policy document.
