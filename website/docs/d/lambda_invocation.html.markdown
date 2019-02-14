---
layout: "aws"
page_title: "AWS: aws_lambda_invocation"
sidebar_current: "docs-aws-datasource-lambda-invocation"
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

output "result_entry" {
  value = "${data.aws_lambda_invocation.result_map["key1"]}"
}

output "result" {
  value = "${data.aws_lambda_invocation.result}"
}
```

## Argument Reference

 * `function_name` - (Required) The name of the lambda function.
 * `input` - (Required) A string in JSON format that is passed as payload to the lambda function.
 * `qualifier` - (Optional) The qualifier (a.k.a version) of the lambda function. Defaults
 to `$LATEST`.

## Attributes Reference

 * `result` - A result of the lambda function invocation.
 * `result_map` - This field is set only if result is a map of primitive types.
