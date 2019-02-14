---
layout: "aws"
page_title: "AWS: aws_gamelift_build"
sidebar_current: "docs-aws-resource-gamelift-build"
description: |-
  Provides a Gamelift Build resource.
---

# aws_gamelift_build

Provides an Gamelift Build resource.

## Example Usage

```hcl
resource "aws_gamelift_build" "test" {
  name             = "example-build"
  operating_system = "WINDOWS_2012"

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

* `name` - (Required) Name of the build
* `operating_system` - (Required) Operating system that the game server binaries are built to run on. e.g. `WINDOWS_2012` or `AMAZON_LINUX`.
* `storage_location` - (Required) Information indicating where your game build files are stored. See below.
* `version` - (Optional) Version that is associated with this build.

### Nested Fields

#### `storage_location`

* `bucket` - (Required) Name of your S3 bucket.
* `key` - (Required) Name of the zip file containing your build files.
* `role_arn` - (Required) ARN of the access role that allows Amazon GameLift to access your S3 bucket.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - Build ID.

## Import

Gamelift Builds cannot be imported at this time.
