---
subcategory: "XRay"
layout: "aws"
page_title: "AWS: aws_xray_encryption_config"
description: |-
    Creates and manages an AWS XRay Encryption Config.
---

# Resource: aws_xray_encryption_config

Creates and manages an AWS XRay Encryption Config.

~> **NOTE:** Removing this resource from Terraform has no effect to the encryption configuration within X-Ray.

## Example Usage

```hcl
resource "aws_xray_encryption_config" "example" {
  type = "NONE"
}
```

## Example Usage with KMS Key

```hcl
resource "aws_kms_key" "example" {
  description             = "Some Key"
  deletion_window_in_days = 7

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Id": "kms-tf-1",
  "Statement": [
    {
      "Sid": "Enable IAM User Permissions",
      "Effect": "Allow",
      "Principal": {
        "AWS": "*"
      },
      "Action": "kms:*",
      "Resource": "*"
    }
  ]
}
POLICY
}

resource "aws_xray_encryption_config" "example" {
  type   = "KMS"
  key_id = aws_kms_key.example.arn
}
```

## Argument Reference

* `type` - (Required) The type of encryption. Set to `KMS` to use your own key for encryption. Set to `NONE` for default encryption.
* `key_id` - (Optional) An AWS KMS customer master key (CMK) ARN.

## Attributes Reference

In addition to the arguments above, the following attributes are exported:

* `id` - Region name.

## Import

XRay Encryption Config can be imported using the region name, e.g.

```
$ terraform import aws_xray_encryption_config.example us-west-2
```
