---
subcategory: "S3"
layout: "aws"
page_title: "AWS: aws_s3_bucket_object"
description: |-
    Provides metadata and optionally content of an S3 object
---

# Data Source: aws_s3_bucket_object

The S3 object data source allows access to the metadata and
_optionally_ (see below) content of an object stored inside S3 bucket.

~> **Note:** The content of an object (`body` field) is available only for objects which have a human-readable `Content-Type` (`text/*` and `application/json`). This is to prevent printing unsafe characters and potentially downloading large amount of data which would be thrown away in favour of metadata.

## Example Usage

The following example retrieves a text object (which must have a `Content-Type`
value starting with `text/`) and uses it as the `user_data` for an EC2 instance:

```hcl
data "aws_s3_bucket_object" "bootstrap_script" {
  bucket = "ourcorp-deploy-config"
  key    = "ec2-bootstrap-script.sh"
}

resource "aws_instance" "example" {
  instance_type = "t2.micro"
  ami           = "ami-2757f631"
  user_data     = data.aws_s3_bucket_object.bootstrap_script.body
}
```

The following, more-complex example retrieves only the metadata for a zip
file stored in S3, which is then used to pass the most recent `version_id`
to AWS Lambda for use as a function implementation. More information about
Lambda functions is available in the documentation for
[`aws_lambda_function`](/docs/providers/aws/r/lambda_function.html).

```hcl
data "aws_s3_bucket_object" "lambda" {
  bucket = "ourcorp-lambda-functions"
  key    = "hello-world.zip"
}

resource "aws_lambda_function" "test_lambda" {
  s3_bucket         = data.aws_s3_bucket_object.lambda.bucket
  s3_key            = data.aws_s3_bucket_object.lambda.key
  s3_object_version = data.aws_s3_bucket_object.lambda.version_id
  function_name     = "lambda_function_name"
  role              = aws_iam_role.iam_for_lambda.arn # (not shown)
  handler           = "exports.test"
}
```

## Argument Reference

The following arguments are supported:

* `bucket` - (Required) The name of the bucket to read the object from. Alternatively, an [S3 access point](https://docs.aws.amazon.com/AmazonS3/latest/dev/using-access-points.html) ARN can be specified
* `key` - (Required) The full path to the object inside the bucket
* `version_id` - (Optional) Specific version ID of the object returned (defaults to latest version)

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `body` - Object data (see **limitations above** to understand cases in which this field is actually available)
* `cache_control` - Specifies caching behavior along the request/reply chain.
* `content_disposition` - Specifies presentational information for the object.
* `content_encoding` - Specifies what content encodings have been applied to the object and thus what decoding mechanisms must be applied to obtain the media-type referenced by the Content-Type header field.
* `content_language` - The language the content is in.
* `content_length` - Size of the body in bytes.
* `content_type` - A standard MIME type describing the format of the object data.
* `etag` - [ETag](https://en.wikipedia.org/wiki/HTTP_ETag) generated for the object (an MD5 sum of the object content in case it's not encrypted)
* `expiration` - If the object expiration is configured (see [object lifecycle management](http://docs.aws.amazon.com/AmazonS3/latest/dev/object-lifecycle-mgmt.html)), the field includes this header. It includes the expiry-date and rule-id key value pairs providing object expiration information. The value of the rule-id is URL encoded.
* `expires` - The date and time at which the object is no longer cacheable.
* `last_modified` - Last modified date of the object in RFC1123 format (e.g. `Mon, 02 Jan 2006 15:04:05 MST`)
* `metadata` - A map of metadata stored with the object in S3
* `object_lock_legal_hold_status` - Indicates whether this object has an active [legal hold](https://docs.aws.amazon.com/AmazonS3/latest/dev/object-lock-overview.html#object-lock-legal-holds). This field is only returned if you have permission to view an object's legal hold status.
* `object_lock_mode` - The object lock [retention mode](https://docs.aws.amazon.com/AmazonS3/latest/dev/object-lock-overview.html#object-lock-retention-modes) currently in place for this object.
* `object_lock_retain_until_date` - The date and time when this object's object lock will expire.
* `server_side_encryption` - If the object is stored using server-side encryption (KMS or Amazon S3-managed encryption key), this field includes the chosen encryption and algorithm used.
* `sse_kms_key_id` - If present, specifies the ID of the Key Management Service (KMS) master encryption key that was used for the object.
* `storage_class` - [Storage class](http://docs.aws.amazon.com/AmazonS3/latest/dev/storage-class-intro.html) information of the object. Available for all objects except for `Standard` storage class objects.
* `version_id` - The latest version ID of the object returned.
* `website_redirect_location` - If the bucket is configured as a website, redirects requests for this object to another object in the same bucket or to an external URL. Amazon S3 stores the value of this header in the object metadata.
* `tags`  - A map of tags assigned to the object.

-> **Note:** Terraform ignores all leading `/`s in the object's `key` and treats multiple `/`s in the rest of the object's `key` as a single `/`, so values of `/index.html` and `index.html` correspond to the same S3 object as do `first//second///third//` and `first/second/third/`.
