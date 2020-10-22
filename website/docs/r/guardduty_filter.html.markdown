---
layout: "aws"
subcategory: "GuardDuty"
page_title: "AWS: aws_guardduty_filter"
description: |-
  Provides a resource to manage a GuardDuty filter
---

# Resource: aws_guardduty_filter

Provides a resource to manage a GuardDuty filter.

## Example Usage

```hcl
resource "aws_guardduty_filter" "MyFilter" {
  name        = "MyFilter"
  action      = "ARCHIVE"
  detector_id = aws_guardduty_detector.example.id
  rank        = 1

  finding_criteria {
    criterion {
      field  = "region"
      equals = ["eu-west-1"]
    }

    criterion {
      field      = "service.additionalInfo.threatListName"
      not_equals = ["some-threat", "another-threat"]
    }

    criterion {
      field        = "updatedAt"
      greater_than = "2020-01-01T00:00:00Z"
      less_than    = "2020-02-01T00:00:00Z"
    }

    criterion {
      field                 = "severity"
      greater_than_or_equal = "4"
    }
  }
}
```

## Argument Reference

The following arguments are supported:

* `detector_id` - (Required) ID of a GuardDuty detector, attached to your account.
* `name` - (Required) The name of your filter.
* `description` - (Optional) Description of the filter.
* `rank` - (Required) Specifies the position of the filter in the list of current filters. Also specifies the order in which this filter is applied to the findings.
* `action` - (Required) Specifies the action that is to be applied to the findings that match the filter. Can be one of `ARCHIVE` or `NOOP`.
* `tags` (Optional) - The tags that you want to add to the Filter resource. A tag consists of a key and a value.
* `finding_criteria` (Required) - Represents the criteria to be used in the filter for querying findings. Contains one or more `criterion` blocks, documented [below](#criterion).

### criterion

The `criterion` block suports the following:

* `field` - (Required) The name of the field to be evaluated. The full list of field names can be found in [AWS documentation](https://docs.aws.amazon.com/guardduty/latest/ug/guardduty_filter-findings.html#filter_criteria).
* `equals` - (Optional) List of string values to be evaluated.
* `not_equals` - (Optional) List of string values to be evaluated.
* `greater_than` - (Optional) A value to be evaluated. Accepts either an integer or a date in [RFC 3339 format](https://tools.ietf.org/html/rfc3339#section-5.8).
* `greater_than_or_equal` - (Optional) A value to be evaluated. Accepts either an integer or a date in [RFC 3339 format](https://tools.ietf.org/html/rfc3339#section-5.8).
* `less_than` - (Optional) A value to be evaluated. Accepts either an integer or a date in [RFC 3339 format](https://tools.ietf.org/html/rfc3339#section-5.8).
* `less_than_or_equal` - (Optional) A value to be evaluated. Accepts either an integer or a date in [RFC 3339 format](https://tools.ietf.org/html/rfc3339#section-5.8).

## Attributes Reference

In addition to all arguments above, the following attribute is exported:

* `arn` - The ARN of the GuardDuty filter.
* `id` - A compound field, consisting of the ID of the GuardDuty detector and the name of the filter.

## Import

GuardDuty filters can be imported using the detector ID and filter's name separated by a colon, e.g.

```
$ terraform import aws_guardduty_filter.MyFilter 00b00fd5aecc0ab60a708659477e9617:MyFilter
```
