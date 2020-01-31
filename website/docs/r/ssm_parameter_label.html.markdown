---
layout: "aws"
page_title: "AWS: aws_ssm_parameter_label"
sidebar_current: "docs-aws-resource-ssm-parameter-label"
description: |-
  Provides a SSM Parameter Label resource
---

# Resource: aws_ssm_parameter_label

Provides an SSM Parameter Label resource.

## Example Usage

To store a basic string parameter:

```hcl

resource "aws_ssm_parameter" "foo" {
  name  = "foo"
  type  = "String"
  value = "bar"
}

resource "aws_ssm_parameter_label" "foo" {
  ssm_parameter_name  = aws_ssm_parameter.foo.name
  ssm_parameter_version = aws_ssm_parameter.foo.version
  labels = ["%s"]
}

```

~> **Note:** SSM Parameter label cannot be deleted. It can only be moved to another version.

## Argument Reference

The following arguments are supported:

* `ssm_parameter_name` - (Required) The name of the parameter.
* `ssm_parameter_version` - (Required) The version of the parameter.
* `labels` - (Required) A list of labels. The labels must satisfy the constraint as described in [AWS SSM Parameter Label guide](https://docs.aws.amazon.com/systems-manager/latest/userguide/sysman-paramstore-labels.html)

## Import

SSM Parameter Label can be imported using the `parameter name` and `parameter version`, e.g.

```$ terraform import aws_ssm_parameter_label.my_param_label '/my_path/my_paramname|label'
```
