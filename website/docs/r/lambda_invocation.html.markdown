---
subcategory: "Lambda"
layout: "aws"
page_title: "AWS: aws_lambda_invocation"
description: |-
Provide AWS Lambda Function result.
---

# Resource: aws_lambda_invocation

Use this resource to invoke custom lambda functions.
The lambda function is invoked with [RequestResponse](https://docs.aws.amazon.com/lambda/latest/dg/API_Invoke.html#API_Invoke_RequestSyntax)
invocation type.

## Example Usage

```hcl
resource "aws_lambda_invocation" "example" {
  function_name = aws_lambda_function.lambda_function_test.function_name

  triggers = {
    redeployment = sha1(jsonencode([
      aws_lambda_function.example
    ]))
  }
  
  input = <<JSON
{
  "key1": "value1",
  "key2": "value2"
}
JSON
}
output "result_entry" {
  value = jsondecode(aws_lambda_invocation.example.result)["key1"]
}
```

## Argument Reference

* `function_name` - (Required) The name of the lambda function.
* `input` - (Required) A string in JSON format that is passed as payload to the lambda function.
* `qualifier` - (Optional) The qualifier (a.k.a version) of the lambda function. Defaults
  to `$LATEST`.
* `triggers` - (Optional) A map of arbitrary keys and values that, when changed, will trigger a re-invocation. To force a re-invocation without changing these keys/values, use the [`terraform taint` command](https://www.terraform.io/docs/commands/taint.html).

## Attributes Reference

* `result` - String result of the lambda function invocation.