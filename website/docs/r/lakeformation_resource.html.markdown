---
subcategory: "Lake Formation"
layout: "aws"
page_title: "AWS: aws_lakeformation_resource"
description: |-
  Registers a Lake Formation resource as managed by the Data Catalog.
---

# Resource: aws_lakeformation_resource

Registers a Lake Formation resource (e.g. S3 bucket) as managed by the Data Catalog. In other words, the S3 path is added to the data lake.

Choose a role that has read/write access to the chosen Amazon S3 path or use the service-linked role. When you register the S3 path, the service-linked role and a new inline policy are created on your behalf. Lake Formation adds the first path to the inline policy and attaches it to the service-linked role. When you register subsequent paths, Lake Formation adds the path to the existing policy.

## Example Usage

```hcl
data "aws_s3_bucket" "example" {
  bucket = "an-example-bucket"
}

resource "aws_lakeformation_resource" "example" {
  arn = data.aws_s3_bucket.example.arn
}
```

## Argument Reference

* `arn` – (Required) Amazon Resource Name (ARN) of the resource, an S3 path.
* `role_arn` – (Optional) Role that has read/write access to the resource. If not provided, the Lake Formation service-linked role must exist and is used.

~> **NOTE:** AWS does not support registering an S3 location with an IAM role and subsequently updating the S3 location registration to a service-linked role.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `last_modified` - (Optional) The date and time the resource was last modified in [RFC 3339 format](https://tools.ietf.org/html/rfc3339#section-5.8).
