---
subcategory: "Lake Formation"
layout: "aws"
page_title: "AWS: aws_lakeformation_resource"
description: |-
    Provides details about a Lake Formation resource.
---

# Data Source: aws_lakeformation_resource

Provides details about a Lake Formation resource.

## Example Usage

```hcl
data "aws_lakeformation_resource" "example" {
  arn = "arn:aws:s3:::tf-acc-test-9151654063908211878"
}
```

## Argument Reference

* `arn` – (Required) Amazon Resource Name (ARN) of the resource, an S3 path.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `last_modified` - The date and time the resource was last modified in [RFC 3339 format](https://tools.ietf.org/html/rfc3339#section-5.8).
* `role_arn` – Role that the resource was registered with.
