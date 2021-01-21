---
subcategory: "Athena"
layout: "aws"
page_title: "AWS: aws_athena_named_query"
description: |-
  Provides an Athena Named Query resource.
---

# Resource: aws_athena_named_query

Provides an Athena Named Query resource.

## Example Usage

```hcl
resource "aws_s3_bucket" "hoge" {
  bucket = "tf-test"
}

resource "aws_kms_key" "test" {
  deletion_window_in_days = 7
  description             = "Athena KMS Key"
}

resource "aws_athena_workgroup" "test" {
  name = "example"

  configuration {
    result_configuration {
      encryption_configuration {
        encryption_option = "SSE_KMS"
        kms_key_arn       = aws_kms_key.test.arn
      }
    }
  }
}

resource "aws_athena_database" "hoge" {
  name   = "users"
  bucket = aws_s3_bucket.hoge.id
}

resource "aws_athena_named_query" "foo" {
  name      = "bar"
  workgroup = aws_athena_workgroup.test.id
  database  = aws_athena_database.hoge.name
  query     = "SELECT * FROM ${aws_athena_database.hoge.name} limit 10;"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The plain language name for the query. Maximum length of 128.
* `workgroup` - (Optional) The workgroup to which the query belongs. Defaults to `primary`
* `database` - (Required) The database to which the query belongs.
* `query` - (Required) The text of the query itself. In other words, all query statements. Maximum length of 262144.
* `description` - (Optional) A brief explanation of the query. Maximum length of 1024.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The unique ID of the query.

## Import

Athena Named Query can be imported using the query ID, e.g.

```
$ terraform import aws_athena_named_query.example 0123456789
```
