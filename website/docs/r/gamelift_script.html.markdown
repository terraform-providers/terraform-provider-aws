---
subcategory: "Gamelift"
layout: "aws"
page_title: "AWS: aws_gamelift_script"
description: |-
  Provides a Gamelift Script resource.
---

# Resource: aws_gamelift_script

Provides an Gamelift Script resource.

## Example Usage

```terraform
resource "aws_gamelift_script" "example" {
  name = "example-script"

  storage_location {
    bucket   = aws_s3_bucket.example.bucket
    key      = aws_s3_bucket_object.example.key
    role_arn = aws_iam_role.example.arn
  }
}
```

## Example with local zip file

```terraform
resource "aws_gamelift_script" "example" {
  name     = "example-script"
  zip_file = "script.zip"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Name of the script
* `storage_location` - (Optional) Information indicating where your game script files are stored. See [Storage Location](#storage-location) details below. Either one of `storage_location` of `zip_file` is required.
* `zip_file` - (Optional) A data object containing your Realtime scripts and dependencies as a zip file. The zip file can have one or multiple files. Maximum size of a zip file is 5 MB. Either one of `storage_location` of `zip_file` is required.
* `version` - (Optional) Version that is associated with this script.
* `tags` - (Optional) Key-value mapping of resource tags

### Storage Location

* `bucket` - (Required) Name of your S3 bucket.
* `key` - (Required) Name of the zip file containing your script files.
* `role_arn` - (Required) ARN of the access role that allows Amazon GameLift to access your S3 bucket.
* `object_version` - (Optional) The version of the file, if object versioning is turned on for the bucket.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - Gamelift Script ID.
* `arn` - Gamelift Script ARN.

## Import

Gamelift Scripts can be imported using the ID, e.g.

```
$ terraform import aws_gamelift_script.example <script-id>
```
