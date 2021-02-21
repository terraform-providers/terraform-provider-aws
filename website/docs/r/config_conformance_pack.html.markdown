---
subcategory: "Config"
layout: "aws"
page_title: "AWS: aws_config_conformance_pack"
description: |-
  Manages a Config Conformance Pack
---

# Resource: aws_config_conformance_pack

Manages a Config Conformance Pack. More information about this collection of Config rules and remediation actions can be found in the
[Conformance Packs](https://docs.aws.amazon.com/config/latest/developerguide/conformance-packs.html) documentation.
Sample Conformance Pack templates may be found in the
[AWS Config Rules Repository](https://github.com/awslabs/aws-config-rules/tree/master/aws-config-conformance-packs).

~> **NOTE:** The account must have a Configuration Recorder with proper IAM permissions before the Conformance Pack will
successfully create or update. See also the
[`aws_config_configuration_recorder` resource](/docs/providers/aws/r/config_configuration_recorder.html).

## Example Usage with Template Body

```hcl
resource "aws_config_conformance_pack" "example" {
  name = "example"

  input_parameter {
    parameter_name  = "AccessKeysRotatedParameterMaxAccessKeyAge"
    parameter_value = "90"
  }

  template_body = <<EOT
Parameters:
  AccessKeysRotatedParameterMaxAccessKeyAge:
    Type: String
Resources:
  IAMPasswordPolicy:
    Properties:
      ConfigRuleName: IAMPasswordPolicy
      Source:
        Owner: AWS
        SourceIdentifier: IAM_PASSWORD_POLICY
    Type: AWS::Config::ConfigRule
EOT

  depends_on = [aws_config_configuration_recorder.example]
}
```

## Example Usage with Template S3 URI

```hcl
resource "aws_config_conformance_pack" "example" {
  name            = "example"
  template_s3_uri = "s3://${aws_s3_bucket.example.bucket}/${aws_s3_bucket_object.example.key}"

  depends_on = [aws_config_configuration_recorder.example]
}

resource "aws_s3_bucket" "example" {
  bucket = "example"
}

resource "aws_s3_bucket_object" "example" {
  bucket  = aws_s3_bucket.example.id
  key     = "example-key"
  content = <<EOT
Resources:
  IAMPasswordPolicy:
    Properties:
      ConfigRuleName: IAMPasswordPolicy
      Source:
        Owner: AWS
        SourceIdentifier: IAM_PASSWORD_POLICY
    Type: AWS::Config::ConfigRule
EOT
}
```

## Argument Reference

~> **Note:** If both `template_body` and `template_s3_uri` are specified, AWS Config uses the `template_s3_uri` and ignores the `template_body`.

The following arguments are supported:

* `name` - (Required, Forces new resource) The name of the conformance pack. Must begin with a letter and contain from 1 to 256 alphanumeric characters and hyphens.
* `delivery_s3_bucket` - (Optional) Amazon S3 bucket where AWS Config stores conformance pack templates. Maximum length of 63.
* `delivery_s3_key_prefix` - (Optional) The prefix for the Amazon S3 bucket. Maximum length of 1024.
* `input_parameter` - (Optional) Set of configuration blocks describing input parameters passed to the conformance pack template. Documented below. When configured, the parameters must also be included in the `template_body` or in the template stored in Amazon S3 if using `template_s3_uri`.
* `template_body` - (Optional, required if `template_s3_uri` is not provided) A string containing full conformance pack template body. Maximum length of 51200. Drift detection is not possible with this argument.
* `template_s3_uri` - (Optional, required if `template_body` is not provided) Location of file, e.g. `s3://bucketname/prefix`, containing the template body. The uri must point to the conformance pack template that is located in an Amazon S3 bucket in the same region as the conformance pack. Maximum length of 1024. Drift detection is not possible with this argument.

### input_parameter Argument Reference

The `input_parameter` configuration block supports the following arguments:

* `parameter_name` - (Required) The input key.
* `parameter_value` - (Required) The input value.

## Attributes Reference

In addition to all arguments above (except for `template_body` and `template_s3_uri`), the following attributes are exported:

* `arn` - Amazon Resource Name (ARN) of the conformance pack.

## Import

Config Conformance Packs can be imported using the `name`, e.g.

```
$ terraform import aws_config_conformance_pack.example example
```
