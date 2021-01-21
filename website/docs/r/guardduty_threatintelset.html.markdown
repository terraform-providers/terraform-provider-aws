---
subcategory: "GuardDuty"
layout: aws
page_title: 'AWS: aws_guardduty_threatintelset'
description: Provides a resource to manage a GuardDuty ThreatIntelSet
---

# Resource: aws_guardduty_threatintelset

Provides a resource to manage a GuardDuty ThreatIntelSet.

~> **Note:** Currently in GuardDuty, users from member accounts cannot upload and further manage ThreatIntelSets. ThreatIntelSets that are uploaded by the primary account are imposed on GuardDuty functionality in its member accounts. See the [GuardDuty API Documentation](https://docs.aws.amazon.com/guardduty/latest/ug/create-threat-intel-set.html)

## Example Usage

```hcl
resource "aws_guardduty_detector" "primary" {
  enable = true
}

resource "aws_s3_bucket" "bucket" {
  acl = "private"
}

resource "aws_s3_bucket_object" "MyThreatIntelSet" {
  acl     = "public-read"
  content = "10.0.0.0/8\n"
  bucket  = aws_s3_bucket.bucket.id
  key     = "MyThreatIntelSet"
}

resource "aws_guardduty_threatintelset" "MyThreatIntelSet" {
  activate    = true
  detector_id = aws_guardduty_detector.primary.id
  format      = "TXT"
  location    = "https://s3.amazonaws.com/${aws_s3_bucket_object.MyThreatIntelSet.bucket}/${aws_s3_bucket_object.MyThreatIntelSet.key}"
  name        = "MyThreatIntelSet"
}
```

## Argument Reference

The following arguments are supported:

* `activate` - (Required) Specifies whether GuardDuty is to start using the uploaded ThreatIntelSet.
* `detector_id` - (Required) The detector ID of the GuardDuty.
* `format` - (Required) The format of the file that contains the ThreatIntelSet. Valid values: `TXT` | `STIX` | `OTX_CSV` | `ALIEN_VAULT` | `PROOF_POINT` | `FIRE_EYE`
* `location` - (Required) The URI of the file that contains the ThreatIntelSet.
* `name` - (Required) The friendly name to identify the ThreatIntelSet.
* `tags` - (Optional) Key-value map of resource tags.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - Amazon Resource Name (ARN) of the GuardDuty ThreatIntelSet.
* `id` - The ID of the GuardDuty ThreatIntelSet and the detector ID. Format: `<DetectorID>:<ThreatIntelSetID>`

## Import

GuardDuty ThreatIntelSet can be imported using the the primary GuardDuty detector ID and ThreatIntelSetID, e.g.

```
$ terraform import aws_guardduty_threatintelset.MyThreatIntelSet 00b00fd5aecc0ab60a708659477e9617:123456789012
```
