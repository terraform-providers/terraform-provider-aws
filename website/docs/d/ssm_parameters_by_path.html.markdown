---
subcategory: "SSM"
layout: "aws"
page_title: "AWS: aws_ssm_parameters_by_path"
description: |-
  Provides SSM Parameters by path
---

# Data Source: aws_ssm_parameters_by_path

Provides SSM Parameters by path.

## Example Usage

```hcl
data "aws_ssm_parameters_by_path" "foo" {
  path = "/foo"
}
```

~> **Note:** The unencrypted value of a SecureString will be stored in the raw state as plain-text.
[Read more about sensitive data in state](/docs/state/sensitive-data.html).


~> **Note:** The data source is currently following the behavior of the [SSM API](https://docs.aws.amazon.com/sdk-for-go/api/service/ssm/#Parameter) to return a string value, regardless of parameter type. For type `StringList`, we can use the built-in [split()](https://www.terraform.io/docs/configuration/functions/split.html) function to get values in a list. Example: `split(",", data.aws_ssm_parameter.subnets.value)`


## Argument Reference

The following arguments are supported:

* `path` - (Required) The prefix path of the parameter.
* `with_decryption` - (Optional) Whether to return decrypted `SecureString` value. Defaults to `true`.


In addition to all arguments above, the following attributes are exported:

* `arns` - The ARNs of the parameter.
* `names` - The names of the parameter.
* `types` - The types of the parameter. Valid types are `String`, `StringList` and `SecureString`.
* `values` - The value of the parameter.
