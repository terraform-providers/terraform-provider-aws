---
subcategory: "Config"
layout: "aws"
page_title: "AWS: aws_config_organization_custom_rule"
description: |-
  Manages a Config Organization Custom Rule
---

# Resource: aws_config_organization_custom_rule

Manages a Config Organization Custom Rule. More information about these rules can be found in the [Enabling AWS Config Rules Across all Accounts in Your Organization](https://docs.aws.amazon.com/config/latest/developerguide/config-rule-multi-account-deployment.html) and [AWS Config Managed Rules](https://docs.aws.amazon.com/config/latest/developerguide/evaluate-config_use-managed-rules.html) documentation. For working with Organization Managed Rules (those invoking an AWS managed rule), see the [`aws_config_organization_managed__rule` resource](/docs/providers/aws/r/config_organization_managed_rule.html).

~> **NOTE:** This resource must be created in the Organization master account and rules will include the master account unless its ID is added to the `excluded_accounts` argument.

~> **NOTE:** The proper Lambda permission to allow the AWS Config service invoke the Lambda Function must be in place before the rule will successfully create or update. See also the [`aws_lambda_permission` resource](/docs/providers/aws/r/lambda_permission.html).

## Example Usage

```hcl
resource "aws_lambda_permission" "example" {
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.example.arn
  principal     = "config.amazonaws.com"
  statement_id  = "AllowExecutionFromConfig"
}

resource "aws_organizations_organization" "example" {
  aws_service_access_principals = ["config-multiaccountsetup.amazonaws.com"]
  feature_set                   = "ALL"
}

resource "aws_config_organization_custom_rule" "example" {
  depends_on = [
    aws_lambda_permission.example,
    aws_organizations_organization.example,
  ]

  lambda_function_arn = aws_lambda_function.example.arn
  name                = "example"
  trigger_types       = ["ConfigurationItemChangeNotification"]
}
```

## Argument Reference

The following arguments are supported:

* `lambda_function_arn` - (Required) Amazon Resource Name (ARN) of the rule Lambda Function
* `name` - (Required) The name of the rule
* `trigger_types` - (Required) List of notification types that trigger AWS Config to run an evaluation for the rule. Valid values: `ConfigurationItemChangeNotification`, `OversizedConfigurationItemChangeNotification`, and `ScheduledNotification`
* `description` - (Optional) Description of the rule
* `excluded_accounts` - (Optional) List of AWS account identifiers to exclude from the rule
* `input_parameters` - (Optional) A string in JSON format that is passed to the AWS Config Rule Lambda Function
* `maximum_execution_frequency` - (Optional) The maximum frequency with which AWS Config runs evaluations for a rule, if the rule is triggered at a periodic frequency. Defaults to `TwentyFour_Hours` for periodic frequency triggered rules. Valid values: `One_Hour`, `Three_Hours`, `Six_Hours`, `Twelve_Hours`, or `TwentyFour_Hours`.
* `resource_id_scope` - (Optional) Identifier of the AWS resource to evaluate
* `resource_types_scope` - (Optional) List of types of AWS resources to evaluate
* `tag_key_scope` - (Optional, Required if `tag_value_scope` is configured) Tag key of AWS resources to evaluate
* `tag_value_scope` - (Optional) Tag value of AWS resources to evaluate


## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - Amazon Resource Name (ARN) of the rule

## Timeouts

`aws_config_organization_custom_rule` provides the following [Timeouts](https://www.terraform.io/docs/configuration/blocks/resources/syntax.html#operation-timeouts)
configuration options:

* `create` - (Default `5m`) How long to wait for the rule to be created.
* `delete` - (Default `5m`) How long to wait for the rule to be deleted.
* `update` - (Default `5m`) How long to wait for the rule to be updated.

## Import

Config Organization Custom Rules can be imported using the name, e.g.

```
$ terraform import aws_config_organization_custom_rule.example example
```
