---
subcategory: "Lambda"
layout: "aws"
page_title: "AWS: aws_lambda_function_event_invoke_config"
description: |-
  Manages an asynchronous invocation configuration for a Lambda Function or Alias.
---

# Resource: aws_lambda_function_event_invoke_config

Manages an asynchronous invocation configuration for a Lambda Function or Alias. More information about asynchronous invocations and the configurable values can be found in the [Lambda Developer Guide](https://docs.aws.amazon.com/lambda/latest/dg/invocation-async.html).

## Example Usage

### Destination Configuration

~> **NOTE:** Ensure the Lambda Function IAM Role has necessary permissions for the destination, such as `sqs:SendMessage` or `sns:Publish`, otherwise the API will return a generic `InvalidParameterValueException: The destination ARN arn:PARTITION:SERVICE:REGION:ACCOUNT:RESOURCE is invalid.` error.

```hcl
resource "aws_lambda_function_event_invoke_config" "example" {
  function_name = aws_lambda_alias.example.function_name

  destination_config {
    on_failure {
      destination = aws_sqs_queue.example.arn
    }

    on_success {
      destination = aws_sns_topic.example.arn
    }
  }
}
```

### Error Handling Configuration

```hcl
resource "aws_lambda_function_event_invoke_config" "example" {
  function_name                = aws_lambda_alias.example.function_name
  maximum_event_age_in_seconds = 60
  maximum_retry_attempts       = 0
}
```

### Configuration for Alias Name

```hcl
resource "aws_lambda_function_event_invoke_config" "example" {
  function_name = aws_lambda_alias.example.function_name
  qualifier     = aws_lambda_alias.example.name

  # ... other configuration ...
}
```

### Configuration for Function Latest Unpublished Version

```hcl
resource "aws_lambda_function_event_invoke_config" "example" {
  function_name = aws_lambda_function.example.function_name
  qualifier     = "$LATEST"

  # ... other configuration ...
}
```

### Configuration for Function Published Version

```hcl
resource "aws_lambda_function_event_invoke_config" "example" {
  function_name = aws_lambda_function.example.function_name
  qualifier     = aws_lambda_function.example.version

  # ... other configuration ...
}
```

## Argument Reference

The following arguments are required:

* `function_name` - (Required) Name or Amazon Resource Name (ARN) of the Lambda Function, omitting any version or alias qualifier.

The following arguments are optional:

* `destination_config` - (Optional) Configuration block with destination configuration. See below for details.
* `maximum_event_age_in_seconds` - (Optional) Maximum age of a request that Lambda sends to a function for processing in seconds. Valid values between 60 and 21600.
* `maximum_retry_attempts` - (Optional) Maximum number of times to retry when the function returns an error. Valid values between 0 and 2. Defaults to 2.
* `qualifier` - (Optional) Lambda Function published version, `$LATEST`, or Lambda Alias name.

### destination_config Configuration Block

~> **NOTE:** At least one of `on_failure` or `on_success` must be configured when using this configuration block, otherwise remove it completely to prevent perpetual differences in Terraform runs.

The following arguments are optional:

* `on_failure` - (Optional) Configuration block with destination configuration for failed asynchronous invocations. See below for details.
* `on_success` - (Optional) Configuration block with destination configuration for successful asynchronous invocations. See below for details.

#### destination_config on_failure Configuration Block

The following arguments are required:

* `destination` - (Required) Amazon Resource Name (ARN) of the destination resource. See the [Lambda Developer Guide](https://docs.aws.amazon.com/lambda/latest/dg/invocation-async.html#invocation-async-destinations) for acceptable resource types and associated IAM permissions.

#### destination_config on_success Configuration Block

The following arguments are required:

* `destination` - (Required) Amazon Resource Name (ARN) of the destination resource. See the [Lambda Developer Guide](https://docs.aws.amazon.com/lambda/latest/dg/invocation-async.html#invocation-async-destinations) for acceptable resource types and associated IAM permissions.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - Fully qualified Lambda Function name or Amazon Resource Name (ARN)

## Import

Lambda Function Event Invoke Configs can be imported using the fully qualified Function name or Amazon Resource Name (ARN), e.g.

ARN without qualifier (all versions and aliases):

```
$ terraform import aws_lambda_function_event_invoke_config.example arn:aws:us-east-1:123456789012:function:my_function
```

ARN with qualifier:

```
$ terraform import aws_lambda_function_event_invoke_config.example arn:aws:us-east-1:123456789012:function:my_function:production
```

Name without qualifier (all versions and aliases):

```
$ terraform import aws_lambda_function_event_invoke_config.example my_function
```

Name with qualifier:

```
$ terraform import aws_lambda_function_event_invoke_config.example my_function:production
```
