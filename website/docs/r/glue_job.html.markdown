---
layout: "aws"
page_title: "AWS: aws_glue_job"
sidebar_current: "docs-aws-resource-glue-job"
description: |-
  Provides an Glue Job resource.
---

# aws_glue_job

Provides a Glue Job resource.

## Example Usage

### Python Job

```hcl
resource "aws_glue_job" "example" {
  name     = "example"
  role_arn = "${aws_iam_role.example.arn}"

  command {
    script_location = "s3://${aws_s3_bucket.example.bucket}/example.py"
  }
}
```

### Scala Job

```hcl
resource "aws_glue_job" "example" {
  name     = "example"
  role_arn = "${aws_iam_role.example.arn}"

  command {
    script_location = "s3://${aws_s3_bucket.example.bucket}/example.scala"
  }

  default_arguments = {
    "--job-language" = "scala"
  }
}
```

## Argument Reference

The following arguments are supported:

* `allocated_capacity` – (Optional) The number of AWS Glue data processing units (DPUs) to allocate to this Job. At least 2 DPUs need to be allocated; the default is 10. A DPU is a relative measure of processing power that consists of 4 vCPUs of compute capacity and 16 GB of memory.
* `command` – (Required) The command of the job. Defined below.
* `connections` – (Optional) The list of connections used for this job.
* `default_arguments` – (Optional) The map of default arguments for this job. You can specify arguments here that your own job-execution script consumes, as well as arguments that AWS Glue itself consumes. For information about how to specify and consume your own Job arguments, see the [Calling AWS Glue APIs in Python](http://docs.aws.amazon.com/glue/latest/dg/aws-glue-programming-python-calling.html) topic in the developer guide. For information about the key-value pairs that AWS Glue consumes to set up your job, see the [Special Parameters Used by AWS Glue](http://docs.aws.amazon.com/glue/latest/dg/aws-glue-programming-python-glue-arguments.html) topic in the developer guide.
* `description` – (Optional) Description of the job.
* `execution_property` – (Optional) Execution property of the job. Defined below.
* `max_retries` – (Optional) The maximum number of times to retry this job if it fails.
* `name` – (Required) The name you assign to this job. It must be unique in your account.
* `role_arn` – (Required) The ARN of the IAM role associated with this job.
* `timeout` – (Optional) The job timeout in minutes. The default is 2880 minutes (48 hours).
* `security_configuration` - (Optional) The name of the Security Configuration to be associated with the job. 

### command Argument Reference

* `name` - (Optional) The name of the job command. Defaults to `glueetl`
* `script_location` - (Required) Specifies the S3 path to a script that executes a job.

### execution_property Argument Reference

* `max_concurrent_runs` - (Optional) The maximum number of concurrent runs allowed for a job. The default is 1.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - Job name

## Import

Glue Jobs can be imported using `name`, e.g.

```
$ terraform import aws_glue_job.MyJob MyJob
```
