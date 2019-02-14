---
layout: "aws"
page_title: "AWS: aws_ssm_parameter"
sidebar_current: "docs-aws-datasource-ssm-parameter"
description: |-
  Provides a SSM Parameter datasource
---

# Data Source: aws_ssm_parameter

Provides an SSM Parameter data source.

## Example Usage

```hcl
data "aws_ssm_parameter" "foo" {
  name = "foo"
}
```

~> **Note:** The unencrypted value of a SecureString will be stored in the raw state as plain-text.
[Read more about sensitive data in state](/docs/state/sensitive-data.html).


~> **Note:** The data source is currently following the behavior of the [SSM API](https://docs.aws.amazon.com/sdk-for-go/api/service/ssm/#Parameter) to return a string value, regardless of parameter type. For type `StringList`, we can use [split()](https://www.terraform.io/docs/configuration/interpolation.html#split-delim-string-) built-in function to get values in a list. Example: `split(",", data.aws_ssm_parameter.subnets.value)`


## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the parameter.
* `with_decryption` - (Optional) Whether to return decrypted `SecureString` value. Defaults to `true`.


In addition to all arguments above, the following attributes are exported:

* `arn` - The ARN of the parameter.
* `name` - The name of the parameter.
* `type` - The type of the parameter. Valid types are `String`, `StringList` and `SecureString`.
* `value` - The value of the parameter.
