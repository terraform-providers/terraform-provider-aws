---
layout: "aws"
page_title: "AWS: aws_media_store_container_cors_policy"
sidebar_current: "docs-aws-resource-media-store-container-cors-policy"
description: |-
  Provides a MediaStore Container Cors Policy.
---

# aws_media_store_container_cors_policy

Provides a MediaStore Container Cors Policy.

## Example Usage

```hcl

resource "aws_media_store_container" "example" {
  name = "example"
}

resource "aws_media_store_container_cors_policy" "example" {
  container_name = "${aws_media_store_container.example.name}"

  cors_policy {
    allowed_headers = ["*"]
    allowed_methods = ["GET"]
    allowed_origins = ["*"]
  }

EOF
}
```

## Argument Reference

The following arguments are supported:

* `container_name` - (Required, ForceNew) The name of the container.
* `cors_policy` - (Required) A rule for a CORS policy. Add up to 100 rules to a CORS policy. If more than one rule applies, the service uses the first applicable rule listed. See [below](#cors_policy) for detail.

### Nested Fields

#### `cors_policy`

* `allowed_headers` - (Required) Specifies which headers are allowed in a preflight `OPTIONS` request through the `Access-Control-Request-Headers` header.
* `allowed_methods` - (Required) Identifies an HTTP method that the origin that is specified in the rule is allowed to execute.  The valid values are: `PUT`, `GET`, `DELETE`, and `HEAD`.
* `allowed_origins` - (Required) One or more response headers that you want users to be able to access from their applications. e.g. from a JavaScript `XMLHttpRequest` object.
* `expose_headers` - (Optional) One or more headers in the response that you want users to be able to access from their applications. e.g. from a JavaScript `XMLHttpRequest` object.
* `max_age_seconds` - (Optional) The time in seconds that your browser caches the preflight response for the specified resource.(Default: `0`).

## Import

MediaStore Container Cors Policy can be imported using the MediaStore Container Name, e.g.

```
$ terraform import aws_media_store_container_cors_policy.example example
```
