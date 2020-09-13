---
subcategory: "CodeBuild"
layout: "aws"
page_title: "AWS: aws_codebuild_report_group"
description: |-
  Provides a CodeBuild Report Group resource.
---

# Resource: aws_codebuild_report_group

Provides a CodeBuild Report Groups Resource.

## Example Usage

```hcl
resource "aws_kms_key" "example" {
  description             = "my test kms key"
  deletion_window_in_days = 7

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Id": "kms-tf-1",
  "Statement": [
    {
      "Sid": "Enable IAM User Permissions",
      "Effect": "Allow",
      "Principal": {
        "AWS": "*"
      },
      "Action": "kms:*",
      "Resource": "*"
    }
  ]
}
POLICY
}

resource "aws_s3_bucket" "example" {
  bucket = "my-test"
}

resource "aws_codebuild_report_group" "example" {
  name = "my test report group"
  type = "TEST"

  export_config {
    type = "S3"

    s3_destination {
      bucket              = aws_s3_bucket.example.id
      encryption_disabled = false
      encryption_key      = aws_kms_key.example.arn
      packaging           = "NONE"
      path                = "/some"
    }
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of a Report Group.
* `type` - (Required) The type of the Report Group. Valid value are `TEST` and `CODE_COVERAGE`.
* `export_config` - (Required) Information about the destination where the raw data of this Report Group is exported. see [Export Config](#export-config) documented below.
* `tags` - (Optional) Key-value mapping of resource tags

### Export Config

* `type` - (Required) The export configuration type. Valid values are `S3` and `NO_EXPORT`.
* `s3_destination` - (Required) contains information about the S3 bucket where the run of a report is exported. see [S3 Destination](#s3-destination) documented below.

#### S3 Destination

* `bucket`- (Required) The name of the S3 bucket where the raw data of a report are exported.
* `encryption_key` - (Required) The encryption key for the report's encrypted raw data. The KMS key ARN.
* `encryption_disabled`- (Optional) A boolean value that specifies if the results of a report are encrypted.
 **Note: the API does not currently allow setting encryption as disabled**
* `packaging` - (Optional) The type of build output artifact to create. Valid values are: `NONE` (default) and `ZIP`.
* `path` - (Optional) The path to the exported report's raw data results.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ARN of Report Group.
* `arn` - The ARN of Report Group.
* `created` - The date and time this Report Group was created.

## Import

CodeBuild Report Group can be imported using the CodeBuild Report Group arn, e.g.

```
$ terraform import aws_codebuild_report_group.example arn:aws:codebuild:us-west-2:123456789:report-group/report-group-name
```
