---
subcategory: "Lambda"
layout: "aws"
page_title: "AWS: aws_lambda_provisioned_concurrency_config"
description: |-
  Manages a Lambda Provisioned Concurrency Configuration
---

# Resource: aws_lambda_provisioned_concurrency_config

Manages a Lambda Provisioned Concurrency Configuration.

## Example Usage

### Alias Name

```hcl
resource "aws_lambda_provisioned_concurrency_config" "example" {
  function_name                     = aws_lambda_alias.example.function_name
  provisioned_concurrent_executions = 1
  qualifier                         = aws_lambda_alias.example.name
}
```

### Function Version

```hcl
resource "aws_lambda_provisioned_concurrency_config" "example" {
  function_name                     = aws_lambda_function.example.function_name
  provisioned_concurrent_executions = 1
  qualifier                         = aws_lambda_function.example.version
}
```

## Argument Reference

The following arguments are required:

* `function_name` - (Required) Name or Amazon Resource Name (ARN) of the Lambda Function.
* `provisioned_concurrent_executions` - (Required) Amount of capacity to allocate. Must be greater than or equal to `1`.
* `qualifier` - (Required) Lambda Function version or Lambda Alias name.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - Lambda Function name and qualifier separated by a colon (`:`).

## Timeouts

`aws_lambda_provisioned_concurrency_config` provides the following [Timeouts](https://www.terraform.io/docs/configuration/blocks/resources/syntax.html#operation-timeouts) configuration options:

* `create` - (Default `15 minutes`) How long to wait for the Lambda Provisioned Concurrency Config to be ready on creation.
* `update` - (Default `15 minutes`) How long to wait for the Lambda Provisioned Concurrency Config to be ready on update.

## Import

Lambda Provisioned Concurrency Configs can be imported using the `function_name` and `qualifier` separated by a colon (`:`), e.g.

```
$ terraform import aws_lambda_provisioned_concurrency_config.example my_function:production
```
