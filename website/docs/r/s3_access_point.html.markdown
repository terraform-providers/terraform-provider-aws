---
subcategory: "S3"
layout: "aws"
page_title: "AWS: aws_s3_access_point"
description: |-
  Manages an S3 Access Point.
---

# Resource: aws_s3_access_point

Provides a resource to manage an S3 Access Point.

-> Advanced usage: To use a custom API endpoint for this Terraform resource, use the [`s3control` endpoint provider configuration](/docs/providers/aws/index.html#s3control), not the `s3` endpoint provider configuration.

## Example Usage

### AWS Partition Bucket

```hcl
resource "aws_s3_bucket" "example" {
  bucket = "example"
}

resource "aws_s3_access_point" "example" {
  bucket = aws_s3_bucket.example.id
  name   = "example"
}
```

### S3 on Outposts Bucket

```hcl
resource "aws_s3control_bucket" "example" {
  bucket = "example"
}

resource "aws_s3_access_point" "example" {
  bucket = aws_s3control_bucket.example.arn
  name   = "example"

  # VPC must be specified for S3 on Outposts
  vpc_configuration {
    vpc_id = aws_vpc.example.id
  }
}

resource "aws_vpc" "example" {
  cidr_block = "10.0.0.0/16"
}
```

## Argument Reference

The following arguments are required:

* `bucket` - (Required) The name of an AWS Partition S3 Bucket or the Amazon Resource Name (ARN) of S3 on Outposts Bucket that you want to associate this access point with.
* `name` - (Required) The name you want to assign to this access point.

The following arguments are optional:

* `account_id` - (Optional) The AWS account ID for the owner of the bucket for which you want to create an access point. Defaults to automatically determined account ID of the Terraform AWS provider.
* `policy` - (Optional) A valid JSON document that specifies the policy that you want to apply to this access point.
* `public_access_block_configuration` - (Optional) Configuration block to manage the `PublicAccessBlock` configuration that you want to apply to this Amazon S3 bucket. You can enable the configuration options in any combination. Detailed below.
* `vpc_configuration` - (Optional) Configuration block to restrict access to this access point to requests from the specified Virtual Private Cloud (VPC). Required for S3 on Outposts. Detailed below.

### public_access_block_configuration Configuration Block

The following arguments are optional:

* `block_public_acls` - (Optional) Whether Amazon S3 should block public ACLs for buckets in this account. Defaults to `true`. Enabling this setting does not affect existing policies or ACLs. When set to `true` causes the following behavior:
    * PUT Bucket acl and PUT Object acl calls fail if the specified ACL is public.
    * PUT Object calls fail if the request includes a public ACL.
    * PUT Bucket calls fail if the request includes a public ACL.
* `block_public_policy` - (Optional) Whether Amazon S3 should block public bucket policies for buckets in this account. Defaults to `true`. Enabling this setting does not affect existing bucket policies. When set to `true` causes Amazon S3 to:
    * Reject calls to PUT Bucket policy if the specified bucket policy allows public access.
* `ignore_public_acls` - (Optional) Whether Amazon S3 should ignore public ACLs for buckets in this account. Defaults to `true`. Enabling this setting does not affect the persistence of any existing ACLs and doesn't prevent new public ACLs from being set. When set to `true` causes Amazon S3 to:
    * Ignore all public ACLs on buckets in this account and any objects that they contain.
* `restrict_public_buckets` - (Optional) Whether Amazon S3 should restrict public bucket policies for buckets in this account. Defaults to `true`. Enabling this setting does not affect previously stored bucket policies, except that public and cross-account access within any public bucket policy, including non-public delegation to specific accounts, is blocked. When set to `true`:
    * Only the bucket owner and AWS Services can access buckets with public policies.

### vpc_configuration Configuration Block

The following arguments are required:

* `vpc_id` - (Required)  This access point will only allow connections from the specified VPC ID.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - Amazon Resource Name (ARN) of the S3 Access Point.
* `domain_name` - The DNS domain name of the S3 Access Point in the format _`name`_-_`account_id`_.s3-accesspoint._region_.amazonaws.com.
Note: S3 access points only support secure access by HTTPS. HTTP isn't supported.
* `has_public_access_policy` - Indicates whether this access point currently has a policy that allows public access.
* `id` - For Access Point of an AWS Partition S3 Bucket, the AWS account ID and access point name separated by a colon (`:`). For S3 on Outposts Bucket, the Amazon Resource Name (ARN) of the Access Point.
* `network_origin` - Indicates whether this access point allows access from the public Internet. Values are `VPC` (the access point doesn't allow access from the public Internet) and `Internet` (the access point allows access from the public Internet, subject to the access point and bucket access policies).

## Import

For Access Points associated with an AWS Partition S3 Bucket, this resource can be imported using the `account_id` and `name` separated by a colon (`:`), e.g.

```
$ terraform import aws_s3_access_point.example 123456789012:example
```

For Access Points associated with an S3 on Outposts Bucket, this resource can be imported using the Amazon Resource Name (ARN), e.g.

```
$ terraform import aws_s3_access_point.example arn:aws:s3-outposts:us-east-1:123456789012:outpost/op-1234567890123456/accesspoint/example
```
