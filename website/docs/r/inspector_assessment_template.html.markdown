---
subcategory: "Inspector"
layout: "aws"
page_title: "AWS: aws_inspector_assessment_template"
description: |-
  Provides a Inspector assessment template.
---

# Resource: aws_inspector_assessment_template

Provides a Inspector assessment template

## Example Usage

```hcl
resource "aws_inspector_assessment_template" "example" {
  name       = "example"
  target_arn = aws_inspector_assessment_target.example.arn
  duration   = 3600

  rules_package_arns = [
    "arn:aws:inspector:us-west-2:758058086616:rulespackage/0-9hgA516p",
    "arn:aws:inspector:us-west-2:758058086616:rulespackage/0-H5hpSawc",
    "arn:aws:inspector:us-west-2:758058086616:rulespackage/0-JJOtZiqQ",
    "arn:aws:inspector:us-west-2:758058086616:rulespackage/0-vg5GGHSD",
  ]
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the assessment template.
* `target_arn` - (Required) The assessment target ARN to attach the template to.
* `duration` - (Required) The duration of the inspector run.
* `rules_package_arns` - (Required) The rules to be used during the run.
* `tags` - (Optional) Key-value map of tags for the Inspector assessment template.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The template assessment ARN.

## Import

`aws_inspector_assessment_template` can be imported by using the template assessment ARN, e.g.

```
$ terraform import aws_inspector_assessment_template.example arn:aws:inspector:us-west-2:123456789012:target/0-9IaAzhGR/template/0-WEcjR8CH
```
