---
subcategory: "KMS"
layout: "aws"
page_title: "AWS: aws_kms_grant"
description: |-
  Provides a resource-based access control mechanism for KMS Customer Master Keys.
---

# Resource: aws_kms_grant

Provides a resource-based access control mechanism for a KMS customer master key.

## Example Usage

```hcl
resource "aws_kms_key" "a" {}

resource "aws_iam_role" "a" {
  name = "iam-role-for-grant"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "lambda.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_kms_grant" "a" {
  name              = "my-grant"
  key_id            = aws_kms_key.a.key_id
  grantee_principal = aws_iam_role.a.arn
  operations        = ["Encrypt", "Decrypt", "GenerateDataKey"]

  constraints {
    encryption_context_equals = {
      Department = "Finance"
    }
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Optional, Forces new resources) A friendly name for identifying the grant.
* `key_id` - (Required, Forces new resources) The unique identifier for the customer master key (CMK) that the grant applies to. Specify the key ID or the Amazon Resource Name (ARN) of the CMK. To specify a CMK in a different AWS account, you must use the key ARN.
* `grantee_principal` - (Required, Forces new resources) The principal that is given permission to perform the operations that the grant permits in ARN format. Note that due to eventual consistency issues around IAM principals, terraform's state may not always be refreshed to reflect what is true in AWS.
* `operations` - (Required, Forces new resources) A list of operations that the grant permits. The permitted values are: `Decrypt, Encrypt, GenerateDataKey, GenerateDataKeyWithoutPlaintext, ReEncryptFrom, ReEncryptTo, CreateGrant, RetireGrant, DescribeKey`
* `retiring_principal` - (Optional, Forces new resources) The principal that is given permission to retire the grant by using RetireGrant operation in ARN format. Note that due to eventual consistency issues around IAM principals, terraform's state may not always be refreshed to reflect what is true in AWS.
* `constraints` - (Optional, Forces new resources) A structure that you can use to allow certain operations in the grant only when the desired encryption context is present. For more information about encryption context, see [Encryption Context](http://docs.aws.amazon.com/kms/latest/developerguide/encryption-context.html).
* `grant_creation_tokens` - (Optional, Forces new resources) A list of grant tokens to be used when creating the grant. See [Grant Tokens](http://docs.aws.amazon.com/kms/latest/developerguide/concepts.html#grant_token) for more information about grant tokens.
* `retire_on_delete` -(Defaults to false, Forces new resources) If set to false (the default) the grants will be revoked upon deletion, and if set to true the grants will try to be retired upon deletion. Note that retiring grants requires special permissions, hence why we default to revoking grants.
  See [RetireGrant](https://docs.aws.amazon.com/kms/latest/APIReference/API_RetireGrant.html) for more information.

The `constraints` block supports the following arguments:

* `encryption_context_equals` - (Optional) A list of key-value pairs that must match the encryption context in subsequent cryptographic operation requests. The grant allows the operation only when the encryption context in the request is the same as the encryption context specified in this constraint. Conflicts with `encryption_context_subset`.
* `encryption_context_subset` - (Optional) A list of key-value pairs that must be included in the encryption context of subsequent cryptographic operation requests. The grant allows the cryptographic operation only when the encryption context in the request includes the key-value pairs specified in this constraint, although it can include additional key-value pairs. Conflicts with `encryption_context_equals`.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `grant_id` - The unique identifier for the grant.
* `grant_token` - The grant token for the created grant. For more information, see [Grant Tokens](http://docs.aws.amazon.com/kms/latest/developerguide/concepts.html#grant_token).

## Import

KMS Grants can be imported using the Key ID and Grant ID separated by a colon (`:`), e.g.

```
$ terraform import aws_kms_grant.test 1234abcd-12ab-34cd-56ef-1234567890ab: abcde1237f76e4ba7987489ac329fbfba6ad343d6f7075dbd1ef191f0120514
```
