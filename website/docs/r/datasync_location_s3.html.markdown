---
subcategory: "DataSync"
layout: "aws"
page_title: "AWS: aws_datasync_location_s3"
description: |-
  Manages an AWS DataSync S3 Location
---

# Resource: aws_datasync_location_s3

Manages an S3 Location within AWS DataSync.

## Example Usage

```terraform
resource "aws_datasync_location_s3" "example" {
  s3_bucket_arn = aws_s3_bucket.example.arn
  subdirectory  = "/example/prefix"

  s3_config {
    bucket_access_role_arn = aws_iam_role.example.arn
  }
}
```

## Argument Reference

The following arguments are supported:

* `agent_arns` - (Optional) A list of DataSync Agent ARNs with which this location will be associated.
* `s3_bucket_arn` - (Required) Amazon Resource Name (ARN) of the S3 Bucket.
* `s3_config` - (Required) Configuration block containing information for connecting to S3.
* `s3_storage_class` - (Optional) The Amazon S3 storage class that you want to store your files in when this location is used as a task destination. [Valid values](https://docs.aws.amazon.com/datasync/latest/userguide/create-s3-location.html#using-storage-classes)  
* `subdirectory` - (Required) Prefix to perform actions as source or destination.
* `tags` - (Optional) Key-value pairs of resource tags to assign to the DataSync Location. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### s3_config Argument Reference

The following arguments are supported inside the `s3_config` configuration block:

* `bucket_access_role_arn` - (Required) Amazon Resource Names (ARN) of the IAM Role used to connect to the S3 Bucket.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - Amazon Resource Name (ARN) of the DataSync Location.
* `arn` - Amazon Resource Name (ARN) of the DataSync Location.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block).

## Import

`aws_datasync_location_s3` can be imported by using the DataSync Task Amazon Resource Name (ARN), e.g.

```
$ terraform import aws_datasync_location_s3.example arn:aws:datasync:us-east-1:123456789012:location/loc-12345678901234567
```
