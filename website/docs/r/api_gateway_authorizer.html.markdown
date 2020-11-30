---
subcategory: "API Gateway (REST APIs)"
layout: "aws"
page_title: "AWS: aws_api_gateway_authorizer"
description: |-
  Provides an API Gateway Authorizer.
---

# Resource: aws_api_gateway_authorizer

Provides an API Gateway Authorizer.

## Example Usage

```hcl
resource "aws_api_gateway_authorizer" "demo" {
  name                   = "demo"
  rest_api_id            = aws_api_gateway_rest_api.demo.id
  authorizer_uri         = aws_lambda_function.authorizer.invoke_arn
  authorizer_credentials = aws_iam_role.invocation_role.arn
}

resource "aws_api_gateway_rest_api" "demo" {
  name = "auth-demo"
}

resource "aws_iam_role" "invocation_role" {
  name = "api_gateway_auth_invocation"
  path = "/"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "apigateway.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "invocation_policy" {
  name = "default"
  role = aws_iam_role.invocation_role.id

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "lambda:InvokeFunction",
      "Effect": "Allow",
      "Resource": "${aws_lambda_function.authorizer.arn}"
    }
  ]
}
EOF
}

resource "aws_iam_role" "lambda" {
  name = "demo-lambda"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "lambda.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_lambda_function" "authorizer" {
  filename      = "lambda-function.zip"
  function_name = "api_gateway_authorizer"
  role          = aws_iam_role.lambda.arn
  handler       = "exports.example"

  source_code_hash = filebase64sha256("lambda-function.zip")
}
```

## Argument Reference

The following arguments are supported:

* `authorizer_uri` - (Optional, required for type `TOKEN`/`REQUEST`) The authorizer's Uniform Resource Identifier (URI). This must be a well-formed Lambda function URI in the form of `arn:aws:apigateway:{region}:lambda:path/{service_api}`,
 e.g. `arn:aws:apigateway:us-west-2:lambda:path/2015-03-31/functions/arn:aws:lambda:us-west-2:012345678912:function:my-function/invocations`
* `name` - (Required) The name of the authorizer
* `rest_api_id` - (Required) The ID of the associated REST API
* `identity_source` - (Optional) The source of the identity in an incoming request. Defaults to `method.request.header.Authorization`. For `REQUEST` type, this may be a comma-separated list of values, including headers, query string parameters and stage variables - e.g. `"method.request.header.SomeHeaderName,method.request.querystring.SomeQueryStringName,stageVariables.SomeStageVariableName"`
* `type` - (Optional) The type of the authorizer. Possible values are `TOKEN` for a Lambda function using a single authorization token submitted in a custom header, `REQUEST` for a Lambda function using incoming request parameters, or `COGNITO_USER_POOLS` for using an Amazon Cognito user pool. Defaults to `TOKEN`.
* `authorizer_credentials` - (Optional) The credentials required for the authorizer. To specify an IAM Role for API Gateway to assume, use the IAM Role ARN.
* `authorizer_result_ttl_in_seconds` - (Optional) The TTL of cached authorizer results in seconds. Defaults to `300`.
* `identity_validation_expression` - (Optional) A validation expression for the incoming identity. For `TOKEN` type, this value should be a regular expression. The incoming token from the client is matched against this expression, and will proceed if the token matches. If the token doesn't match, the client receives a 401 Unauthorized response.
* `provider_arns` - (Optional, required for type `COGNITO_USER_POOLS`) A list of the Amazon Cognito user pool ARNs. Each element is of this format: `arn:aws:cognito-idp:{region}:{account_id}:userpool/{user_pool_id}`.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The Authorizer identifier.

## Import

AWS API Gateway Authorizer can be imported using the `REST-API-ID/AUTHORIZER-ID`, e.g.

```sh
$ terraform import aws_api_gateway_authorizer.authorizer 12345abcde/example
```
