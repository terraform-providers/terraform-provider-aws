---
subcategory: "GuardDuty"
layout: aws
page_title: 'AWS: aws_guardduty_ipset'
description: Provides a resource to manage a GuardDuty IPSet
---

# Resource: aws_guardduty_ipset

Provides a resource to manage a GuardDuty IPSet.

~> **Note:** Currently in GuardDuty, users from member accounts cannot upload and further manage IPSets. IPSets that are uploaded by the primary account are imposed on GuardDuty functionality in its member accounts. See the [GuardDuty API Documentation](https://docs.aws.amazon.com/guardduty/latest/ug/create-ip-set.html)

## Example Usage

```hcl
resource "aws_guardduty_ipset" "example" {
  activate    = true
  detector_id = aws_guardduty_detector.primary.id
  format      = "TXT"
  location    = "https://s3.amazonaws.com/${aws_s3_bucket_object.MyIPSet.bucket}/${aws_s3_bucket_object.MyIPSet.key}"
  name        = "MyIPSet"
}

resource "aws_guardduty_detector" "primary" {
  enable = true
}

resource "aws_s3_bucket" "bucket" {
  acl = "private"
}

resource "aws_s3_bucket_object" "MyIPSet" {
  acl     = "public-read"
  content = "10.0.0.0/8\n"
  bucket  = aws_s3_bucket.bucket.id
  key     = "MyIPSet"
}
```

## Argument Reference

The following arguments are supported:

* `activate` - (Required) Specifies whether GuardDuty is to start using the uploaded IPSet.
* `detector_id` - (Required) The detector ID of the GuardDuty.
* `format` - (Required) The format of the file that contains the IPSet. Valid values: `TXT` | `STIX` | `OTX_CSV` | `ALIEN_VAULT` | `PROOF_POINT` | `FIRE_EYE`
* `location` - (Required) The URI of the file that contains the IPSet.
* `name` - (Required) The friendly name to identify the IPSet.
* `tags` - (Optional) Key-value map of resource tags.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - Amazon Resource Name (ARN) of the GuardDuty IPSet.
* `id` - The ID of the GuardDuty IPSet.

## Import

GuardDuty IPSet can be imported using the the primary GuardDuty detector ID and IPSet ID, e.g.

```
$ terraform import aws_guardduty_ipset.MyIPSet 00b00fd5aecc0ab60a708659477e9617:123456789012
```
