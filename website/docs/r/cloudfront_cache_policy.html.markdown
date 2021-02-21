---
subcategory: "CloudFront"
layout: "aws"
page_title: "AWS: aws_cloudfront_cache_policy"
description: |-
  Provides a cache policy for a CloudFront ditribution. When it’s attached to a cache behavior, 
  the cache policy determines the the values that CloudFront includes in the cache key. These 
  values can include HTTP headers, cookies, and URL query strings. CloudFront uses the cache 
  key to find an object in its cache that it can return to the viewer. It also determines the 
  default, minimum, and maximum time to live (TTL) values that you want objects to stay in the 
  CloudFront cache. 
---

# Resource: aws_cloudfront_cache_policy

## Example Usage

The following example below creates a CloudFront cache policy.

```hcl
resource "aws_cloudfront_cache_policy" "example" {
  name        = "example-policy"
  comment     = "test comment"
  default_ttl = 50
  max_ttl     = 100
  min_ttl     = 1
  parameters_in_cache_key_and_forwarded_to_origin {
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
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) A unique name to identify the cache policy.
* `min_ttl` - (Required) The minimum amount of time, in seconds, that you want objects to stay in the CloudFront cache before CloudFront sends another request to the origin to see if the object has been updated.
* `max_ttl` - (Optional) The maximum amount of time, in seconds, that objects stay in the CloudFront cache before CloudFront sends another request to the origin to see if the object has been updated.
* `default_ttl` - (Optional) The default amount of time, in seconds, that you want objects to stay in the CloudFront cache before CloudFront sends another request to the origin to see if the object has been updated.
* `comment` - (Optional) A comment to describe the cache policy.
* `parameters_in_cache_key_and_forwarded_to_origin` - (Optional) The HTTP headers, cookies, and URL query strings to include in the cache key. See [Parameters In Cache Key And Forwarded To Origin](#parameters-in-cache-key-and-forwarded-to-origin) for more information.

### Parameters In Cache Key And Forwarded To Origin

* `cookies_config` - (Required) Object that determines whether any cookies in viewer requests (and if so, which cookies) are included in the cache key and automatically included in requests that CloudFront sends to the origin. See [Cookies Config](#cookies-config) for more information.
* `headers_config` - (Required) Object that determines whether any HTTP headers (and if so, which headers) are included in the cache key and automatically included in requests that CloudFront sends to the origin. See [Headers Config](#headers-config) for more information.
* `query_strings_config` - (Required) Object that determines whether any URL query strings in viewer requests (and if so, which query strings) are included in the cache key and automatically included in requests that CloudFront sends to the origin. See [Query Strings Config](#query-strings-config) for more information.
* `enable_accept_encoding_brotli` - (Optional) A flag that can affect whether the Accept-Encoding HTTP header is included in the cache key and included in requests that CloudFront sends to the origin.
* `enable_accept_encoding_gzip` - (Optional) A flag that can affect whether the Accept-Encoding HTTP header is included in the cache key and included in requests that CloudFront sends to the origin.

### Cookies Config

`cookie_behavior` - (Required) Determines whether any cookies in viewer requests are included in the cache key and automatically included in requests that CloudFront sends to the origin. Valid values are `none`, `whitelist`, `allExcept`, `all`.
`cookies` - (Optional) Object that contains a list of cookie names. See [Items](#items) for more information.

### Headers Config

`header_behavior` - (Required) Determines whether any HTTP headers are included in the cache key and automatically included in requests that CloudFront sends to the origin. Valid values are `none`, `whitelist`.
`headers` - (Optional) Object that contains a list of header names. See [Items](#items) for more information.

### Query String Config

`query_string_behavior` - (Required) Determines whether any URL query strings in viewer requests are included in the cache key and automatically included in requests that CloudFront sends to the origin. Valid values are `none`, `whitelist`, `allExcept`, `all`.
`query_strings` - (Optional) Object that contains a list of query string names. See [Items](#items) for more information.

### Items

`items` - (Required) A list of item names (cookies, headers, or query strings).

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `etag` - The current version of the cache policy.
* `id` - The identifier for the cache policy.
