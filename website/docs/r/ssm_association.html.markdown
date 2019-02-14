---
layout: "aws"
page_title: "AWS: aws_ssm_association"
sidebar_current: "docs-aws-resource-ssm-association"
description: |-
  Associates an SSM Document to an instance or EC2 tag.
---

# aws_ssm_association

Associates an SSM Document to an instance or EC2 tag.

## Example Usage

```hcl
resource "aws_ssm_association" "example" {
  name = "${aws_ssm_document.example.name}"

  targets {
    key    = "InstanceIds"
    values = "${aws_instance.example.id}"
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the SSM document to apply.
* `association_name` - (Optional) The descriptive name for the association.
* `document_version` - (Optional) The document version you want to associate with the target(s). Can be a specific version or the default version.
* `instance_id` - (Optional) The instance ID to apply an SSM document to. Use `targets` with key `InstanceIds` for document schema versions 2.0 and above.
* `output_location` - (Optional) An output location block. Output Location is documented below.
* `parameters` - (Optional) A block of arbitrary string parameters to pass to the SSM document.
* `schedule_expression` - (Optional) A cron expression when the association will be applied to the target(s).
* `targets` - (Optional) A block containing the targets of the SSM association. Targets are documented below. AWS currently supports a maximum of 5 targets.

Output Location (`output_location`) is an S3 bucket where you want to store the results of this association:

* `s3_bucket_name` - (Required) The S3 bucket name.
* `s3_key_prefix` - (Optional) The S3 bucket prefix. Results stored in the root if not configured.

Targets specify what instance IDs or tags to apply the document to and has these keys:

* `key` - (Required) Either `InstanceIds` or `tag:Tag Name` to specify an EC2 tag.
* `values` - (Required) A list of instance IDs or tag values. AWS currently limits this to 1 target value.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `name` - The name of the SSM document to apply.
* `instance_ids` - The instance id that the SSM document was applied to.
* `parameters` - Additional parameters passed to the SSM document.
