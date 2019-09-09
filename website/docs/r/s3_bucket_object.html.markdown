---
layout: "aws"
page_title: "AWS: aws_s3_bucket_object"
sidebar_current: "docs-aws-resource-s3-bucket-object"
description: |-
  Provides a S3 bucket object resource.
---

# Resource: aws_s3_bucket_object

Provides a S3 bucket object resource.

## Example Usage

### Uploading a file to a bucket

```hcl
resource "aws_s3_bucket_object" "object" {
  bucket = "your_bucket_name"
  key    = "new_object_key"
  source = "path/to/file"

  # The filemd5() function is available in Terraform 0.11.12 and later
  # For Terraform 0.11.11 and earlier, use the md5() function and the file() function:
  # etag = "${md5(file("path/to/file"))}"
  etag = "${filemd5("path/to/file")}"
}
```

### Encrypting with KMS Key

```hcl
resource "aws_kms_key" "examplekms" {
  description             = "KMS key 1"
  deletion_window_in_days = 7
}

resource "aws_s3_bucket" "examplebucket" {
  bucket = "examplebuckettftest"
  acl    = "private"
}

resource "aws_s3_bucket_object" "examplebucket_object" {
  key        = "someobject"
  bucket     = "${aws_s3_bucket.examplebucket.id}"
  source     = "index.html"
  kms_key_id = "${aws_kms_key.examplekms.arn}"
}
```

### Server Side Encryption with S3 Default Master Key

```hcl
resource "aws_s3_bucket" "examplebucket" {
  bucket = "examplebuckettftest"
  acl    = "private"
}

resource "aws_s3_bucket_object" "examplebucket_object" {
  key                    = "someobject"
  bucket                 = "${aws_s3_bucket.examplebucket.id}"
  source                 = "index.html"
  server_side_encryption = "aws:kms"
}
```

### Server Side Encryption with AWS-Managed Key

```hcl
resource "aws_s3_bucket" "examplebucket" {
  bucket = "examplebuckettftest"
  acl    = "private"
}

resource "aws_s3_bucket_object" "examplebucket_object" {
  key                    = "someobject"
  bucket                 = "${aws_s3_bucket.examplebucket.id}"
  source                 = "index.html"
  server_side_encryption = "AES256"
}
```

## Argument Reference

-> **Note:** If you specify `content_encoding` you are responsible for encoding the body appropriately. `source`, `content`, and `content_base64` all expect already encoded/compressed bytes.

The following arguments are supported:

* `bucket` - (Required) The name of the bucket to put the file in.
* `key` - (Required) The name of the object once it is in the bucket.
* `source` - (Optional, conflicts with `content` and `content_base64`) The path to a file that will be read and uploaded as raw bytes for the object content.
* `content` - (Optional, conflicts with `source` and `content_base64`) Literal string value to use as the object content, which will be uploaded as UTF-8-encoded text.
* `content_base64` - (Optional, conflicts with `source` and `content`) Base64-encoded data that will be decoded and uploaded as raw bytes for the object content. This allows safely uploading non-UTF8 binary data, but is recommended only for small content such as the result of the `gzipbase64` function with small text strings. For larger objects, use `source` to stream the content from a disk file.
* `acl` - (Optional) The [canned ACL](https://docs.aws.amazon.com/AmazonS3/latest/dev/acl-overview.html#canned-acl) to apply. Defaults to "private".
* `cache_control` - (Optional) Specifies caching behavior along the request/reply chain Read [w3c cache_control](http://www.w3.org/Protocols/rfc2616/rfc2616-sec14.html#sec14.9) for further details.
* `content_disposition` - (Optional) Specifies presentational information for the object. Read [w3c content_disposition](http://www.w3.org/Protocols/rfc2616/rfc2616-sec19.html#sec19.5.1) for further information.
* `content_encoding` - (Optional) Specifies what content encodings have been applied to the object and thus what decoding mechanisms must be applied to obtain the media-type referenced by the Content-Type header field. Read [w3c content encoding](http://www.w3.org/Protocols/rfc2616/rfc2616-sec14.html#sec14.11) for further information.
* `content_language` - (Optional) The language the content is in e.g. en-US or en-GB.
* `content_type` - (Optional) A standard MIME type describing the format of the object data, e.g. application/octet-stream. All Valid MIME Types are valid for this input.
* `website_redirect` - (Optional) Specifies a target URL for [website redirect](http://docs.aws.amazon.com/AmazonS3/latest/dev/how-to-page-redirect.html).
* `storage_class` - (Optional) Specifies the desired [Storage Class](http://docs.aws.amazon.com/AmazonS3/latest/dev/storage-class-intro.html)
for the object. Can be either "`STANDARD`", "`REDUCED_REDUNDANCY`", "`ONEZONE_IA`", "`INTELLIGENT_TIERING`", "`GLACIER`", "`DEEP_ARCHIVE`", or "`STANDARD_IA`". Defaults to "`STANDARD`".
* `etag` - (Optional) Used to trigger updates. The only meaningful value is `${filemd5("path/to/file")}` (Terraform 0.11.12 or later) or `${md5(file("path/to/file"))}` (Terraform 0.11.11 or earlier).
This attribute is not compatible with KMS encryption, `kms_key_id` or `server_side_encryption = "aws:kms"`.
* `server_side_encryption` - (Optional) Specifies server-side encryption of the object in S3. Valid values are "`AES256`" and "`aws:kms`".
* `kms_key_id` - (Optional) Specifies the AWS KMS Key ARN to use for object encryption.
This value is a fully qualified **ARN** of the KMS Key. If using `aws_kms_key`,
use the exported `arn` attribute:
      `kms_key_id = "${aws_kms_key.foo.arn}"`
* `metadata` - (Optional) A mapping of keys/values to provision metadata (will be automatically prefixed by `x-amz-meta-`, note that only lowercase label are currently supported by the AWS Go API).
* `tags` - (Optional) A mapping of tags to assign to the object.

If no content is provided through `source`, `content` or `content_base64`, then the object will be empty.

## Attributes Reference

The following attributes are exported

* `id` - the `key` of the resource supplied above
* `etag` - the ETag generated for the object (an MD5 sum of the object content). For plaintext objects or objects encrypted with an AWS-managed key, the hash is an MD5 digest of the object data. For objects encrypted with a KMS key or objects created by either the Multipart Upload or Part Copy operation, the hash is not an MD5 digest, regardless of the method of encryption. More information on possible values can be found on [Common Response Headers](https://docs.aws.amazon.com/AmazonS3/latest/API/RESTCommonResponseHeaders.html).
* `version_id` - A unique version ID value for the object, if bucket versioning
is enabled.

## Import

Objects can be imported using the `id`. The `id` is the bucket name and the key together e.g.

```
$ terraform import aws_s3_bucket_object.object some-bucket-name/some/key.txt
```

Additionally, s3 url syntax can be used, e.g.

```
$ terraform import aws_s3_bucket_object.object s3://some-bucket-name/some/key.txt
```