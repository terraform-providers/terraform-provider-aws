---
layout: "aws"
page_title: "AWS: aws_lambda_alias"
sidebar_current: "docs-aws-resource-lambda-alias"
description: |-
  Creates a Lambda function alias.
---

# aws_lambda_alias

Creates a Lambda function alias. Creates an alias that points to the specified Lambda function version.

For information about Lambda and how to use it, see [What is AWS Lambda?][1]
For information about function aliases, see [CreateAlias][2] and [AliasRoutingConfiguration][3] in the API docs.

## Example Usage

```hcl
resource "aws_lambda_alias" "test_alias" {
  name             = "testalias"
  description      = "a sample description"
  function_name    = "${aws_lambda_function.lambda_function_test.arn}"
  function_version = "1"

  routing_config = {
    additional_version_weights = {
      "2" = 0.5
    }
  }
}
```

## Argument Reference

* `name` - (Required) Name for the alias you are creating. Pattern: `(?!^[0-9]+$)([a-zA-Z0-9-_]+)`
* `description` - (Optional) Description of the alias.
* `function_name` - (Required) The function ARN of the Lambda function for which you want to create an alias.
* `function_version` - (Required) Lambda function version for which you are creating the alias. Pattern: `(\$LATEST|[0-9]+)`.
* `routing_config` - (Optional) The Lambda alias' route configuration settings. Fields documented below

For **routing_config** the following attributes are supported:

* `additional_version_weights` - (Optional) A map that defines the proportion of events that should be sent to different versions of a lambda function.

## Attributes Reference

* `arn` - The Amazon Resource Name (ARN) identifying your Lambda function alias.
* `invoke_arn` - The ARN to be used for invoking Lambda Function from API Gateway - to be used in [`aws_api_gateway_integration`](/docs/providers/aws/r/api_gateway_integration.html)'s `uri`

[1]: http://docs.aws.amazon.com/lambda/latest/dg/welcome.html
[2]: http://docs.aws.amazon.com/lambda/latest/dg/API_CreateAlias.html
[3]: https://docs.aws.amazon.com/lambda/latest/dg/API_AliasRoutingConfiguration.html
