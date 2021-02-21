---
subcategory: "CloudFront"
layout: "aws"
page_title: "AWS: aws_cloudfront_cache_policy"
description: |-
  Use this data source to retrieve information about a CloudFront cache policy.
---

# Resource: aws_cloudfront_cache_policy

## Example Usage

The following example below creates a CloudFront cache policy.

```hcl
data "aws_cloudfront_cache_policy" "example" {
  name = "example-policy"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Optional) A unique name to identify the cache policy.
* `id` - (Optional) The identifier for the cache policy.

## Attributes Reference

* `etag` - The current version of the cache policy.
* `min_ttl` - The minimum amount of time, in seconds, that you want objects to stay in the CloudFront cache before CloudFront sends another request to the origin to see if the object has been updated.
* `max_ttl` - The maximum amount of time, in seconds, that objects stay in the CloudFront cache before CloudFront sends another request to the origin to see if the object has been updated.
* `default_ttl` - The default amount of time, in seconds, that you want objects to stay in the CloudFront cache before CloudFront sends another request to the origin to see if the object has been updated.
* `comment` - A comment to describe the cache policy.
* `parameters_in_cache_key_and_forwarded_to_origin` - The HTTP headers, cookies, and URL query strings to include in the cache key. See [Parameters In Cache Key And Forwarded To Origin](#parameters-in-cache-key-and-forwarded-to-origin) for more information.

### Parameters In Cache Key And Forwarded To Origin

* `cookies_config` - Object that determines whether any cookies in viewer requests (and if so, which cookies) are included in the cache key and automatically included in requests that CloudFront sends to the origin. See [Cookies Config](#cookies-config) for more information.
* `headers_config` - Object that determines whether any HTTP headers (and if so, which headers) are included in the cache key and automatically included in requests that CloudFront sends to the origin. See [Headers Config](#headers-config) for more information.
* `query_strings_config` - Object that determines whether any URL query strings in viewer requests (and if so, which query strings) are included in the cache key and automatically included in requests that CloudFront sends to the origin. See [Query Strings Config](#query-strings-config) for more information.
* `enable_accept_encoding_brotli` - A flag that can affect whether the Accept-Encoding HTTP header is included in the cache key and included in requests that CloudFront sends to the origin.
* `enable_accept_encoding_gzip` - A flag that can affect whether the Accept-Encoding HTTP header is included in the cache key and included in requests that CloudFront sends to the origin.

### Cookies Config

`cookie_behavior` - Determines whether any cookies in viewer requests are included in the cache key and automatically included in requests that CloudFront sends to the origin. Valid values are `none`, `whitelist`, `allExcept`, `all`.
`cookies` - Object that contains a list of cookie names. See [Items](#items) for more information.

### Headers Config

`header_behavior` - Determines whether any HTTP headers are included in the cache key and automatically included in requests that CloudFront sends to the origin. Valid values are `none`, `whitelist`.
`headers` - Object that contains a list of header names. See [Items](#items) for more information.

### Query String Config

`query_string_behavior` - Determines whether any URL query strings in viewer requests are included in the cache key and automatically included in requests that CloudFront sends to the origin. Valid values are `none`, `whitelist`, `allExcept`, `all`.
`query_strings` - Object that contains a list of query string names. See [Items](#items) for more information.

### Items

`items` - A list of item names (cookies, headers, or query strings).



