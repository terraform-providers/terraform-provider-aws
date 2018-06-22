---
layout: "aws"
page_title: "AWS: aws_s3_bucket_replication"
sidebar_current: "docs-aws-resource-s3-bucket-replication"
description: |-
  Configures replication for an S3 bucket resource.
---

# aws_s3_bucket_replication

Configures replication for an S3 bucket resource.

## Example Usage

### Basic Usage

```hcl
provider "aws" {
  region = "eu-west-1"
}

provider "aws" {
  alias  = "central"
  region = "eu-central-1"
}

resource "aws_s3_bucket" "bucket" {
  provider = "aws.central"
  bucket   = "tf-test-bucket-12345"
  region   = "eu-central-1"

  versioning {
    enabled = true
  }
}


resource "aws_s3_bucket" "destination" {
  bucket   = "tf-test-bucket-destination-12345"
  region   = "eu-west-1"

  versioning {
    enabled = true
  }
}

resource "aws_iam_role" "replication" {
  name = "tf-iam-role-replication-12345"

  assume_role_policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "s3.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
POLICY
}

resource "aws_iam_policy" "replication" {
  name = "tf-iam-role-policy-replication-12345"

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "s3:GetReplicationConfiguration",
        "s3:ListBucket"
      ],
      "Effect": "Allow",
      "Resource": [
        "${aws_s3_bucket.bucket.arn}"
      ]
    },
    {
      "Action": [
        "s3:GetObjectVersion",
        "s3:GetObjectVersionAcl"
      ],
      "Effect": "Allow",
      "Resource": [
        "${aws_s3_bucket.bucket.arn}/*"
      ]
    },
    {
      "Action": [
        "s3:ReplicateObject",
        "s3:ReplicateDelete"
      ],
      "Effect": "Allow",
      "Resource": "${aws_s3_bucket.destination.arn}/*"
    }
  ]
}
POLICY
}

resource "aws_iam_policy_attachment" "replication" {
  name       = "tf-iam-role-attachment-replication-12345"
  roles      = ["${aws_iam_role.replication.name}"]
  policy_arn = "${aws_iam_policy.replication.arn}"
}

resource "aws_s3_bucket_replication" "replication" {
  provider = "aws.central"
  bucket = "${aws_s3_bucket.bucket.id}"

  replication_configuration {
    role = "${aws_iam_role.replication.arn}"
    rules {
      id     = "foobar"
      prefix = "foo"
      status = "Enabled"

      destination {
        bucket        = "${aws_s3_bucket.destination.arn}"
        storage_class = "STANDARD"
      }
    }
  }
}
```

## Argument Reference

The following arguments are supported:

* `bucket` - (Required) The name of the bucket to replicate from.
* `replication_configuration` - (Required) A configuration of [replication configuration](http://docs.aws.amazon.com/AmazonS3/latest/dev/crr.html) (documented below).

The `replication_configuration` object supports the following:

* `role` - (Required) The ARN of the IAM role for Amazon S3 to assume when replicating the objects.
* `rules` - (Required) Specifies the rules managing the replication (documented below).

The `rules` object supports the following:

* `id` - (Optional) Unique identifier for the rule.
* `destination` - (Required) Specifies the destination for the rule (documented below).
* `source_selection_criteria` - (Optional) Specifies special object selection criteria (documented below).
* `prefix` - (Required) Object keyname prefix identifying one or more objects to which the rule applies. Set as an empty string to replicate the whole bucket.
* `status` - (Required) The status of the rule. Either `Enabled` or `Disabled`. The rule is ignored if status is not Enabled.

The `destination` object supports the following:

* `bucket` - (Required) The ARN of the S3 bucket where you want Amazon S3 to store replicas of the object identified by the rule.
* `storage_class` - (Optional) The class of storage used to store the object.
* `replica_kms_key_id` - (Optional) Destination KMS encryption key ID for SSE-KMS replication. Must be used in conjunction with
  `sse_kms_encrypted_objects` source selection criteria.

The `source_selection_criteria` object supports the following:

* `sse_kms_encrypted_objects` - (Optional) Match SSE-KMS encrypted objects (documented below). If specified, `replica_kms_key_id`
   in `destination` must be specified as well.

The `sse_kms_encrypted_objects` object supports the following:

* `enabled` - (Required) Boolean which indicates if this criteria is enabled.
