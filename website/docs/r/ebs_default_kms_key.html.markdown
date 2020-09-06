---
subcategory: "EC2"
layout: "aws"
page_title: "AWS: aws_ebs_default_kms_key"
description: |-
  Manages the default customer master key (CMK) that your AWS account uses to encrypt EBS volumes.
---

# Resource: aws_ebs_default_kms_key

Provides a resource to manage the default customer master key (CMK) that your AWS account uses to encrypt EBS volumes.

Your AWS account has an AWS-managed default CMK that is used for encrypting an EBS volume when no CMK is specified in the API call that creates the volume.
By using the `aws_ebs_default_kms_key` resource, you can specify a customer-managed CMK to use in place of the AWS-managed default CMK.

~> **NOTE:** Creating an `aws_ebs_default_kms_key` resource does not enable default EBS encryption. Use the [`aws_ebs_encryption_by_default`](ebs_encryption_by_default.html) to enable default EBS encryption.

~> **NOTE:** Destroying this resource will reset the default CMK to the account's AWS-managed default CMK for EBS.

## Example Usage

```hcl
resource "aws_ebs_default_kms_key" "example" {
  key_arn = aws_kms_key.example.arn
}
```

## Argument Reference

The following arguments are supported:

* `key_arn` - (Required, ForceNew) The ARN of the AWS Key Management Service (AWS KMS) customer master key (CMK) to use to encrypt the EBS volume.

## Import

The EBS default KMS CMK can be imported with the KMS key ARN, e.g.

```console
$ terraform import aws_ebs_default_kms_key.example arn:aws:kms:us-east-1:123456789012:key/abcd-1234
```
