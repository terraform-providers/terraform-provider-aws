---
layout: "aws"
page_title: "AWS: cloudtrail"
sidebar_current: "docs-aws-resource-cloudtrail"
description: |-
  Provides a CloudTrail resource.
---

# aws_cloudtrail

Provides a CloudTrail resource.

~> *NOTE:* For a multi-region trail, this resource must be in the home region of the trail.

~> *NOTE:* For an organization trail, this resource must be in the master account of the organization.

## Example Usage

### Basic

Enable CloudTrail to capture all compatible management events in region.
For capturing events from services like IAM, `include_global_service_events` must be enabled.

```hcl
resource "aws_cloudtrail" "foobar" {
  name                          = "tf-trail-foobar"
  s3_bucket_name                = "${aws_s3_bucket.foo.id}"
  s3_key_prefix                 = "prefix"
  include_global_service_events = false
}

resource "aws_s3_bucket" "foo" {
  bucket        = "tf-test-trail"
  force_destroy = true

  policy = <<POLICY
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "AWSCloudTrailAclCheck",
            "Effect": "Allow",
            "Principal": {
              "Service": "cloudtrail.amazonaws.com"
            },
            "Action": "s3:GetBucketAcl",
            "Resource": "arn:aws:s3:::tf-test-trail"
        },
        {
            "Sid": "AWSCloudTrailWrite",
            "Effect": "Allow",
            "Principal": {
              "Service": "cloudtrail.amazonaws.com"
            },
            "Action": "s3:PutObject",
            "Resource": "arn:aws:s3:::tf-test-trail/*",
            "Condition": {
                "StringEquals": {
                    "s3:x-amz-acl": "bucket-owner-full-control"
                }
            }
        }
    ]
}
POLICY
}
```

### Data Event Logging

CloudTrail can log [Data Events](https://docs.aws.amazon.com/awscloudtrail/latest/userguide/logging-management-and-data-events-with-cloudtrail.html#logging-data-events) for certain services such as S3 bucket objects and Lambda function invocations. Additional information about data event configuration can be found in the [CloudTrail API DataResource documentation](https://docs.aws.amazon.com/awscloudtrail/latest/APIReference/API_DataResource.html).

#### Logging All Lambda Function Invocations

```hcl
resource "aws_cloudtrail" "example" {
  # ... other configuration ...

  event_selector {
    read_write_type           = "All"
    include_management_events = true

    data_resource {
      type   = "AWS::Lambda::Function"
      values = ["arn:aws:lambda"]
    }
  }
}
```

#### Logging All S3 Bucket Object Events

```hcl
resource "aws_cloudtrail" "example" {
  # ... other configuration ...

  event_selector {
    read_write_type           = "All"
    include_management_events = true

    data_resource {
      type   = "AWS::S3::Object"
      values = ["arn:aws:s3:::"]
    }
  }
}
```

#### Logging Individual S3 Bucket Events

```hcl
data "aws_s3_bucket" "important-bucket" {
  bucket = "important-bucket"
}

resource "aws_cloudtrail" "example" {
  # ... other configuration ...

  event_selector {
    read_write_type           = "All"
    include_management_events = true

    data_resource {
      type = "AWS::S3::Object"

      # Make sure to append a trailing '/' to your ARN if you want
      # to monitor all objects in a bucket.
      values = ["${data.aws_s3_bucket.important-bucket.arn}/"]
    }
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Specifies the name of the trail.
* `s3_bucket_name` - (Required) Specifies the name of the S3 bucket designated for publishing log files.
* `s3_key_prefix` - (Optional) Specifies the S3 key prefix that precedes
    the name of the bucket you have designated for log file delivery.
* `cloud_watch_logs_role_arn` - (Optional) Specifies the role for the CloudWatch Logs
    endpoint to assume to write to a user’s log group.
* `cloud_watch_logs_group_arn` - (Optional) Specifies a log group name using an Amazon Resource Name (ARN),
    that represents the log group to which CloudTrail logs will be delivered.
* `enable_logging` - (Optional) Enables logging for the trail. Defaults to `true`.
    Setting this to `false` will pause logging.
* `include_global_service_events` - (Optional) Specifies whether the trail is publishing events
    from global services such as IAM to the log files. Defaults to `true`.
* `is_multi_region_trail` - (Optional) Specifies whether the trail is created in the current
    region or in all regions. Defaults to `false`.
* `is_organization_trail` - (Optional) Specifies whether the trail is an AWS Organizations trail. Organization trails log events for the master account and all member accounts. Can only be created in the organization master account. Defaults to `false`.
* `sns_topic_name` - (Optional) Specifies the name of the Amazon SNS topic
    defined for notification of log file delivery.
* `enable_log_file_validation` - (Optional) Specifies whether log file integrity validation is enabled.
    Defaults to `false`.
* `kms_key_id` - (Optional) Specifies the KMS key ARN to use to encrypt the logs delivered by CloudTrail.
* `event_selector` - (Optional) Specifies an event selector for enabling data event logging. Fields documented below. Please note the [CloudTrail limits](https://docs.aws.amazon.com/awscloudtrail/latest/userguide/WhatIsCloudTrail-Limits.html) when configuring these.
* `tags` - (Optional) A mapping of tags to assign to the trail

### Event Selector Arguments
For **event_selector** the following attributes are supported.

* `read_write_type` (Optional) - Specify if you want your trail to log read-only events, write-only events, or all. By default, the value is All. You can specify only the following value: "ReadOnly", "WriteOnly", "All". Defaults to `All`.
* `include_management_events` (Optional) - Specify if you want your event selector to include management events for your trail.
* `data_resource` (Optional) - Specifies logging data events. Fields documented below.

#### Data Resource Arguments
For **data_resource** the following attributes are supported.

* `type` (Required) - The resource type in which you want to log data events. You can specify only the follwing value: "AWS::S3::Object", "AWS::Lambda::Function"
* `values` (Required) - A list of ARN for the specified S3 buckets and object prefixes..

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The name of the trail.
* `home_region` - The region in which the trail was created.
* `arn` - The Amazon Resource Name of the trail.


## Import

Cloudtrails can be imported using the `name`, e.g.

```
$ terraform import aws_cloudtrail.sample my-sample-trail
```
