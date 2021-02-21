---
subcategory: "SSM"
layout: "aws"
page_title: "AWS: aws_ssm_association"
description: |-
  Associates an SSM Document to an instance or EC2 tag.
---

# Resource: aws_ssm_association

Associates an SSM Document to an instance or EC2 tag.

## Example Usage

```hcl
resource "aws_ssm_association" "example" {
  name = aws_ssm_document.example.name

  targets {
    key    = "InstanceIds"
    values = [aws_instance.example.id]
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the SSM document to apply.
* `apply_only_at_cron_interval` - (Optional) By default, when you create a new or update associations, the system runs it immediately and then according to the schedule you specified. Enable this option if you do not want an association to run immediately after you create or update it. This parameter is not supported for rate expressions. Default: `false`.
* `association_name` - (Optional) The descriptive name for the association.
* `document_version` - (Optional) The document version you want to associate with the target(s). Can be a specific version or the default version.
* `instance_id` - (Optional) The instance ID to apply an SSM document to. Use `targets` with key `InstanceIds` for document schema versions 2.0 and above.
* `output_location` - (Optional) An output location block. Output Location is documented below.
* `parameters` - (Optional) A block of arbitrary string parameters to pass to the SSM document.
* `schedule_expression` - (Optional) A cron expression when the association will be applied to the target(s).
* `targets` - (Optional) A block containing the targets of the SSM association. Targets are documented below. AWS currently supports a maximum of 5 targets.
* `compliance_severity` - (Optional) The compliance severity for the association. Can be one of the following: `UNSPECIFIED`, `LOW`, `MEDIUM`, `HIGH` or `CRITICAL`
* `max_concurrency` - (Optional) The maximum number of targets allowed to run the association at the same time. You can specify a number, for example 10, or a percentage of the target set, for example 10%.
* `max_errors` - (Optional) The number of errors that are allowed before the system stops sending requests to run the association on additional targets. You can specify a number, for example 10, or a percentage of the target set, for example 10%.
* `automation_target_parameter_name` - (Optional) Specify the target for the association. This target is required for associations that use an `Automation` document and target resources by using rate controls.

Output Location (`output_location`) is an S3 bucket where you want to store the results of this association:

* `s3_bucket_name` - (Required) The S3 bucket name.
* `s3_key_prefix` - (Optional) The S3 bucket prefix. Results stored in the root if not configured.

Targets specify what instance IDs or tags to apply the document to and has these keys:

* `key` - (Required) Either `InstanceIds` or `tag:Tag Name` to specify an EC2 tag.
* `values` - (Required) A list of instance IDs or tag values. AWS currently limits this list size to one value.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `association_id` - The ID of the SSM association.
* `instance_id` - The instance id that the SSM document was applied to.
* `name` - The name of the SSM document to apply.
* `parameters` - Additional parameters passed to the SSM document.

## Import

SSM associations can be imported using the `association_id`, e.g.

```
$ terraform import aws_ssm_association.test-association 10abcdef-0abc-1234-5678-90abcdef123456
```
