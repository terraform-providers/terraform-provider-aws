---
subcategory: "S3"
layout: "aws"
page_title: "AWS: aws_s3_object_copy"
description: |-
  Provides a resource for copying an S3 object.
---

# Resource: aws_s3_object_copy

Provides a resource for copying an S3 object.

## Example Usage

```hcl
resource "aws_s3_object_copy" "test" {
  bucket = "destination_bucket"
  key    = "destination_key"
  source = "source_bucket/source_key"

  grant {
    uri         = "http://acs.amazonaws.com/groups/global/AllUsers"
    type        = "Group"
    permissions = ["READ"]
  }
}
```

## Argument Reference

The following arguments are required:

* `bucket` - (Required) Name of the bucket to put the file in.
* `key` - (Required) Name of the object once it is in the bucket.
* `source` - (Required) Specifies the source object for the copy operation. You specify the value in one of two formats. For objects not accessed through an access point, specify the name of the source bucket and the key of the source object, separated by a slash (`/`). For example, `testbucket/test1.json`. For objects accessed through access points, specify the Amazon Resource Name (ARN) of the object as accessed through the access point, in the format `arn:aws:s3:<Region>:<account-id>:accesspoint/<access-point-name>/object/<key>`. For example, `arn:aws:s3:us-west-2:9999912999:accesspoint/my-access-point/object/testbucket/test1.json`.

The following arguments are optional:

* `acl` - (Optional) [Canned ACL](https://docs.aws.amazon.com/AmazonS3/latest/dev/acl-overview.html#canned-acl) to apply. Defaults to `private`. Valid values are `private`, `public-read`, `public-read-write`, `authenticated-read`, `aws-exec-read`, `bucket-owner-read`, and `bucket-owner-full-control`. Conflicts with `grant`.
* `cache_control` - (Optional) Specifies caching behavior along the request/reply chain Read [w3c cache_control](http://www.w3.org/Protocols/rfc2616/rfc2616-sec14.html#sec14.9) for further details.
* `content_disposition` - (Optional) Specifies presentational information for the object. Read [w3c content_disposition](http://www.w3.org/Protocols/rfc2616/rfc2616-sec19.html#sec19.5.1) for further information.
* `content_encoding` - (Optional) Specifies what content encodings have been applied to the object and thus what decoding mechanisms must be applied to obtain the media-type referenced by the Content-Type header field. Read [w3c content encoding](http://www.w3.org/Protocols/rfc2616/rfc2616-sec14.html#sec14.11) for further information.
* `content_language` - (Optional) Language the content is in e.g. en-US or en-GB.
* `content_type` - (Optional) Standard MIME type describing the format of the object data, e.g. `application/octet-stream`. All Valid MIME Types are valid for this input.
* `copy_if_match` - (Optional) Copies the object if its entity tag (ETag) matches the specified tag.
* `copy_if_modified_since` - (Optional) Copies the object if it has been modified since the specified time, in [RFC3339 format](https://tools.ietf.org/html/rfc3339#section-5.8).
* `copy_if_none_match` - (Optional) Copies the object if its entity tag (ETag) is different than the specified ETag.
* `copy_if_unmodified_since` - (Optional) Copies the object if it hasn't been modified since the specified time, in [RFC3339 format](https://tools.ietf.org/html/rfc3339#section-5.8).
* `customer_algorithm` - (Optional) Specifies the algorithm to use to when encrypting the object (for example, AES256).
* `customer_key` - (Optional) Specifies the customer-provided encryption key for Amazon S3 to use in encrypting data. This value is used to store the object and then it is discarded; Amazon S3 does not store the encryption key. The key must be appropriate for use with the algorithm specified in the x-amz-server-side-encryption-customer-algorithm header.
* `customer_key_md5` - (Optional) Specifies the 128-bit MD5 digest of the encryption key according to RFC 1321. Amazon S3 uses this header for a message integrity check to ensure that the encryption key was transmitted without error.
* `expected_bucket_owner` - (Optional) Account id of the expected destination bucket owner. If the destination bucket is owned by a different account, the request will fail with an HTTP 403 (Access Denied) error.
* `expected_source_bucket_owner` - (Optional) Account id of the expected source bucket owner. If the source bucket is owned by a different account, the request will fail with an HTTP 403 (Access Denied) error.
* `expires` - (Optional) Date and time at which the object is no longer cacheable, in [RFC3339 format](https://tools.ietf.org/html/rfc3339#section-5.8).
* `force_destroy` - (Optional) Allow the object to be deleted by removing any legal hold on any object version. Default is `false`. This value should be set to `true` only if the bucket has S3 object lock enabled.
* `grant` - (Optional) Configuration block for header grants. Documented below. Conflicts with `acl`.
* `kms_encryption_context` - (Optional) Specifies the AWS KMS Encryption Context to use for object encryption. The value is a base64-encoded UTF-8 string holding JSON with the encryption context key-value pairs.
* `kms_key_id` - (Optional) Specifies the AWS KMS Key ARN to use for object encryption. This value is a fully qualified **ARN** of the KMS Key. If using `aws_kms_key`, use the exported `arn` attribute: `kms_key_id = aws_kms_key.foo.arn`
* `metadata` - (Optional) A map of keys/values to provision metadata (will be automatically prefixed by `x-amz-meta-`, note that only lowercase label are currently supported by the AWS Go API).
* `metadata_directive` - (Optional) Specifies whether the metadata is copied from the source object or replaced with metadata provided in the request. Valid values are `COPY` and `REPLACE`.
* `object_lock_legal_hold_status` - (Optional) The [legal hold](https://docs.aws.amazon.com/AmazonS3/latest/dev/object-lock-overview.html#object-lock-legal-holds) status that you want to apply to the specified object. Valid values are `ON` and `OFF`.
* `object_lock_mode` - (Optional) The object lock [retention mode](https://docs.aws.amazon.com/AmazonS3/latest/dev/object-lock-overview.html#object-lock-retention-modes) that you want to apply to this object. Valid values are `GOVERNANCE` and `COMPLIANCE`.
* `object_lock_retain_until_date` - (Optional) The date and time, in [RFC3339 format](https://tools.ietf.org/html/rfc3339#section-5.8), when this object's object lock will [expire](https://docs.aws.amazon.com/AmazonS3/latest/dev/object-lock-overview.html#object-lock-retention-periods).
* `request_payer` - (Optional) Confirms that the requester knows that they will be charged for the request. Bucket owners need not specify this parameter in their requests. For information about downloading objects from requester pays buckets, see Downloading Objects in Requestor Pays Buckets (https://docs.aws.amazon.com/AmazonS3/latest/dev/ObjectsinRequesterPaysBuckets.html) in the Amazon S3 Developer Guide. If included, the only valid value is `requester`.
* `server_side_encryption` - (Optional) Specifies server-side encryption of the object in S3. Valid values are `AES256` and `aws:kms`.
* `source_customer_algorithm` - (Optional) Specifies the algorithm to use when decrypting the source object (for example, AES256).
* `source_customer_key` - (Optional) Specifies the customer-provided encryption key for Amazon S3 to use to decrypt the source object. The encryption key provided in this header must be one that was used when the source object was created.
* `source_customer_key_md5` - (Optional) Specifies the 128-bit MD5 digest of the encryption key according to RFC 1321. Amazon S3 uses this header for a message integrity check to ensure that the encryption key was transmitted without error.
* `storage_class` - (Optional) Specifies the desired [Storage Class](http://docs.aws.amazon.com/AmazonS3/latest/dev/storage-class-intro.html)
for the object. Can be either `STANDARD`, `REDUCED_REDUNDANCY`, `ONEZONE_IA`, `INTELLIGENT_TIERING`, `GLACIER`, `DEEP_ARCHIVE`, or `STANDARD_IA`. Defaults to `STANDARD`.
* `tagging_directive` - (Optional) Specifies whether the object tag-set are copied from the source object or replaced with tag-set provided in the request. Valid values are `COPY` and `REPLACE`.
* `tags` - (Optional) A map of tags to assign to the object.
* `website_redirect` - (Optional) Specifies a target URL for [website redirect](http://docs.aws.amazon.com/AmazonS3/latest/dev/how-to-page-redirect.html).

### grant

-> For more information on header grants, see the Amazon Simple Storage Service (S3) [API Reference: PutObjectAcl](https://docs.aws.amazon.com/AmazonS3/latest/API/API_PutObjectAcl.html).

This configuration block has the following required arguments:

* `permissions` - (Required) List of permissions to grant to grantee. Valid values are `READ`, `READ_ACP`, `WRITE_ACP`, `FULL_CONTROL`.
* `type` - (Required) - Type of grantee. Valid values are `CanonicalUser`, `Group`, and `AmazonCustomerByEmail`.

This configuration block has the following optional arguments (one of the three is required):

* `email` - (Optional) Email address of the grantee. Used only when `type` is `AmazonCustomerByEmail`.  
* `id` - (Optional) The canonical user ID of the grantee. Used only when `type` is `CanonicalUser`.  
* `uri` - (Optional) URI of the grantee group. Used only when `type` is `Group`.

-> **Note:** Terraform ignores all leading `/`s in the object's `key` and treats multiple `/`s in the rest of the object's `key` as a single `/`, so values of `/index.html` and `index.html` correspond to the same S3 object as do `first//second///third//` and `first/second/third/`.

## Attributes Reference

The following attributes are exported

* `etag` - The ETag generated for the object (an MD5 sum of the object content). For plaintext objects or objects encrypted with an AWS-managed key, the hash is an MD5 digest of the object data. For objects encrypted with a KMS key or objects created by either the Multipart Upload or Part Copy operation, the hash is not an MD5 digest, regardless of the method of encryption. More information on possible values can be found on [Common Response Headers](https://docs.aws.amazon.com/AmazonS3/latest/API/RESTCommonResponseHeaders.html).
* `expiration` - If the object expiration is configured, this attribute will be set.
* `id` - The `key` of the resource supplied above.
* `last_modified` - Returns the date that the object was last modified, in [RFC3339 format](https://tools.ietf.org/html/rfc3339#section-5.8).
* `request_charged` - If present, indicates that the requester was successfully charged for the request.
* `source_version_id` - Version of the copied object in the source bucket.
* `version_id` - Version ID of the newly created copy.
