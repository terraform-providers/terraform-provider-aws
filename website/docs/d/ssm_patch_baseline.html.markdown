---
subcategory: "SSM"
layout: "aws"
page_title: "AWS: aws_ssm_patch_baseline"
description: |-
  Provides an SSM Patch Baseline data source
---

# Data Source: aws_ssm_patch_baseline

Provides an SSM Patch Baseline data source. Useful if you wish to reuse the default baselines provided.

## Example Usage

To retrieve a baseline provided by AWS:

```hcl
data "aws_ssm_patch_baseline" "centos" {
  owner            = "AWS"
  name_prefix      = "AWS-"
  operating_system = "CENTOS"
}
```

To retrieve a baseline on your account:

```hcl
data "aws_ssm_patch_baseline" "default_custom" {
  owner            = "Self"
  name_prefix      = "MyCustomBaseline"
  default_baseline = true
  operating_system = "WINDOWS"
}
```

## Argument Reference

The following arguments are supported:

* `owner` - (Required) The owner of the baseline. Valid values: `All`, `AWS`, `Self` (the current account).

* `name_prefix` - (Optional) Filter results by the baseline name prefix.

* `default_baseline` - (Optional) Filters the results against the baselines default_baseline field.

* `operating_system` - (Optional) The specified OS for the baseline.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The id of the baseline.
* `name` - The name of the baseline.
* `description` - The description of the baseline.
