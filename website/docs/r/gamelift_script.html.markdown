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

```hcl
resource "aws_gamelift_script" "test" {
  name = "example-script"

  storage_location {
    bucket   = "${aws_s3_bucket.test.bucket}"
    key      = "${aws_s3_bucket_object.test.key}"
    role_arn = "${aws_iam_role.test.arn}"
  }

  depends_on = ["aws_iam_role_policy.test"]
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Name of the script
* `storage_location` - (Required) Information indicating where your game script files are stored. See below.
* `version` - (Optional) Version that is associated with this script.
* `tags` - (Optional) Key-value mapping of resource tags

### Nested Fields

#### `storage_location`

* `bucket` - (Required) Name of your S3 bucket.
* `key` - (Required) Name of the zip file containing your script files.
* `role_arn` - (Required) ARN of the access role that allows Amazon GameLift to access your S3 bucket.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - Gamelift Script ID.
* `arn` - Gamelift Script ARN.


## Import

Gamelift Scripts can be imported using the ID, e.g.

```
$ terraform import aws_gamelift_script.example <script-id>
```
