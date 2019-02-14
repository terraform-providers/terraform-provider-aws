---
layout: "aws"
page_title: "AWS: aws_glue_security_configuration"
sidebar_current: "docs-aws-resource-glue-security-configuration"
description: |-
  Manages a Glue Security Configuration
---

# aws_glue_security_configuration

Manages a Glue Security Configuration.

## Example Usage

```hcl
resource "aws_glue_security_configuration" "example" {
  name = "example"

  encryption_configuration {
    cloudwatch_encryption {
      cloudwatch_encryption_mode = "DISABLED"
    }

    job_bookmarks_encryption {
      job_bookmarks_encryption_mode = "DISABLED"
    }

    s3_encryption {
      kms_key_arn        = "${data.aws_kms_key.example.arn}"
      s3_encryption_mode = "SSE-KMS"
    }
  }
}
```

## Argument Reference

The following arguments are supported:

* `encryption_configuration` – (Required) Configuration block containing encryption configuration. Detailed below.
* `name` – (Required) Name of the security configuration.

### encryption_configuration Argument Reference

* `cloudwatch_encryption ` - (Required) A `cloudwatch_encryption ` block as described below, which contains encryption configuration for CloudWatch.
* `job_bookmarks_encryption ` - (Required) A `job_bookmarks_encryption ` block as described below, which contains encryption configuration for job bookmarks.
* `s3_encryption` - (Required) A `s3_encryption ` block as described below, which contains encryption configuration for S3 data.

#### cloudwatch_encryption Argument Reference

* `cloudwatch_encryption_mode` - (Optional) Encryption mode to use for CloudWatch data. Valid values: `DISABLED`, `SSE-KMS`. Default value: `DISABLED`.
* `kms_key_arn` - (Optional) Amazon Resource Name (ARN) of the KMS key to be used to encrypt the data.

#### job_bookmarks_encryption Argument Reference

* `job_bookmarks_encryption_mode` - (Optional) Encryption mode to use for job bookmarks data. Valid values: `CSE-KMS`, `DISABLED`. Default value: `DISABLED`.
* `kms_key_arn` - (Optional) Amazon Resource Name (ARN) of the KMS key to be used to encrypt the data.

#### s3_encryption Argument Reference

* `s3_encryption_mode` - (Optional) Encryption mode to use for S3 data. Valid values: `DISABLED`, `SSE-KMS`, `SSE-S3`. Default value: `DISABLED`.
* `kms_key_arn` - (Optional) Amazon Resource Name (ARN) of the KMS key to be used to encrypt the data.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - Glue security configuration name

## Import

Glue Security Configurations can be imported using `name`, e.g.

```
$ terraform import aws_glue_security_configuration.example example
```
