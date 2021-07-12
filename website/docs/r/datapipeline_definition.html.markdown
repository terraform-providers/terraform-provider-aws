---
layout: "aws"
page_title: "AWS: aws_datapipeline_definition"
sidebar_current: "docs-aws-resource-datapipeline-definition"
description: |-
  Provides a AWS DataPipeline Definition.
---

# Resource: aws_datapipeline_definition

Provides a Data Pipeline Definition resource.

## Example Usage

```hcl
resource "aws_iam_role" "role" {
  name = "tf-test-datapipeline-role-%[1]s"
      
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": [
          "elasticmapreduce.amazonaws.com",
          "datapipeline.amazonaws.com"
        ]
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}
data "aws_iam_policy" "role" {
  arn = "arn:aws:iam::aws:policy/service-role/AWSDataPipelineRole"
}
resource "aws_iam_role_policy_attachment" "role" {
  role       = "${aws_iam_role.role.name}"
  policy_arn = "${data.aws_iam_policy.role.arn}"
}
resource "aws_iam_role" "resource_role" {
  name = "tf-test-datapipeline-resource-role-%[1]s"
      
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": [
          "ec2.amazonaws.com"
        ]
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}
data "aws_iam_policy" "resource_role" {
  arn = "arn:aws:iam::aws:policy/service-role/AWSDataPipelineRole"
}
resource "aws_iam_role_policy_attachment" "resource_role" {
  role       = "${aws_iam_role.resource_role.name}"
  policy_arn = "${data.aws_iam_policy.resource_role.arn}"
}
resource "aws_iam_instance_profile" "resource_role" {
  name = "tf-test-datapipeline-resource-role-profile-%[1]s"
  role = "${aws_iam_role.resource_role.name}"
}
resource "aws_datapipeline_pipeline" "default" {
	name = "tf-pipeline-default"
}
resource "aws_datapipeline_definition" "default" {
  pipeline_id = "${aws_datapipeline_pipeline.default.id}"
  default {
	schedule_type = "ondemand"
	role          = "${aws_iam_role.role.arn}"
	resource_role = "${aws_iam_instance_profile.resource_role.arn}"
  }
}
```

## Argument Reference

The following arguments are supported:

* `pipeline_id` - (Required) The id of Pipeline.
* `default` - (Required) The default configuration to assign to the resource. Documented below.

The `default` configuration supports the following:

* `resource_role` - (Required) The arn of iam instance profile attached to resources.
* `role` - (Required) The arn of iam role, execute pipeline.
* `schedule_type` - (Required) The parameters for the RUN_COMMAND task execution. Valid values are `cron`, `ondemand` and `none`. Default to `none`.
* `failure_and_rerun_mode` - (Optional) The date and time to start the scheduled runs. Valid values are `cascade` and `timeseries`.
* `pipeline_log_uri` - (Optional) The s3 bucket uri for default to output logs. Example: `s3://<bucket-name>/<prefix>`.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The identifier of the pipeline.

## Import

`aws_datapipeline_definition` can be imported by using the id (Pipeline ID), e.g.

```
$ terraform import aws_datapipeline_definition.default df-1234567890
```