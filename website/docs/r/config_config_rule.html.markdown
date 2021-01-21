---
subcategory: "Config"
layout: "aws"
page_title: "AWS: aws_config_config_rule"
description: |-
  Provides an AWS Config Rule.
---

# Resource: aws_config_config_rule

Provides an AWS Config Rule.

~> **Note:** Config Rule requires an existing [Configuration Recorder](/docs/providers/aws/r/config_configuration_recorder.html) to be present. Use of `depends_on` is recommended (as shown below) to avoid race conditions.

## Example Usage

### AWS Managed Rules

AWS managed rules can be used by setting the source owner to `AWS` and the source identifier to the name of the managed rule. More information about AWS managed rules can be found in the [AWS Config Developer Guide](https://docs.aws.amazon.com/config/latest/developerguide/evaluate-config_use-managed-rules.html).

```hcl
resource "aws_config_config_rule" "r" {
  name = "example"

  source {
    owner             = "AWS"
    source_identifier = "S3_BUCKET_VERSIONING_ENABLED"
  }

  depends_on = [aws_config_configuration_recorder.foo]
}

resource "aws_config_configuration_recorder" "foo" {
  name     = "example"
  role_arn = aws_iam_role.r.arn
}

resource "aws_iam_role" "r" {
  name = "my-awsconfig-role"

  assume_role_policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "config.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
POLICY
}

resource "aws_iam_role_policy" "p" {
  name = "my-awsconfig-policy"
  role = aws_iam_role.r.id

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
  	{
  		"Action": "config:Put*",
  		"Effect": "Allow",
  		"Resource": "*"

  	}
  ]
}
POLICY
}
```

### Custom Rules

Custom rules can be used by setting the source owner to `CUSTOM_LAMBDA` and the source identifier to the Amazon Resource Name (ARN) of the Lambda Function. The AWS Config service must have permissions to invoke the Lambda Function, e.g. via the [`aws_lambda_permission` resource](/docs/providers/aws/r/lambda_permission.html). More information about custom rules can be found in the [AWS Config Developer Guide](https://docs.aws.amazon.com/config/latest/developerguide/evaluate-config_develop-rules.html).

```hcl
resource "aws_config_configuration_recorder" "example" {
  # ... other configuration ...
}

resource "aws_lambda_function" "example" {
  # ... other configuration ...
}

resource "aws_lambda_permission" "example" {
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.example.arn
  principal     = "config.amazonaws.com"
  statement_id  = "AllowExecutionFromConfig"
}

resource "aws_config_config_rule" "example" {
  # ... other configuration ...

  source {
    owner             = "CUSTOM_LAMBDA"
    source_identifier = aws_lambda_function.example.arn
  }

  depends_on = [
    aws_config_configuration_recorder.example,
    aws_lambda_permission.example,
  ]
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the rule
* `description` - (Optional) Description of the rule
* `input_parameters` - (Optional) A string in JSON format that is passed to the AWS Config rule Lambda function.
* `maximum_execution_frequency` - (Optional) The maximum frequency with which AWS Config runs evaluations for a rule.
* `scope` - (Optional) Scope defines which resources can trigger an evaluation for the rule as documented below.
* `source` - (Required) Source specifies the rule owner, the rule identifier, and the notifications that cause the function to evaluate your AWS resources as documented below.
* `tags` - (Optional) A map of tags to assign to the resource.

### `scope`

Defines which resources can trigger an evaluation for the rule.
If you do not specify a scope, evaluations are triggered when any resource in the recording group changes.

* `compliance_resource_id` - (Optional) The IDs of the only AWS resource that you want to trigger an evaluation for the rule. If you specify a resource ID, you must specify one resource type for `compliance_resource_types`.
* `compliance_resource_types` - (Optional) A list of resource types of only those AWS resources that you want to trigger an evaluation for the rule. e.g. `AWS::EC2::Instance`. You can only specify one type if you also specify a resource ID for `compliance_resource_id`. See [relevant part of AWS Docs](http://docs.aws.amazon.com/config/latest/APIReference/API_ResourceIdentifier.html#config-Type-ResourceIdentifier-resourceType) for available types.
* `tag_key` - (Optional, Required if `tag_value` is specified) The tag key that is applied to only those AWS resources that you want you want to trigger an evaluation for the rule.
* `tag_value` - (Optional) The tag value applied to only those AWS resources that you want to trigger an evaluation for the rule.

### `source`

Provides the rule owner (AWS or customer), the rule identifier, and the notifications that cause the function to evaluate your AWS resources.

* `owner` - (Required) Indicates whether AWS or the customer owns and manages the AWS Config rule. Valid values are `AWS` or `CUSTOM_LAMBDA`. For more information about managed rules, see the [AWS Config Managed Rules documentation](https://docs.aws.amazon.com/config/latest/developerguide/evaluate-config_use-managed-rules.html). For more information about custom rules, see the [AWS Config Custom Rules documentation](https://docs.aws.amazon.com/config/latest/developerguide/evaluate-config_develop-rules.html). Custom Lambda Functions require permissions to allow the AWS Config service to invoke them, e.g. via the [`aws_lambda_permission` resource](/docs/providers/aws/r/lambda_permission.html).
* `source_identifier` - (Required) For AWS Config managed rules, a predefined identifier, e.g `IAM_PASSWORD_POLICY`. For custom Lambda rules, the identifier is the ARN of the Lambda Function, such as `arn:aws:lambda:us-east-1:123456789012:function:custom_rule_name` or the [`arn` attribute of the `aws_lambda_function` resource](/docs/providers/aws/r/lambda_function.html#arn).
* `source_detail` - (Optional) Provides the source and type of the event that causes AWS Config to evaluate your AWS resources. Only valid if `owner` is `CUSTOM_LAMBDA`.
    * `event_source` - (Optional) The source of the event, such as an AWS service, that triggers AWS Config to evaluate your AWS resources. This defaults to `aws.config` and is the only valid value.
    * `maximum_execution_frequency` - (Optional) The frequency that you want AWS Config to run evaluations for a rule that is triggered periodically. If specified, requires `message_type` to be `ScheduledNotification`.
    * `message_type` - (Optional) The type of notification that triggers AWS Config to run an evaluation for a rule. You can specify the following notification types:
        * `ConfigurationItemChangeNotification` - Triggers an evaluation when AWS Config delivers a configuration item as a result of a resource change.
        * `OversizedConfigurationItemChangeNotification` - Triggers an evaluation when AWS Config delivers an oversized configuration item. AWS Config may generate this notification type when a resource changes and the notification exceeds the maximum size allowed by Amazon SNS.
        * `ScheduledNotification` - Triggers a periodic evaluation at the frequency specified for `maximum_execution_frequency`.
        * `ConfigurationSnapshotDeliveryCompleted` - Triggers a periodic evaluation when AWS Config delivers a configuration snapshot.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The ARN of the config rule
* `rule_id` - The ID of the config rule

## Import

Config Rule can be imported using the name, e.g.

```
$ terraform import aws_config_config_rule.foo example
```
