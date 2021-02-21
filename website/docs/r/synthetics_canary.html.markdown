---
subcategory: "Synthetics"
layout: "aws"
page_title: "AWS: aws_synthetics_canary"
description: |-
  Provides a Synthetics Canary resource
---

# Resource: aws_synthetics_canary

Provides a Synthetics Canary resource.

~> **NOTE:** When you create a canary, AWS creates supporting implicit resources. See the Amazon CloudWatch Synthetics documentation on [DeleteCanary](https://docs.aws.amazon.com/AmazonSynthetics/latest/APIReference/API_DeleteCanary.html) for a full list. Neither AWS nor Terraform deletes these implicit resources automatically when the canary is deleted. Before deleting a canary, ensure you have all the information about the canary that you need to delete the implicit resources using Terraform shell commands, the AWS Console, or AWS CLI.

## Example Usage

```hcl
resource "aws_synthetics_canary" "some" {
  name                 = "some-canary"
  artifact_s3_location = "s3://some-bucket/"
  execution_role_arn   = "some-role"
  handler              = "exports.handler"
  zip_file             = "test-fixtures/lambdatest.zip"
  runtime_version      = "syn-1.0"

  schedule {
    expression = "rate(0 minute)"
  }
}
```

## Argument Reference

The following arguments are required:

* `artifact_s3_location` - (Required) Location in Amazon S3 where Synthetics stores artifacts from the test runs of this canary.
* `execution_role_arn` - (Required) ARN of the IAM role to be used to run the canary. see [AWS Docs](https://docs.aws.amazon.com/AmazonSynthetics/latest/APIReference/API_CreateCanary.html#API_CreateCanary_RequestSyntax) for permissions needs for IAM Role.
* `handler` - (Required) Entry point to use for the source code when running the canary. This value must end with the string `.handler` .
* `name` - (Required) Name for this canary.
* `runtime_version` - (Required) Runtime version to use for the canary. Versions change often so consult the [Amazon CloudWatch documentation](https://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/CloudWatch_Synthetics_Canaries_Library.html) for the latest valid versions. Values include `syn-python-selenium-1.0`, `syn-nodejs-puppeteer-3.0`, `syn-nodejs-2.2`, `syn-nodejs-2.1`, `syn-nodejs-2.0`, and `syn-1.0`.
* `schedule` -  (Required) Configuration block providing how often the canary is to run and when these test runs are to stop. Detailed below.

The following arguments are optional:

* `failure_retention_period` - (Optional) Number of days to retain data about failed runs of this canary. If you omit this field, the default of 31 days is used. The valid range is 1 to 455 days.
* `run_config` - (Optional) Configuration block for individual canary runs. Detailed below.
* `s3_bucket` - (Optional) Full bucket name which is used if your canary script is located in S3. The bucket must already exist. Specify the full bucket name including s3:// as the start of the bucket name. **Conflicts with `zip_file`.**
* `s3_key` - (Optional) S3 key of your script. **Conflicts with `zip_file`.**
* `s3_version` - (Optional) S3 version ID of your script. **Conflicts with `zip_file`.**
* `start_canary` - (Optional) Whether to run or stop the canary.
* `success_retention_period` - (Optional) Number of days to retain data about successful runs of this canary. If you omit this field, the default of 31 days is used. The valid range is 1 to 455 days.
* `tags` - (Optional) Key-value map of resource tags
* `vpc_config` - (Optional) Configuration block. Detailed below.
* `zip_file` - (Optional) ZIP file that contains the script, if you input your canary script directly into the canary instead of referring to an S3 location. It can be up to 5 MB. **Conflicts with `s3_bucket`, `s3_key`, and `s3_version`.**

### schedule

* `expression` - (Required) Rate expression that defines how often the canary is to run. The syntax is rate(number unit). unit can be minute, minutes, or hour.
* `duration_in_seconds` - (Optional) Duration in seconds, for the canary to continue making regular runs according to the schedule in the Expression value.

### run_config

* `timeout_in_seconds` - (Optional) Number of seconds the canary is allowed to run before it must stop. If you omit this field, the frequency of the canary is used, up to a maximum of 840 (14 minutes).
* `memory_in_mb` - (Optional) Maximum amount of memory available to the canary while it is running, in MB. The value you specify must be a multiple of 64.
* `active_tracing` - (Optional) Whether this canary is to use active AWS X-Ray tracing when it runs. You can enable active tracing only for canaries that use version syn-nodejs-2.0 or later for their canary runtime.

### vpc_config

If this canary tests an endpoint in a VPC, this structure contains information about the subnet and security groups of the VPC endpoint. For more information, see [Running a Canary in a VPC](https://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/CloudWatch_Synthetics_Canaries_VPC.html).

* `subnet_ids` - (Required) IDs of the subnets where this canary is to run.
* `security_group_ids` - (Required) IDs of the security groups for this canary.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - Amazon Resource Name (ARN) of the Canary.
* `engine_arn` - ARN of the Lambda function that is used as your canary's engine.
* `id` - Name for this canary.
* `source_location_arn` - ARN of the Lambda layer where Synthetics stores the canary script code.
* `status` - Canary status.
* `timeline` - Structure that contains information about when the canary was created, modified, and most recently run. see [Timeline](#timeline).

### vpc_config

* `vpc_id` - ID of the VPC where this canary is to run.

### timeline

* `created` - Date and time the canary was created.
* `last_modified` - Date and time the canary was most recently modified.
* `last_started` - Date and time that the canary's most recent run started.
* `last_stopped` - Date and time that the canary's most recent run ended.

## Import

Synthetics Canaries can be imported using the `name`, e.g.

```
$ terraform import aws_synthetics_canary.some some-canary
```
