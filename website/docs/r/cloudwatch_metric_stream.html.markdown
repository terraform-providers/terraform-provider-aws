---
subcategory: "CloudWatch"
layout: "aws"
page_title: "AWS: aws_cloudwatch_metric_stream"
description: |-
  Provides a CloudWatch Metric Stream resource.
---

# Resource: aws_cloudwatch_metric_stream

Provides a CloudWatch Metric Stream resource.

## Example Usage

```terraform
resource "aws_cloudwatch_metric_stream" "main" {
  name          = "my-metric-stream"
  role_arn      = aws_iam_role.metric_stream_to_firehose.arn
  firehose_arn  = aws_kinesis_firehose_delivery_stream.s3_stream.arn
  output_format = "json"

  include_filter {
    namespace = "AWS/EC2"
  }

  include_filter {
    namespace = "AWS/EBS"
  }
}

# https://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/CloudWatch-metric-streams-trustpolicy.html
resource "aws_iam_role" "metric_stream_to_firehose" {
  name = "metric_stream_to_firehose_role"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "streams.metrics.cloudwatch.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

# https://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/CloudWatch-metric-streams-trustpolicy.html
resource "aws_iam_role_policy" "metric_stream_to_firehose" {
  name = "default"
  role = aws_iam_role.metric_stream_to_firehose.id

  policy = <<EOF
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "firehose:PutRecord",
                "firehose:PutRecordBatch"
            ],
            "Resource": "${aws_kinesis_firehose_delivery_stream.s3_stream.arn}"
        }
    ]
}
EOF
}

resource "aws_s3_bucket" "bucket" {
  bucket = "metric-stream-test-bucket"
  acl    = "private"
}

resource "aws_iam_role" "firehose_to_s3" {
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "firehose.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "firehose_to_s3" {
  name = "default"
  role = aws_iam_role.firehose_to_s3.id

  policy = <<EOF
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "s3:AbortMultipartUpload",
                "s3:GetBucketLocation",
                "s3:GetObject",
                "s3:ListBucket",
                "s3:ListBucketMultipartUploads",
                "s3:PutObject"
            ],      
            "Resource": [        
                "${aws_s3_bucket.bucket.arn}",
                "${aws_s3_bucket.bucket.arn}/*"		    
            ]
        }
    ]
}
EOF
}

resource "aws_kinesis_firehose_delivery_stream" "s3_stream" {
  name        = "metric-stream-test-stream"
  destination = "s3"

  s3_configuration {
    role_arn   = aws_iam_role.firehose_to_s3.arn
    bucket_arn = aws_s3_bucket.bucket.arn
  }
}
```

## Argument Reference

The following arguments are required:

* `firehose_arn` - (Required) ARN of the Amazon Kinesis Firehose delivery stream to use for this metric stream.
* `role_arn` - (Required) ARN of the IAM role that this metric stream will use to access Amazon Kinesis Firehose resources. For more information about role permissions, see [Trust between CloudWatch and Kinesis Data Firehose](https://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/CloudWatch-metric-streams-trustpolicy.html).
* `output_format` - (Required) Output format for the stream. For more information about output formats, see [Metric streams output formats](https://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/CloudWatch-metric-streams-formats.html).
* `tags` - (Optional) Map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

The following arguments are optional:

* `exclude_filter` - (Optional) List of exclusive metric filters. If you specify this parameter, the stream sends metrics from all metric namespaces except for the namespaces that you specify here. Conflicts with `include_filter`.
* `include_filter` - (Optional) List of inclusive metric filters. If you specify this parameter, the stream sends only the metrics from the metric namespaces that you specify here. Conflicts with `exclude_filter`.
* `name` - (Optional, Forces new resource) Friendly name of the metric stream. If omitted, Terraform will assign a random, unique name. Conflicts with `name_prefix`.
* `name_prefix` - (Optional, Forces new resource) Creates a unique friendly name beginning with the specified prefix. Conflicts with `name`.

### `exclude_filter`

* `namespace` - (Required) Name of the metric namespace in the filter.

### `include_filter`

* `namespace` - (Required) Name of the metric namespace in the filter.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - ARN of the metric stream.
* `creation_date` - Date and time in [RFC3339 format](https://tools.ietf.org/html/rfc3339#section-5.8) that the metric stream was created.
* `last_update_date` - Date and time in [RFC3339 format](https://tools.ietf.org/html/rfc3339#section-5.8) that the metric stream was last updated.
* `state` - State of the metric stream. Possible values are `running` and `stopped`.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block).

## Import

CloudWatch metric streams can be imported using the `name`, e.g.

```
$ terraform import aws_cloudwatch_metric_stream.sample <name>
```
