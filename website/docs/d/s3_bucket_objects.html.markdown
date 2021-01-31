---
subcategory: "S3"
layout: "aws"
page_title: "AWS: aws_s3_bucket_objects"
description: |-
    Returns keys and metadata of S3 objects
---

# Data Source: aws_s3_bucket_objects

~> **NOTE on `max_keys`:** Retrieving very large numbers of keys can adversely affect Terraform's performance.

The bucket-objects data source returns keys (i.e., file names) and other metadata about objects in an S3 bucket.

## Example Usage

The following example retrieves a list of all object keys in an S3 bucket and creates corresponding Terraform object data sources:

```hcl
data "aws_s3_bucket_objects" "my_objects" {
  bucket = "ourcorp"
}

data "aws_s3_bucket_object" "object_info" {
  count  = length(data.aws_s3_bucket_objects.my_objects.keys)
  key    = element(data.aws_s3_bucket_objects.my_objects.keys, count.index)
  bucket = data.aws_s3_bucket_objects.my_objects.bucket
}
```

## Argument Reference

The following arguments are supported:

* `bucket` - (Required) Lists object keys in this S3 bucket. Alternatively, an [S3 access point](https://docs.aws.amazon.com/AmazonS3/latest/dev/using-access-points.html) ARN can be specified
* `prefix` - (Optional) Limits results to object keys with this prefix (Default: none)
* `delimiter` - (Optional) A character used to group keys (Default: none)
* `encoding_type` - (Optional) Encodes keys using this method (Default: none; besides none, only "url" can be used)
* `max_keys` - (Optional) Maximum object keys to return (Default: 1000)
* `start_after` - (Optional) Returns key names lexicographically after a specific object key in your bucket (Default: none; S3 lists object keys in UTF-8 character encoding in lexicographical order)
* `fetch_owner` - (Optional) Boolean specifying whether to populate the owner list (Default: false)

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `keys` - List of strings representing object keys
* `common_prefixes` - List of any keys between `prefix` and the next occurrence of `delimiter` (i.e., similar to subdirectories of the `prefix` "directory"); the list is only returned when you specify `delimiter`
* `id` - S3 Bucket.
* `owners` - List of strings representing object owner IDs (see `fetch_owner` above)
