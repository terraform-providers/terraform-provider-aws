---
subcategory: "Lambda"
layout: "aws"
page_title: "AWS: aws_lambda_invocation"
description: |-
  Invoke AWS Lambda Function as data source
---

# Data Source: aws_lambda_invocation

Use this data source to invoke custom lambda functions as data source.
The lambda function is invoked with [RequestResponse](https://docs.aws.amazon.com/lambda/latest/dg/API_Invoke.html#API_Invoke_RequestSyntax)
invocation type.

## Example Usage

```hcl
data "aws_lambda_invocation" "example" {
  function_name = "${aws_lambda_function.lambda_function_test.function_name}"

  input = <<JSON
{
  "key1": "value1",
  "key2": "value2"
}
JSON
}

output "result" {
  description = "String result of Lambda execution"
  value       = "${data.aws_lambda_invocation.example.result}"
}

# In Terraform 0.11 and earlier, the result_map attribute can be used
# to convert a result JSON string to a map of string keys to string values.
output "result_entry_tf011" {
  value = "${data.aws_lambda_invocation.example.result_map["key1"]}"
}

# In Terraform 0.12 and later, the jsondecode() function can be used
# to convert a result JSON string to native Terraform types.
output "result_entry_tf012" {
  value = jsondecode(data.aws_lambda_invocation.example.result)["key1"]
}
```

## Argument Reference

 * `function_name` - (Required) The name of the lambda function.
 * `input` - (Required) A string in JSON format that is passed as payload to the lambda function.
 * `qualifier` - (Optional) The qualifier (a.k.a version) of the lambda function. Defaults
 to `$LATEST`.

## Attributes Reference

 * `result` - String result of the lambda function invocation.
 * `result_map` - This field is set only if result is a map of primitive types, where the map is string keys and string values. In Terraform 0.12 and later, use the [`jsondecode()` function](/docs/configuration/functions/jsondecode.html) with the `result` attribute instead to convert the result to all supported native Terraform types.
