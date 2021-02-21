---
subcategory: "Config"
layout: "aws"
page_title: "AWS: aws_config_remediation_configuration"
description: |-
  Provides an AWS Config Remediation Configuration.
---

# Resource: aws_config_remediation_configuration

Provides an AWS Config Remediation Configuration.

~> **Note:** Config Remediation Configuration requires an existing [Config Rule](/docs/providers/aws/r/config_config_rule.html) to be present.

## Example Usage

AWS managed rules can be used by setting the source owner to `AWS` and the source identifier to the name of the managed rule. More information about AWS managed rules can be found in the [AWS Config Developer Guide](https://docs.aws.amazon.com/config/latest/developerguide/evaluate-config_use-managed-rules.html).

```hcl
resource "aws_config_config_rule" "this" {
  name = "example"

  source {
    owner             = "AWS"
    source_identifier = "S3_BUCKET_VERSIONING_ENABLED"
  }
}

resource "aws_config_remediation_configuration" "this" {
  config_rule_name = aws_config_config_rule.this.name
  resource_type    = "AWS::S3::Bucket"
  target_type      = "SSM_DOCUMENT"
  target_id        = "AWS-EnableS3BucketEncryption"
  target_version   = "1"

  parameter {
    name         = "AutomationAssumeRole"
    static_value = "arn:aws:iam::875924563244:role/security_config"
  }
  parameter {
    name           = "BucketName"
    resource_value = "RESOURCE_ID"
  }
  parameter {
    name         = "SSEAlgorithm"
    static_value = "AES256"
  }
}
```

## Argument Reference

The following arguments are supported:

* `config_rule_name` - (Required) The name of the AWS Config rule
* `resource_type` - (Optional) The type of a resource
* `target_id` - (Required) Target ID is the name of the public document
* `target_type` - (Required) The type of the target. Target executes remediation. For example, SSM document
* `target_version` - (Optional) Version of the target. For example, version of the SSM document
* `parameter` - (Optional) Can be specified multiple times for each
   parameter. Each parameter block supports fields documented below.

The `parameter` block supports:

The value is either a dynamic (resource) value or a static value.
You must select either a dynamic value or a static value.

* `name` - (Required) The name of the attribute.
* `resource_value` - (Optional) The value is dynamic and changes at run-time.
* `static_value` - (Optional) The value is static and does not change at run-time.

## Import

Remediation Configurations can be imported using the name config_rule_name, e.g.

```
$ terraform import aws_config_remediation_configuration.this example
```
