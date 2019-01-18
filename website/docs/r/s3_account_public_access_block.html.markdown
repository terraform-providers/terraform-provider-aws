---
layout: "aws"
page_title: "AWS: aws_s3_account_public_access_block"
sidebar_current: "docs-aws-resource-s3-account-public-access-block"
description: |-
  Manages S3 account-level Public Access Block Configuration
---

# aws_s3_account_public_access_block

Manages S3 account-level Public Access Block configuration. For more information about these settings, see the [AWS S3 Block Public Access documentation](https://docs.aws.amazon.com/AmazonS3/latest/dev/access-control-block-public-access.html).

~> **NOTE:** Each AWS account may only have one S3 Public Access Block configuration. Multiple configurations of the resource against the same AWS account will cause a perpetual difference.

-> Advanced usage: To use a custom API endpoint for this Terraform resource, use the [`s3control` endpoint provider configuration](/docs/providers/aws/index.html#s3control), not the `s3` endpoint provider configuration.

## Example Usage

```hcl
resource "aws_s3_account_public_access_block" "example" {
  block_public_acls   = true
  block_public_policy = true
}
```

## Argument Reference

The following arguments are supported:

* `account_id` - (Optional) AWS account ID to configure. Defaults to automatically determined account ID of the Terraform AWS provider.
* `block_public_acls` - (Optional) Whether Amazon S3 should block public ACLs for buckets in this account. Defaults to `false`. Enabling this setting does not affect existing policies or ACLs. When set to `true` causes the following behavior:
  * PUT Bucket acl and PUT Object acl calls will fail if the specified ACL allows public access.
  * PUT Object calls will fail if the request includes an object ACL.
* `block_public_policy` - (Optional) Whether Amazon S3 should block public bucket policies for buckets in this account. Defaults to `false`. Enabling this setting does not affect existing bucket policies. When set to `true` causes Amazon S3 to:
  * Reject calls to PUT Bucket policy if the specified bucket policy allows public access.
* `ignore_public_acls` - (Optional) Whether Amazon S3 should ignore public ACLs for buckets in this account. Defaults to `false`. Enabling this setting does not affect the persistence of any existing ACLs and doesn't prevent new public ACLs from being set. When set to `true` causes Amazon S3 to:
  * Ignore all public ACLs on buckets in this account and any objects that they contain.
* `restrict_public_buckets` - (Optional) Whether Amazon S3 should restrict public bucket policies for buckets in this account. Defaults to `false`. Enabling this setting does not affect previously stored bucket policies, except that public and cross-account access within any public bucket policy, including non-public delegation to specific accounts, is blocked. When set to `true`:
  * Only the bucket owner and AWS Services can access buckets with public policies.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - AWS account ID

## Import

`aws_s3_account_public_access_block` can be imported by using the AWS account ID, e.g.

```
$ terraform import aws_s3_account_public_access_block.example 123456789012
```
