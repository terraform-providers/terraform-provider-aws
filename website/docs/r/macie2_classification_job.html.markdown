---
subcategory: "Macie"
layout: "aws"
page_title: "AWS: aws_macie2_classification_job"
description: |-
  Provides a resource to manage an AWS Macie Classification Job.
---

# Resource: aws_macie2_classification_job

Provides a resource to manage an [AWS Macie Classification Job](https://docs.aws.amazon.com/macie/latest/APIReference/jobs.html).

## Example Usage

```terraform
resource "aws_macie2_account" "test" {}

resource "aws_macie2_classification_job" "test" {
  job_type = "ONE_TIME"
  name     = "NAME OF THE CLASSIFICATION JOB"
  s3_job_definition {
    bucket_definitions {
      account_id = "ACCOUNT ID"
      buckets    = ["S3 BUCKET NAME"]
    }
  }
  depends_on = [aws_macie2_account.test]
}
```

## Argument Reference

The following arguments are supported:

* `schedule_frequency` -  (Optional) The recurrence pattern for running the job. To run the job only once, don't specify a value for this property and set the value for the `job_type` property to `ONE_TIME`. (documented below)
* `custom_data_identifier_ids` -  (Optional) The custom data identifiers to use for data analysis and classification.
* `sampling_percentage` -  (Optional) The sampling depth, as a percentage, to apply when processing objects. This value determines the percentage of eligible objects that the job analyzes. If this value is less than 100, Amazon Macie selects the objects to analyze at random, up to the specified percentage, and analyzes all the data in those objects.
* `name` -  (Optional) A custom name for the job. The name can contain as many as 500 characters. If omitted, Terraform will assign a random, unique name. Conflicts with `name_prefix`.
* `name_prefix` -  (Optional) Creates a unique name beginning with the specified prefix. Conflicts with `name`.
* `description` -  (Optional) A custom description of the job. The description can contain as many as 200 characters.
* `initial_run` -  (Optional) Specifies whether to analyze all existing, eligible objects immediately after the job is created.
* `job_type` -  (Required) The schedule for running the job. Valid values are: `ONE_TIME` - Run the job only once. If you specify this value, don't specify a value for the `schedule_frequency` property. `SCHEDULED` - Run the job on a daily, weekly, or monthly basis. If you specify this value, use the `schedule_frequency` property to define the recurrence pattern for the job.
* `s3_job_definition` -  (Optional) The S3 buckets that contain the objects to analyze, and the scope of that analysis. (documented below)
* `tags` -  (Optional) A map of key-value pairs that specifies the tags to associate with the job. A job can have a maximum of 50 tags. Each tag consists of a tag key and an associated tag value. The maximum length of a tag key is 128 characters. The maximum length of a tag value is 256 characters.
* `job_status` -  (Optional) The status for the job. Valid values are: `CANCELLED`, `RUNNING` and `USER_PAUSED`

The `schedule_frequency` object supports the following:

* `daily_schedule` -  (Optional) Specifies a daily recurrence pattern for running the job.
* `weekly_schedule` -  (Optional) Specifies a weekly recurrence pattern for running the job.
* `monthly_schedule` -  (Optional) Specifies a monthly recurrence pattern for running the job.

The `s3_job_definition` object supports the following:

* `bucket_definitions` -  (Optional) An array of objects, one for each AWS account that owns buckets to analyze. Each object specifies the account ID for an account and one or more buckets to analyze for the account. (documented below)
* `scoping` -  (Optional) The property- and tag-based conditions that determine which objects to include or exclude from the analysis. (documented below)

The `bucket_definitions` object supports the following:

* `account_id` -  (Required) The unique identifier for the AWS account that owns the buckets.
* `buckets` -  (Required) An array that lists the names of the buckets.

The `scoping` object supports the following:

* `excludes` -  (Optional) The property- or tag-based conditions that determine which objects to exclude from the analysis. (documented below)
* `includes` -  (Optional) The property- or tag-based conditions that determine which objects to include in the analysis. (documented below)

The `excludes` and `includes` object supports the following:

* `and` -  (Optional) An array of conditions, one for each condition that determines which objects to include or exclude from the job. (documented below)

The `and` object supports the following:

* `simple_scope_term` -  (Optional) A property-based condition that defines a property, operator, and one or more values for including or excluding an object from the job.  (documented below)
* `tag_scope_term` -  (Optional) A tag-based condition that defines the operator and tag keys or tag key and value pairs for including or excluding an object from the job.  (documented below)

The `simple_scope_term` object supports the following:

* `comparator` -  (Optional) The operator to use in a condition. Valid values are: `EQ`, `GT`, `GTE`, `LT`, `LTE`, `NE`, `CONTAINS`, `STARTS_WITH`
* `values` -  (Optional) An array that lists the values to use in the condition.
* `key` -  (Optional) The object property to use in the condition.

The `tag_scope_term` object supports the following:

* `comparator` -  (Optional) The operator to use in the condition.
* `tag_values` -  (Optional) The tag keys or tag key and value pairs to use in the condition.
* `key` -  (Optional) The tag key to use in the condition.
* `target` -  (Optional) The type of object to apply the condition to.


## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The unique identifier (ID) of the macie classification job.
* `created_at` -  The date and time, in UTC and extended RFC 3339 format, when the job was created.
* `user_paused_details` - If the current status of the job is `USER_PAUSED`, specifies when the job was paused and when the job or job run will expire and be cancelled if it isn't resumed. This value is present only if the value for `job-status` is `USER_PAUSED`.

## Import

`aws_macie2_classification_job` can be imported using the id, e.g.

```
$ terraform import aws_macie2_classification_job.example abcd1
```
