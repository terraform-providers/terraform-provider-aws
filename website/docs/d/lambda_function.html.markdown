---
subcategory: "Lambda"
layout: "aws"
page_title: "AWS: aws_lambda_function"
description: |-
  Provides a Lambda Function data source.
---

# aws_lambda_function

Provides information about a Lambda Function.

## Example Usage

```hcl
variable "function_name" {
  type = string
}

data "aws_lambda_function" "existing" {
  function_name = var.function_name
}
```

## Argument Reference

The following arguments are supported:

* `function_name` - (Required) Name of the lambda function.
* `qualifier` - (Optional) Alias name or version number of the lambda function. e.g. `$LATEST`, `my-alias`, or `1`

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - Unqualified (no `:QUALIFIER` or `:VERSION` suffix) Amazon Resource Name (ARN) identifying your Lambda Function. See also `qualified_arn`.
* `code_signing_config_arn` - Amazon Resource Name (ARN) for a Code Signing Configuration.
* `dead_letter_config` - Configure the function's *dead letter queue*.
* `description` - Description of what your Lambda Function does.
* `environment` - The Lambda environment's configuration settings.
* `file_system_config` - The connection settings for an Amazon EFS file system.
* `handler` - The function entrypoint in your code.
* `invoke_arn` - The ARN to be used for invoking Lambda Function from API Gateway.
* `kms_key_arn` - The ARN for the KMS encryption key.
* `last_modified` - The date this resource was last modified.
* `layers` - A list of Lambda Layer ARNs attached to your Lambda Function.
* `memory_size` - Amount of memory in MB your Lambda Function can use at runtime.
* `qualified_arn` - Qualified (`:QUALIFIER` or `:VERSION` suffix) Amazon Resource Name (ARN) identifying your Lambda Function. See also `arn`.
* `reserved_concurrent_executions` - The amount of reserved concurrent executions for this lambda function or `-1` if unreserved.
* `role` - IAM role attached to the Lambda Function.
* `runtime` - The runtime environment for the Lambda function.
* `signing_job_arn` - The Amazon Resource Name (ARN) of a signing job.
* `signing_profile_version_arn` - The Amazon Resource Name (ARN) for a signing profile version.
* `source_code_hash` - Base64-encoded representation of raw SHA-256 sum of the zip file.
* `source_code_size` - The size in bytes of the function .zip file.
* `timeout` - The function execution time at which Lambda should terminate the function.
* `tracing_config` - Tracing settings of the function.
* `version` - The version of the Lambda function.
* `vpc_config` - VPC configuration associated with your Lambda function.
