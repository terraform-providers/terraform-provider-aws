---
subcategory: "Access Analyzer"
layout: "aws"
page_title: "AWS: aws_accessanalyzer_analyzer"
description: |-
  Manages an Access Analyzer Analyzer
---

# Resource: aws_accessanalyzer_analyzer

Manages an Access Analyzer Analyzer. More information can be found in the [Access Analyzer User Guide](https://docs.aws.amazon.com/IAM/latest/UserGuide/what-is-access-analyzer.html).

## Example Usage

### Account Analyzer

```hcl
resource "aws_accessanalyzer_analyzer" "example" {
  analyzer_name = "example"
}
```

### Organization Analyzer

```hcl
resource "aws_organizations_organization" "example" {
  aws_service_access_principals = ["access-analyzer.amazonaws.com"]
}

resource "aws_accessanalyzer_analyzer" "example" {
  depends_on = [aws_organizations_organization.example]

  analyzer_name = "example"
  type          = "ORGANIZATION"
}
```

## Argument Reference

The following arguments are required:

* `analyzer_name` - (Required) Name of the Analyzer.

The following arguments are optional:

* `tags` - (Optional) Key-value map of resource tags.
* `type` - (Optional) Type of Analyzer. Valid values are `ACCOUNT` or `ORGANIZATION`. Defaults to `ACCOUNT`.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - Analyzer name.

## Import

Access Analyzer Analyzers can be imported using the `analyzer_name`, e.g.

```
$ terraform import aws_accessanalyzer_analyzer.example example
```
