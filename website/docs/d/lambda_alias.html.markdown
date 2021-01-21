---
subcategory: "Lambda"
layout: "aws"
page_title: "AWS: aws_lambda_alias"
description: |-
  Provides a Lambda Alias data source.
---

# Data Source: aws_lambda_alias

Provides information about a Lambda Alias.

## Example Usage

```hcl
data "aws_lambda_alias" "production" {
  function_name = "my-lambda-func"
  name          = "production"
}
```

## Argument Reference

The following arguments are supported:

* `function_name` - (Required) Name of the aliased Lambda function.
* `name` - (Required) Name of the Lambda alias.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The Amazon Resource Name (ARN) identifying the Lambda function alias.
* `description` - Description of alias.
* `function_version` - Lambda function version which the alias uses.
* `invoke_arn` - The ARN to be used for invoking Lambda Function from API Gateway - to be used in aws_api_gateway_integration's `uri`.
