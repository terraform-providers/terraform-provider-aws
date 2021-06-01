---
subcategory: "GuardDuty"
layout: aws
page_title: 'AWS: aws_guardduty_publishing_destination'
description: Provides a resource to manage a GuardDuty PublishingDestination
---

# Resource: aws_guardduty_publishing_destination

Provides a resource to manage a GuardDuty PublishingDestination. Requires an existing GuardDuty Detector.

## Example Usage

```hcl
data "aws_caller_identity" "current" {}

data "aws_region" "current" {}

data "aws_iam_policy_document" "bucket_pol" {
  statement {
    sid = "Allow PutObject"
    actions = [
      "s3:PutObject"
    ]

    resources = [
      "${aws_s3_bucket.gd_bucket.arn}/*"
    ]

    principals {
      type        = "Service"
      identifiers = ["guardduty.amazonaws.com"]
    }
  }

  statement {
    sid = "Allow GetBucketLocation"
    actions = [
      "s3:GetBucketLocation"
    ]

    resources = [
      aws_s3_bucket.gd_bucket.arn
    ]

    principals {
      type        = "Service"
      identifiers = ["guardduty.amazonaws.com"]
    }
  }
}

data "aws_iam_policy_document" "kms_pol" {

  statement {
    sid = "Allow GuardDuty to encrypt findings"
    actions = [
      "kms:GenerateDataKey"
    ]

    resources = [
      "arn:aws:kms:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:key/*"
    ]

    principals {
      type        = "Service"
      identifiers = ["guardduty.amazonaws.com"]
    }
  }

  statement {
    sid = "Allow all users to modify/delete key (test only)"
    actions = [
      "kms:*"
    ]

    resources = [
      "arn:aws:kms:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:key/*"
    ]

    principals {
      type        = "AWS"
      identifiers = ["arn:aws:iam::${data.aws_caller_identity.current.account_id}:root"]
    }
  }

}

resource "aws_guardduty_detector" "test_gd" {
  enable = true
}

resource "aws_s3_bucket" "gd_bucket" {
  bucket        = "example"
  acl           = "private"
  force_destroy = true
}

resource "aws_s3_bucket_policy" "gd_bucket_policy" {
  bucket = aws_s3_bucket.gd_bucket.id
  policy = data.aws_iam_policy_document.bucket_pol.json
}

resource "aws_kms_key" "gd_key" {
  description             = "Temporary key for AccTest of TF"
  deletion_window_in_days = 7
  policy                  = data.aws_iam_policy_document.kms_pol.json
}

resource "aws_guardduty_publishing_destination" "test" {
  detector_id     = aws_guardduty_detector.test_gd.id
  destination_arn = aws_s3_bucket.gd_bucket.arn
  kms_key_arn     = aws_kms_key.gd_key.arn

  depends_on = [
    aws_s3_bucket_policy.gd_bucket_policy,
  ]
}
```

~> **Note:** Please do not use this simple example for Bucket-Policy and KMS Key Policy in a production environment. It is much too open for such a use-case. Refer to the AWS documentation here: https://docs.aws.amazon.com/guardduty/latest/ug/guardduty_exportfindings.html

## Argument Reference

The following arguments are supported:

* `detector_id` - (Required) The detector ID of the GuardDuty.
* `destination_arn` - (Required) The bucket arn and prefix under which the findings get exported. Bucket-ARN is required, the prefix is optional and will be `AWSLogs/[Account-ID]/GuardDuty/[Region]/` if not provided
* `kms_key_arn` - (Required) The ARN of the KMS key used to encrypt GuardDuty findings. GuardDuty enforces this to be encrypted.
* `destination_type`- (Optional) Currently there is only "S3" available as destination type which is also the default value


~> **Note:** In case of missing permissions (S3 Bucket Policy _or_ KMS Key permissions) the resource will fail to create. If the permissions are changed after resource creation, this can be asked from the AWS API via the "DescribePublishingDestination" call (https://docs.aws.amazon.com/cli/latest/reference/guardduty/describe-publishing-destination.html).

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the GuardDuty PublishingDestination and the detector ID. Format: `<DetectorID>:<PublishingDestinationID>`

## Import

GuardDuty PublishingDestination can be imported using the the master GuardDuty detector ID and PublishingDestinationID, e.g.

```
$ terraform import aws_guardduty_publishing_destination.test a4b86f26fa42e7e7cf0d1c333ea77777:a4b86f27a0e464e4a7e0516d242f1234
```
