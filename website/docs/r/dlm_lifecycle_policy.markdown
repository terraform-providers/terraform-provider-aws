---
subcategory: "Data Lifecycle Manager (DLM)"
layout: "aws"
page_title: "AWS: aws_dlm_lifecycle_policy"
description: |-
  Provides a Data Lifecycle Manager (DLM) lifecycle policy for managing snapshots.
---

# Resource: aws_dlm_lifecycle_policy

Provides a [Data Lifecycle Manager (DLM) lifecycle policy](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/snapshot-lifecycle.html) for managing snapshots.

## Example Usage

```hcl
resource "aws_iam_role" "dlm_lifecycle_role" {
  name = "dlm-lifecycle-role"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "dlm.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "dlm_lifecycle" {
  name = "dlm-lifecycle-policy"
  role = "${aws_iam_role.dlm_lifecycle_role.id}"

  policy = <<EOF
{
   "Version": "2012-10-17",
   "Statement": [
      {
         "Effect": "Allow",
         "Action": [
            "ec2:CreateSnapshot",
            "ec2:DeleteSnapshot",
            "ec2:DescribeVolumes",
            "ec2:DescribeSnapshots"
         ],
         "Resource": "*"
      },
      {
         "Effect": "Allow",
         "Action": [
            "ec2:CreateTags"
         ],
         "Resource": "arn:aws:ec2:*::snapshot/*"
      }
   ]
}
EOF
}

resource "aws_kms_key" "dlm_cross_region_copy_cmk" {
  description = "Terraform %[1]s"

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Id": "dlm-cross-region-copy-cmk",
  "Statement": [
    {
      "Sid": "Enable IAM User Permissions",
      "Effect": "Allow",
      "Principal": {
        "AWS": "*"
      },
      "Action": "kms:*",
      "Resource": "*"
    }
  ]
}
POLICY
}

resource "aws_dlm_lifecycle_policy" "example" {
  description        = "example DLM lifecycle policy"
  execution_role_arn = "${aws_iam_role.dlm_lifecycle_role.arn}"
  state              = "ENABLED"

  policy_details {
    resource_types = ["VOLUME"]

    schedule {
      name = "2 weeks of daily snapshots"

      create_rule {
        interval      = 24
        interval_unit = "HOURS"
        times         = ["23:45"]
      }

      retain_rule {
        count = 14
      }

      tags_to_add = {
        SnapshotCreator = "DLM"
      }

      copy_tags = false

      cross_region_copy_rule {
        target_region = "us-west-2"
        encrypted     = true
        cmk_arn       = "${aws_kms_key.dlm_cross_region_copy_cmk.arn}"
        copy_tags     = true
        retain_rule {
          interval      = 30
          interval_unit = "DAYS"
        }
      }
    }

    target_tags = {
      Snapshot = "true"
    }
  }
}
```

## Argument Reference

The following arguments are supported:

* `description` - (Required) A description for the DLM lifecycle policy.
* `execution_role_arn` - (Required) The ARN of an IAM role that is able to be assumed by the DLM service.
* `policy_details` - (Required) See the [`policy_details` configuration](#policy-details-arguments) block. Max of 1.
* `state` - (Optional) Whether the lifecycle policy should be enabled or disabled. `ENABLED` or `DISABLED` are valid values. Defaults to `ENABLED`.
* `tags` - (Optional) Key-value map of resource tags.

#### Policy Details arguments

* `resource_types` - (Required) A list of resource types that should be targeted by the lifecycle policy. `VOLUME` is currently the only allowed value.
* `schedule` - (Required) See the [`schedule` configuration](#schedule-arguments) block.
* `target_tags` (Required) A map of tag keys and their values. Any resources that match the `resource_types` and are tagged with _any_ of these tags will be targeted.

~> Note: You cannot have overlapping lifecycle policies that share the same `target_tags`. Terraform is unable to detect this at plan time but it will fail during apply.

#### Schedule arguments

* `copy_tags` - (Optional) Copy all user-defined tags on a source volume to snapshots of the volume created by this policy.
* `create_rule` - (Required) See the [`create_rule`](#create-rule-arguments) block. Max of 1 per schedule.
* `cross_region_copy_rule` (Optional) - See the [`cross_region_copy_rule`](#cross-region-copy-rule-arguments) block.
* `name` - (Required) A name for the schedule.
* `retain_rule` - (Required) See the [`retain_rule`](#retain-rule-arguments) block. Max of 1 per schedule.
* `tags_to_add` - (Optional) A map of tag keys and their values. DLM lifecycle policies will already tag the snapshot with the tags on the volume. This configuration adds extra tags on top of these.


#### Create Rule arguments

* `interval` - (Required) How often this lifecycle policy should be evaluated. `1`, `2`,`3`,`4`,`6`,`8`,`12` or `24` are valid values.
* `interval_unit` - (Optional) The unit for how often the lifecycle policy should be evaluated. `HOURS` is currently the only allowed value and also the default value.
* `times` - (Optional) A list of times in 24 hour clock format that sets when the lifecycle policy should be evaluated. Max of 1.

#### Retain Rule arguments

* `count` - (Required) How many snapshots to keep. Must be an integer between 1 and 1000.

#### Cross Region Copy Rule arguments

* `target_region` - (Required) The target AWS region.
* `encrypted` - (Required) To encrypt a copy of an unencrypted snapshot if encryption by default is not enabled, enable encryption using this parameter. Copies of encrypted snapshots are encrypted, even if this parameter is false or if encryption by default is not enabled.
* `cmk_arn` - (Optional) The Amazon Resource Name (ARN) of the AWS KMS customer master key (CMK) to use for EBS encryption.
* `copy_tags` - (Optional) Copy all user-defined tags from the source snapshot to the copied snapshot.
* `retain_rule` - (Required) See the [`cross_region_copy_retain_rule`](#cross-region-copy-retain-rule-arguments) block. Max of 1 per schedule.

#### Cross Region Copy Retain Rule arguments

* `interval` - (Required) The amount of time to retain each snapshot. The maximum is 100 years.
* `interval_unit` - (Required) The unit of time for time-based retention. `DAYS`, `WEEKS`, `MONTHS`, `YEARS` are currently the only allowed values.


## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - Amazon Resource Name (ARN) of the DLM Lifecycle Policy.
* `id` - Identifier of the DLM Lifecycle Policy.

## Import

DLM lifecyle policies can be imported by their policy ID:

```
$ terraform import aws_dlm_lifecycle_policy.example policy-abcdef12345678901
```
