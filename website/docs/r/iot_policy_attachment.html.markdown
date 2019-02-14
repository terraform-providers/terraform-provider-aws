---
layout: "aws"
page_title: "AWS: aws_iot_policy_attachment"
sidebar_current: "docs-aws-resource-iot-policy-attachment"
description: |-
  Provides an IoT policy attachment.
---

# aws_iot_policy_attachment

Provides an IoT policy attachment.

## Example Usage

```hcl
resource "aws_iot_policy" "pubsub" {
  name   = "PubSubToAnyTopic"
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

resource "aws_iot_certificate" "cert" {
  csr    = "${file("csr.pem")}"
  active = true
}

resource "aws_iot_policy_attachment" "att" {
  policy = "${aws_iot_policy.pubsub.name}"
  target = "${aws_iot_certificate.cert.arn}"
}
```

## Argument Reference

The following arguments are supported:

* `policy` - (Required) The name of the policy to attach.
* `target` - (Required) The identity to which the policy is attached.
