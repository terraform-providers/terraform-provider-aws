---
subcategory: "CloudFront"
layout: "aws"
page_title: "AWS: aws_cloudfront_origin_request_policy"
description: |-
  Determines the values that CloudFront includes in requests that it sends to the origin.
---

# Resource: aws_cloudfront_origin_request_policy

## Example Usage

The following example below creates a CloudFront origin request policy.

```hcl
resource "aws_cloudfront_origin_request_policy" "example" {
  name    = "example-policy"
  comment = "example comment"
  cookies_config {
    cookie_behavior = "whitelist"
    cookies {
      items = ["example"]
    }
  }
  headers_config {
    header_behavior = "whitelist"
    headers {
      items = ["example"]
    }
  }
  query_strings_config {
    query_string_behavior = "whitelist"
    query_strings {
      items = ["example"]
    }
  }
}

```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Unique name to identify the origin request policy.
* `comment` - (Optional) Comment to describe the origin request policy.
* `cookies_config` - (Required) Object that determines whether any cookies in viewer requests (and if so, which cookies) are included in the origin request key and automatically included in requests that CloudFront sends to the origin. See [Cookies Config](#cookies-config) for more information.
* `headers_config` - (Required) Object that determines whether any HTTP headers (and if so, which headers) are included in the origin request key and automatically included in requests that CloudFront sends to the origin. See [Headers Config](#headers-config) for more information.
* `query_strings_config` - (Required) Object that determines whether any URL query strings in viewer requests (and if so, which query strings) are included in the origin request key and automatically included in requests that CloudFront sends to the origin. See [Query Strings Config](#query-strings-config) for more information.

### Cookies Config

`cookie_behavior` - (Required) Determines whether any cookies in viewer requests are included in the origin request key and automatically included in requests that CloudFront sends to the origin. Valid values are `none`, `whitelist` `all`.
`cookies` - (Optional) Object that contains a list of cookie names. See [Items](#items) for more information.

### Headers Config

`header_behavior` - (Required) Determines whether any HTTP headers are included in the origin request key and automatically included in requests that CloudFront sends to the origin. Valid values are `none`, `whitelist`, `allViewer`, `allViewerAndWhitelistCloudFront`.
`headers` - (Optional) Object that contains a list of header names. See [Items](#items) for more information.

### Query String Config

`query_string_behavior` - (Required) Determines whether any URL query strings in viewer requests are included in the origin request key and automatically included in requests that CloudFront sends to the origin. Valid values are `none`, `whitelist`, `all`.
`query_strings` - (Optional) Object that contains a list of query string names. See [Items](#items) for more information.

### Items

`items` - (Required) List of item names (cookies, headers, or query strings).

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `etag` - The current version of the origin request policy.
* `id` - The identifier for the origin request policy.
