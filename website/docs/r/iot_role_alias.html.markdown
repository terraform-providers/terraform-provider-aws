---
layout: "aws"
page_title: "AWS: aws_iot_role_alias"
sidebar_current: "docs-aws-resource-iot-role_alias"
description: |-
  Provides an IoT role alias.
---

# aws_iot_role_alias

Provides an IoT role alias.

## Example Usage

```hcl
resource "aws_iam_role" "role" {
  name   = "dynamodb-access-role"
  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {"Service": "credentials.iot.amazonaws.com"},
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

resource "aws_iot_role_alias" "alias" {
  alias    = "Thermostat-dynamodb-access-role-alias"
  role_arn = "${aws_iam_role.role.arn}"
}
```

## Argument Reference

The following arguments are supported:

* `alias` - (Required) The name of the role alias.
* `role_arn` - (Required) The identity of the role to which the alias refers.
* `credential_duration` - (Optional) The duration of the credential, in seconds. If you do not specify a value for this setting, the default maximum of one hour is applied. This setting can have a value from 900 seconds (15 minutes) to 3600 seconds (60 minutes).

## Import

IOT Role Alias can be imported via the alias, e.g.

```sh
$ terraform import aws_iot_role_alias.example myalias
```
