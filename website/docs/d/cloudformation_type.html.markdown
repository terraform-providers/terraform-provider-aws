---
subcategory: "CloudFormation"
layout: "aws"
page_title: "AWS: aws_cloudformation_type"
description: |-
    Provides details about a CloudFormation Type.
---

# Data Source: aws_cloudformation_type

Provides details about a CloudFormation Type.

## Example Usage

```terraform
data "aws_cloudformation_type" "example" {
  type      = "RESOURCE"
  type_name = "AWS::Athena::WorkGroup"
}
```

## Argument Reference

The following arguments are supported:

* `arn` - (Optional) Amazon Resource Name (ARN) of the CloudFormation Type. For example, `arn:aws:cloudformation:us-west-2::type/resource/AWS-EC2-VPC`.
* `type` - (Optional) CloudFormation Registry Type. For example, `RESOURCE`.
* `type_name` - (Optional) CloudFormation Type name. For example, `AWS::EC2::VPC`.
* `version_id` - (Optional) Identifier of the CloudFormation Type version.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `default_version_id` - Identifier of the CloudFormation Type default version.
* `deprecated_status` - Deprecation status of the CloudFormation Type.
* `description` - Description of the CloudFormation Type.
* `documentation_url` - URL of the documentation for the CloudFormation Type.
* `execution_role_arn` - Amazon Resource Name (ARN) of the IAM Role used to register the CloudFormation Type.
* `is_default_version` - Whether the CloudFormation Type version is the default version.
* `logging_config` - List of objects containing logging configuration.
    * `log_group_name` - Name of the CloudWatch Log Group where CloudFormation sends error logging information when invoking the type's handlers.
    * `log_role_arn` - Amazon Resource Name (ARN) of the IAM Role CloudFormation assumes when sending error logging information to CloudWatch Logs.
* `provisioning_type` - Provisioning behavior of the CloudFormation Type.
* `schema` - JSON document of the CloudFormation Type schema.
* `source_url` - URL of the source code for the CloudFormation Type.
* `visibility` - Scope of the CloudFormation Type.
