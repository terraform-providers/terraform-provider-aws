---
subcategory: ""
layout: "aws"
page_title: "Terraform AWS Provider Version 2 Upgrade Guide"
description: |-
  Terraform AWS Provider Version 2 Upgrade Guide
---

# Terraform AWS Provider Version 2 Upgrade Guide

Version 2.0.0 of the AWS provider for Terraform is a major release and includes some changes that you will need to consider when upgrading. This guide is intended to help with that process and focuses only on changes from version 1.60.0 to version 2.0.0.

Most of the changes outlined in this guide have been previously marked as deprecated in the Terraform plan/apply output throughout previous provider releases. These changes, such as deprecation notices, can always be found in the [Terraform AWS Provider CHANGELOG](https://github.com/hashicorp/terraform-provider-aws/blob/main/CHANGELOG.md).

Upgrade topics:

<!-- TOC depthFrom:2 depthTo:2 -->

- [Provider Version Configuration](#provider-version-configuration)
- [Provider: Configuration](#provider-configuration)
- [Data Source: aws_ami](#data-source-aws_ami)
- [Data Source: aws_ami_ids](#data-source-aws_ami_ids)
- [Data Source: aws_iam_role](#data-source-aws_iam_role)
- [Data Source: aws_kms_secret](#data-source-aws_kms_secret)
- [Data Source: aws_lambda_function](#data-source-aws_lambda_function)
- [Data Source: aws_region](#data-source-aws_region)
- [Resource: aws_api_gateway_api_key](#resource-aws_api_gateway_api_key)
- [Resource: aws_api_gateway_integration](#resource-aws_api_gateway_integration)
- [Resource: aws_api_gateway_integration_response](#resource-aws_api_gateway_integration_response)
- [Resource: aws_api_gateway_method](#resource-aws_api_gateway_method)
- [Resource: aws_api_gateway_method_response](#resource-aws_api_gateway_method_response)
- [Resource: aws_appautoscaling_policy](#resource-aws_appautoscaling_policy)
- [Resource: aws_autoscaling_policy](#resource-aws_autoscaling_policy)
- [Resource: aws_batch_compute_environment](#resource-aws_batch_compute_environment)
- [Resource: aws_cloudfront_distribution](#resource-aws_cloudfront_distribution)
- [Resource: aws_cognito_user_pool](#resource-aws_cognito_user_pool)
- [Resource: aws_dx_lag](#resource-aws_dx_lag)
- [Resource: aws_ecs_service](#resource-aws_ecs_service)
- [Resource: aws_efs_file_system](#resource-aws_efs_file_system)
- [Resource: aws_elasticache_cluster](#resource-aws_elasticache_cluster)
- [Resource: aws_iam_user_login_profile](#resource-aws_iam_user_login_profile)
- [Resource: aws_instance](#resource-aws_instance)
- [Resource: aws_lambda_function](#resource-aws_lambda_function)
- [Resource: aws_lambda_layer_version](#resource-aws_lambda_layer_version)
- [Resource: aws_network_acl](#resource-aws_network_acl)
- [Resource: aws_redshift_cluster](#resource-aws_redshift_cluster)
- [Resource: aws_route_table](#resource-aws_route_table)
- [Resource: aws_route53_record](#resource-aws_route53_record)
- [Resource: aws_route53_zone](#resource-aws_route53_zone)
- [Resource: aws_wafregional_byte_match_set](#resource-aws_wafregional_byte_match_set)

<!-- /TOC -->

## Provider Version Configuration

-> Before upgrading to version 2.0.0 or later, it is recommended to upgrade to the most recent 1.X version of the provider (version 1.60.0) and ensure that your environment successfully runs [`terraform plan`](https://www.terraform.io/docs/commands/plan.html) without unexpected changes or deprecation notices.

It is recommended to use [version constraints when configuring Terraform providers](https://www.terraform.io/docs/configuration/providers.html#provider-versions). If you are following that recommendation, update the version constraints in your Terraform configuration and run [`terraform init`](https://www.terraform.io/docs/commands/init.html) to download the new version.

Update to latest 1.X version:

```hcl
provider "aws" {
  # ... other configuration ...

  version = "~> 1.60"
}
```

Update to latest 2.X version:

```hcl
provider "aws" {
  # ... other configuration ...

  version = "~> 2.0"
}
```

## Provider: Configuration

### skip_requesting_account_id Argument Now Required to Skip Account ID Lookup Errors

If the provider is unable to determine the AWS account ID from a provider assume role configuration or the STS GetCallerIdentity call used to verify the credentials (if `skip_credentials_validation = false`), it will attempt to lookup the AWS account ID via EC2 metadata, IAM GetUser, IAM ListRoles, and STS GetCallerIdentity. Previously, the provider would silently allow the failure of all the above methods.

The provider will now return an error to ensure operators understand the implications of the missing AWS account ID in the provider.

If necessary, the AWS account ID lookup logic can be skipped via:

```hcl
provider "aws" {
  # ... other configuration ...

  skip_requesting_account_id = true
}
```

## Data Source: aws_ami

### owners Argument Now Required

The `owners` argument is now required. Specifying `owner-id` or `owner-alias` under `filter` does not satisfy this requirement.

## Data Source: aws_ami_ids

### owners Argument Now Required

The `owners` argument is now required. Specifying `owner-id` or `owner-alias` under `filter` does not satisfy this requirement.

## Data Source: aws_iam_role

### assume_role_policy_document Attribute Removal

Switch your attribute references to the `assume_role_policy` attribute instead.

### role_id Attribute Removal

Switch your attribute references to the `unique_id` attribute instead.

### role_name Argument Removal

Switch your Terraform configuration to the `name` argument instead.

## Data Source: aws_kms_secret

### Data Source Removal and Migrating to aws_kms_secrets Data Source

The implementation of the `aws_kms_secret` data source, prior to Terraform AWS provider version 2.0.0, used dynamic attribute behavior which is not supported with Terraform 0.12 and beyond (full details available in [this GitHub issue](https://github.com/hashicorp/terraform-provider-aws/issues/5144)).

Terraform configuration migration steps:

* Change the data source type from `aws_kms_secret` to `aws_kms_secrets`
* Change any attribute reference (e.g. `"${data.aws_kms_secret.example.ATTRIBUTE}"`) from `.ATTRIBUTE` to `.plaintext["ATTRIBUTE"]`

As an example, lets take the below sample configuration and migrate it.

```hcl
# Below example configuration will not be supported in Terraform AWS provider version 2.0.0

data "aws_kms_secret" "example" {
  secret {
    # ... potentially other configuration ...
    name    = "master_password"
    payload = "AQEC..."
  }

  secret {
    # ... potentially other configuration ...
    name    = "master_username"
    payload = "AQEC..."
  }
}

resource "aws_rds_cluster" "example" {
  # ... other configuration ...
  master_password = "${data.aws_kms_secret.example.master_password}"
  master_username = "${data.aws_kms_secret.example.master_username}"
}
```

Notice that the `aws_kms_secret` data source previously was taking the two `secret` configuration block `name` arguments and generating those as attribute names (`master_password` and `master_username` in this case). To remove the incompatible behavior, this updated version of the data source provides the decrypted value of each of those `secret` configuration block `name` arguments within a map attribute named `plaintext`.

Updating the sample configuration from above:

```hcl
data "aws_kms_secrets" "example" {
  secret {
    # ... potentially other configuration ...
    name    = "master_password"
    payload = "AQEC..."
  }

  secret {
    # ... potentially other configuration ...
    name    = "master_username"
    payload = "AQEC..."
  }
}

resource "aws_rds_cluster" "example" {
  # ... other configuration ...
  master_password = "${data.aws_kms_secrets.example.plaintext["master_password"]}"
  master_username = "${data.aws_kms_secrets.example.plaintext["master_username"]}"
}
```

## Data Source: aws_lambda_function

### arn and qualified_arn Attribute Behavior Changes

The `arn` attribute now always returns the unqualified (no `:QUALIFIER` or `:VERSION` suffix) Amazon Resource Name (ARN) value and the `qualified_arn` attribute now always returns the qualified (includes `:QUALIFIER` or `:VERSION` suffix) ARN value. Previously by default, the `arn` attribute included `:$LATEST` suffix when not setting the optional `qualifier` argument, which was not compatible with many other resources. To restore the previous default behavior, set the `qualifier` argument to `$LATEST` and reference the `qualified_arn` attribute.

## Data Source: aws_region

### current Argument Removal

Simply remove `current = true` from your Terraform configuration. The data source defaults to the current provider region if no other filtering is enabled.

## Resource: aws_api_gateway_api_key

### stage_key Argument Removal

Since the API Gateway usage plans feature was launched on August 11, 2016, usage plans are now required to associate an API key with an API stage. To migrate your Terraform configuration, the AWS provider implements support for usage plans with the following resources:

* [`aws_api_gateway_usage_plan`](/docs/providers/aws/r/api_gateway_usage_plan.html)
* [`aws_api_gateway_usage_plan_key`](/docs/providers/aws/r/api_gateway_usage_plan_key.html)

For example, given this previous configuration:

```hcl
resource "aws_api_gateway_rest_api" "example" {
  name = "example"
}

resource "aws_api_gateway_deployment" "example" {
  rest_api_id = "${aws_api_gateway_rest_api.example.id}"
  stage_name  = "example"
}

resource "aws_api_gateway_api_key" "example" {
  name = "example"

  stage_key {
    rest_api_id = "${aws_api_gateway_rest_api.example.id}"
    stage_name  = "${aws_api_gateway_deployment.example.stage_name}"
  }
}
```

An updated configuration:

```hcl
resource "aws_api_gateway_rest_api" "example" {
  name = "example"
}

resource "aws_api_gateway_deployment" "example" {
  rest_api_id = "${aws_api_gateway_rest_api.example.id}"
  stage_name  = "example"
}

resource "aws_api_gateway_api_key" "example" {
  name = "example"
}

resource "aws_api_gateway_usage_plan" "example" {
  name = "example"

  api_stages {
    api_id = "${aws_api_gateway_rest_api.example.id}"
    stage  = "${aws_api_gateway_deployment.example.stage_name}"
  }
}

resource "aws_api_gateway_usage_plan_key" "example" {
  key_id        = "${aws_api_gateway_api_key.example.id}"
  key_type      = "API_KEY"
  usage_plan_id = "${aws_api_gateway_usage_plan.example.id}"
}
```

## Resource: aws_api_gateway_integration

### request_parameters_in_json Argument Removal

Switch your Terraform configuration to the `request_parameters` argument instead.

For example, given this previous configuration:

```hcl
resource "aws_api_gateway_integration" "example" {
  # ... other configuration ...

  request_parameters_in_json = <<PARAMS
{
    "integration.request.header.X-Authorization": "'static'"
}
PARAMS
}
```

An updated configuration:

```hcl
resource "aws_api_gateway_integration" "example" {
  # ... other configuration ...

  request_parameters = {
    "integration.request.header.X-Authorization" = "'static'"
  }
}
```

## Resource: aws_api_gateway_integration_response

### response_parameters_in_json Argument Removal

Switch your Terraform configuration to the `response_parameters` argument instead.

For example, given this previous configuration:

```hcl
resource "aws_api_gateway_integration_response" "example" {
  # ... other configuration ...

  response_parameters_in_json = <<PARAMS
{
    "method.response.header.Content-Type": "integration.response.body.type"
}
PARAMS
}
```

An updated configuration:

```hcl
resource "aws_api_gateway_integration_response" "example" {
  # ... other configuration ...

  response_parameters = {
    "method.response.header.Content-Type" = "integration.response.body.type"
  }
}
```

## Resource: aws_api_gateway_method

### request_parameters_in_json Argument Removal

Switch your Terraform configuration to the `request_parameters` argument instead.

For example, given this previous configuration:

```hcl
resource "aws_api_gateway_method" "example" {
  # ... other configuration ...

  request_parameters_in_json = <<PARAMS
{
    "method.request.header.Content-Type": false,
    "method.request.querystring.page": true
}
PARAMS
}
```

An updated configuration:

```hcl
resource "aws_api_gateway_method" "example" {
  # ... other configuration ...

  request_parameters = {
    "method.request.header.Content-Type" = false
    "method.request.querystring.page"    = true
  }
}
```

## Resource: aws_api_gateway_method_response

### response_parameters_in_json Argument Removal

Switch your Terraform configuration to the `response_parameters` argument instead.

For example, given this previous configuration:

```hcl
resource "aws_api_gateway_method_response" "example" {
  # ... other configuration ...

  response_parameters_in_json = <<PARAMS
{
    "method.response.header.Content-Type": true
}
PARAMS
}
```

An updated configuration:

```hcl
resource "aws_api_gateway_method_response" "example" {
  # ... other configuration ...

  response_parameters = {
    "method.response.header.Content-Type" = true
  }
}
```

## Resource: aws_appautoscaling_policy

### Argument Removals

The following arguments have been moved into a nested argument named `step_scaling_policy_configuration`:

* `adjustment_type`
* `cooldown`
* `metric_aggregation_type`
* `min_adjustment_magnitude`
* `step_adjustment`

For example, given this previous configuration:

```hcl
resource "aws_appautoscaling_policy" "example" {
  # ... other configuration ...

  adjustment_type         = "ChangeInCapacity"
  cooldown                = 60
  metric_aggregation_type = "Maximum"

  step_adjustment {
    metric_interval_upper_bound = 0
    scaling_adjustment          = -1
  }
}
```

An updated configuration:

```hcl
resource "aws_appautoscaling_policy" "example" {
  # ... other configuration ...

  step_scaling_policy_configuration {
    adjustment_type         = "ChangeInCapacity"
    cooldown                = 60
    metric_aggregation_type = "Maximum"

    step_adjustment {
      metric_interval_upper_bound = 0
      scaling_adjustment          = -1
    }
  }
}
```

## Resource: aws_autoscaling_policy

### min_adjustment_step Argument Removal

Switch your Terraform configuration to the `min_adjustment_magnitude` argument instead.

For example, given this previous configuration:

```hcl
resource "aws_autoscaling_policy" "example" {
  # ... other configuration ...

  min_adjustment_step = 2
}
```

An updated configuration:

```hcl
resource "aws_autoscaling_policy" "example" {
  # ... other configuration ...

  min_adjustment_magnitude = 2
}
```

## Resource: aws_batch_compute_environment

### ecc_cluster_arn Attribute Removal

Switch your attribute references to the `ecs_cluster_arn` attribute instead.

## Resource: aws_cloudfront_distribution

### cache_behavior Argument Removal

Switch your Terraform configuration to the `ordered_cache_behavior` argument instead. It behaves similar to the previous `cache_behavior` argument, however the ordering of the configurations in Terraform is now reflected in the distribution where previously it was indeterminate.

For example, given this previous configuration:

```hcl
resource "aws_cloudfront_distribution" "example" {
  # ... other configuration ...

  cache_behavior {
    # ... other configuration ...
  }

  cache_behavior {
    # ... other configuration ...
  }
}
```

An updated configuration:

```hcl
resource "aws_cloudfront_distribution" "example" {
  # ... other configuration ...

  ordered_cache_behavior {
    # ... other configuration ...
  }

  ordered_cache_behavior {
    # ... other configuration ...
  }
}
```

## Resource: aws_cognito_user_pool

### email_verification_subject Argument Now Conflicts With verification_message_template Configuration Block email_subject Argument

Choose one argument or the other. These arguments update the same underlying information in Cognito and the selection is indeterminate if differing values are provided.

### email_verification_message Argument Now Conflicts With verification_message_template Configuration Block email_message Argument

Choose one argument or the other. These arguments update the same underlying information in Cognito and the selection is indeterminate if differing values are provided.

### sms_verification_message Argument Now Conflicts With verification_message_template Configuration Block sms_message Argument

Choose one argument or the other. These arguments update the same underlying information in Cognito and the selection is indeterminate if differing values are provided.

## Resource: aws_dx_lag

### number_of_connections Argument Removal

Default connections have been removed as part of LAG creation. To migrate your Terraform configuration, the AWS provider implements the following resources:

* [`aws_dx_connection`](/docs/providers/aws/r/dx_connection.html)
* [`aws_dx_connection_association`](/docs/providers/aws/r/dx_connection_association.html)

For example, given this previous configuration:

```hcl
resource "aws_dx_lag" "example" {
  name                  = "example"
  connections_bandwidth = "1Gbps"
  location              = "EqSe2-EQ"
  number_of_connections = 1
}
```

An updated configuration:

```hcl
resource "aws_dx_connection" "example" {
  name      = "example"
  bandwidth = "1Gbps"
  location  = "EqSe2-EQ"
}

resource "aws_dx_lag" "example" {
  name                  = "example"
  connections_bandwidth = "1Gbps"
  location              = "EqSe2-EQ"
}

resource "aws_dx_connection_association" "example" {
  connection_id = "${aws_dx_connection.example.id}"
  lag_id        = "${aws_dx_lag.example.id}"
}
```

## Resource: aws_ecs_service

### placement_strategy Argument Removal

Switch your Terraform configuration to the `ordered_placement_strategy` argument instead. It behaves similar to the previous `placement_strategy` argument, however the ordering of the configurations in Terraform is now reflected in the distribution where previously it was indeterminate.

For example, given this previous configuration:

```hcl
resource "aws_ecs_service" "example" {
  # ... other configuration ...

  placement_strategy {
    # ... other configuration ...
  }

  placement_strategy {
    # ... other configuration ...
  }
}
```

An updated configuration:

```hcl
resource "aws_ecs_service" "example" {
  # ... other configuration ...

  ordered_placement_strategy {
    # ... other configuration ...
  }

  ordered_placement_strategy {
    # ... other configuration ...
  }
}
```

## Resource: aws_efs_file_system

### reference_name Argument Removal

Switch your Terraform configuration to the `creation_token` argument instead.

For example, given this previous configuration:

```hcl
resource "aws_efs_file_system" "example" {
  # ... other configuration ...

  reference_name = "example"
}
```

An updated configuration:

```hcl
resource "aws_efs_file_system" "example" {
  # ... other configuration ...

  creation_token = "example"
}
```

## Resource: aws_elasticache_cluster

### availability_zones Argument Removal

Switch your Terraform configuration to the `preferred_availability_zones` argument instead. The argument is still optional and the API will continue to automatically choose Availability Zones for nodes if not specified. The new argument will also continue to match the APIs required behavior that the length of the list must be the same as `num_cache_nodes`.

For example, given this previous configuration:

```hcl
resource "aws_elasticache_cluster" "example" {
  # ... other configuration ...

  availability_zones = ["us-west-2a", "us-west-2b"]
}
```

An updated configuration:

```hcl
resource "aws_elasticache_cluster" "example" {
  # ... other configuration ...

  preferred_availability_zones = ["us-west-2a", "us-west-2b"]
}
```

## Resource: aws_iam_user_login_profile

### Import Now Required For Existing Infrastructure

When attempting to bring existing IAM User Login Profiles under Terraform management, `terraform import` is now required. See the [`aws_iam_user_login_profile` resource documentation](https://www.terraform.io/docs/providers/aws/r/iam_user_login_profile.html) for more information.

## Resource: aws_instance

### network_interface_id Attribute Removal

Switch your attribute references to the `primary_network_interface_id` attribute instead.

## Resource: aws_lambda_function

### reserved_concurrent_executions Argument Behavior Change

Setting `reserved_concurrent_executions` to `0` will now disable Lambda Function invocations, causing downtime for the Lambda Function.

Previously `reserved_concurrent_executions` accepted `0` and below for unreserved concurrency, which means it was not previously possible to disable invocations. The argument now differentiates between a new value for unreserved concurrency (`-1`) and disabling Lambda invocations (`0`). If previously configuring this value to `0` for unreserved concurrency, update the configured value to `-1` or the resource will disable Lambda Function invocations on update. If previously unconfigured, the argument does not require any changes.

See the [Lambda User Guide](https://docs.aws.amazon.com/lambda/latest/dg/concurrent-executions.html) for more information about concurrency.

## Resource: aws_lambda_layer_version

### arn and layer_arn Attribute Value Swap

Switch your `arn` attribute references to the `layer_arn` attribute instead and vice-versa.

## Resource: aws_network_acl

### subnet_id Argument Removal

Switch your Terraform configuration to the `subnet_ids` argument instead.

For example, given this previous configuration:

```hcl
resource "aws_network_acl" "example" {
  # ... other configuration ...

  subnet_id = "subnet-12345678"
}
```

An updated configuration:

```hcl
resource "aws_network_acl" "example" {
  # ... other configuration ...

  subnet_ids = ["subnet-12345678"]
}
```

## Resource: aws_redshift_cluster

### Argument Removals

The following arguments have been moved into a nested argument named `logging`:

* `bucket_name`
* `enable_logging` (also renamed to just `enable`)
* `s3_key_prefix`

For example, given this previous configuration:

```hcl
resource "aws_redshift_cluster" "example" {
  # ... other configuration ...

  bucket_name    = "example"
  enable_logging = true
  s3_key_prefix  = "example"
}
```

An updated configuration:

```hcl
resource "aws_redshift_cluster" "example" {
  # ... other configuration ...

  logging {
    bucket_name   = "example"
    enable        = true
    s3_key_prefix = "example"
  }
}
```

## Resource: aws_route_table

### Import Change

Previously, importing this resource resulted in an `aws_route` resource for each route, in
addition to the `aws_route_table`, in the Terraform state. Support for importing `aws_route` resources has been added and importing this resource only adds the `aws_route_table`
resource, with in-line routes, to the state.

## Resource: aws_route53_record

### allow_overwrite Default Value Change

The resource now requires existing Route 53 Records to be imported into the Terraform state for management unless the `allow_overwrite` argument is enabled.

For example, if the `www.example.com` Route 53 Record in the `example.com` Route 53 Hosted Zone existed previously and this new Terraform configuration was introduced:

```hcl
resource "aws_route53_record" "www" {
  # ... other configuration ...
  name = "www.example.com"
}
```

During resource creation in version 1.X and prior, it would silently perform an `UPSERT` changeset to the existing Route 53 Record and not report back an error. In version 2.0.0 of the Terraform AWS Provider, the resource now performs a `CREATE` changeset, which will error for existing Route 53 Records.

The `allow_overwrite` argument provides a workaround to keep the old behavior, but most existing workflows should be updated to perform a `terraform import` command like the following instead:

```console
$ terraform import aws_route53_record.www ZONEID_www.example.com_TYPE
```

More information can be found in the [`aws_route53_record` resource documentation](https://www.terraform.io/docs/providers/aws/r/route53_record.html#import).

## Resource: aws_route53_zone

### vpc_id and vpc_region Argument Removal

Switch your Terraform configuration to `vpc` configuration block(s) instead.

For example, given this previous configuration:

```hcl
resource "aws_route53_zone" "example" {
  # ... other configuration ...

  vpc_id = "..."
}
```

An updated configuration:

```hcl
resource "aws_route53_zone" "example" {
  # ... other configuration ...

  vpc {
    vpc_id = "..."
  }
}
```

## Resource: aws_wafregional_byte_match_set

### byte_match_tuple Argument Removal

Switch your Terraform configuration to the `byte_match_tuples` argument instead.

For example, given this previous configuration:

```hcl
resource "aws_wafregional_byte_match_set" "example" {
  # ... other configuration ...

  byte_match_tuple {
    # ... other configuration ...
  }

  byte_match_tuple {
    # ... other configuration ...
  }
}
```

An updated configuration:

```hcl
resource "aws_wafregional_byte_match_set" "example" {
  # ... other configuration ...

  byte_match_tuples {
    # ... other configuration ...
  }

  byte_match_tuples {
    # ... other configuration ...
  }
}
```
