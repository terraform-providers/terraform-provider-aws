---
subcategory: "Cost and Usage Report"
layout: "aws"
page_title: "AWS: aws_cur_report_definition"
description: |-
  Provides a Cost and Usage Report Definition.
---

# Resource: aws_cur_report_definition

Manages Cost and Usage Report Definitions.

~> *NOTE:* The AWS Cost and Usage Report service is only available in `us-east-1` currently.

~> *NOTE:* If AWS Organizations is enabled, only the master account can use this resource.

## Example Usage

```hcl
resource "aws_cur_report_definition" "example_cur_report_definition" {
  report_name                = "example-cur-report-definition"
  time_unit                  = "HOURLY"
  format                     = "textORcsv"
  compression                = "GZIP"
  additional_schema_elements = ["RESOURCES"]
  s3_bucket                  = "example-bucket-name"
  s3_region                  = "us-east-1"
  additional_artifacts       = ["REDSHIFT", "QUICKSIGHT"]
}
```

## Argument Reference

The following arguments are supported:

* `report_name` - (Required) Unique name for the report. Must start with a number/letter and is case sensitive. Limited to 256 characters.
* `time_unit` - (Required) The frequency on which report data are measured and displayed.  Valid values are: HOURLY, DAILY.
* `format` - (Required) Format for report. Valid values are: textORcsv.
* `compression` - (Required) Compression format for report. Valid values are: GZIP, ZIP.
* `additional_schema_elements` - (Required) A list of schema elements. Valid values are: RESOURCES.
* `s3_bucket` - (Required) Name of the existing S3 bucket to hold generated reports.
* `s3_prefix` - (Optional) Report path prefix. Limited to 256 characters.
* `s3_region` - (Required) Region of the existing S3 bucket to hold generated reports.
* `additional_artifacts` - (Required)  A list of additional artifacts. Valid values are: REDSHIFT, QUICKSIGHT.

## Import

Report Definitions can be imported using the `report_name`, e.g.

```
$ terraform import aws_cur_report_definition.example_cur_report_definition example-cur-report-definition
```
