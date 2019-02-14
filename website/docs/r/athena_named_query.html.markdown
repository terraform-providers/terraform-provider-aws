---
layout: "aws"
page_title: "AWS: aws_athena_named_query"
sidebar_current: "docs-aws-resource-athena-named-query"
description: |-
  Provides an Athena Named Query resource.
---

# aws_athena_named_query

Provides an Athena Named Query resource.

## Example Usage

```hcl
resource "aws_s3_bucket" "hoge" {
  bucket = "tf-test"
}

resource "aws_athena_database" "hoge" {
  name   = "users"
  bucket = "${aws_s3_bucket.hoge.bucket}"
}

resource "aws_athena_named_query" "foo" {
  name     = "bar"
  database = "${aws_athena_database.hoge.name}"
  query    = "SELECT * FROM ${aws_athena_database.hoge.name} limit 10;"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The plain language name for the query. Maximum length of 128.
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
