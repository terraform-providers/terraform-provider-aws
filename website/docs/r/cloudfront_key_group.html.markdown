---
subcategory: "CloudFront"
layout: "aws"
page_title: "AWS: aws_cloudfront_key_group"
description: |-
  Provides a CloudFront key group.
---

# Resource: aws_cloudfront_key_group

## Example Usage

The following example below creates a CloudFront key group.

```terraform
resource "aws_cloudfront_public_key" "example" {
  comment     = "example public key"
  encoded_key = file("public_key.pem")
  name        = "example-key"
}

resource "aws_cloudfront_key_group" "example" {
  comment = "example key group"
  items   = [aws_cloudfront_public_key.example.id]
  name    = "example-key-group"
}
```

## Argument Reference

The following arguments are supported:

* `comment` - (Optional) A comment to describe the key group..
* `items` - (Required) A list of the identifiers of the public keys in the key group.
* `name` - (Required) A name to identify the key group.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `etag` - The identifier for this version of the key group.
* `id` - The identifier for the key group.
