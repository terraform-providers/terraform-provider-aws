---
layout: "aws"
page_title: "AWS: aws_inspector_rules_packages"
sidebar_current: "docs-aws-datasource-inspector-rules-packages"
description: |-
    Provides a list of AWS Inspector Rules packages which can be used by AWS Inspector.
---

# Data Source: aws_inspector_rules_packages

The AWS Inspector Rules Packages data source allows access to the list of AWS
Inspector Rules Packages which can be used by AWS Inspector within the region
configured in the provider.

## Example Usage

```hcl
# Declare the data source
data "aws_inspector_rules_packages" "rules" {}

# e.g. Use in aws_inspector_assessment_template
resource "aws_inspector_resource_group" "group" {
  tags = {
    test = "test"
  }
}

resource "aws_inspector_assessment_target" "assessment" {
  name               = "test"
  resource_group_arn = "${aws_inspector_resource_group.group.arn}"
}

resource "aws_inspector_assessment_template" "assessment" {
  name       = "Test"
  target_arn = "${aws_inspector_assessment_target.assessment.arn}"
  duration   = "60"

  rules_package_arns = ["${data.aws_inspector_rules_packages.rules.arns}"]
}
```

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arns` - A list of the AWS Inspector Rules Packages arns available in the AWS region.
