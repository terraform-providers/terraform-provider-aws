---
subcategory: "WAFv2"
layout: "aws"
page_title: "AWS: aws_wafv2_web_acl_logging_configuration"
description: |-
  Creates a WAFv2 Web ACL Logging Configuration resource.
---

# Resource: aws_wafv2_web_acl_logging_configuration

Creates a WAFv2 Web ACL Logging Configuration resource.

-> **Note:** To start logging from a WAFv2 Web ACL, an Amazon Kinesis Data Firehose (e.g. [`aws_kinesis_firehose_delivery_stream` resource](/docs/providers/aws/r/kinesis_firehose_delivery_stream.html) must also be created with a PUT source (not a stream) and in the region that you are operating.
If you are capturing logs for Amazon CloudFront, always create the firehose in US East (N. Virginia).
Be sure to give the data firehose a name that starts with the prefix `aws-waf-logs-`.

## Example Usage

```hcl
resource "aws_wafv2_web_acl_logging_configuration" "example" {
  log_destination_configs = [aws_kinesis_firehose_delivery_stream.example.arn]
  resource_arn            = aws_wafv2_web_acl.example.arn
  redacted_fields {
    single_header {
      name = "user-agent"
    }
  }
}
```

## Argument Reference

The following arguments are supported:

* `log_destination_configs` - (Required) The Amazon Kinesis Data Firehose Amazon Resource Name (ARNs) that you want to associate with the web ACL. Currently, only 1 ARN is supported.
* `resource_arn` - (Required) The Amazon Resource Name (ARN) of the web ACL that you want to associate with `log_destination_configs`.
* `redacted_fields` - (Optional) The parts of the request that you want to keep out of the logs. Up to 100 `redacted_fields` blocks are supported.

The `redacted_fields` block supports the following arguments:

* `all_query_arguments` - (Optional) Redact all query arguments.
* `body` - (Optional) Redact the request body, which immediately follows the request headers.
* `method` - (Optional) Redact the HTTP method. The method indicates the type of operation that the request is asking the origin to perform.
* `query_string` - (Optional) Redact the query string. This is the part of a URL that appears after a `?` character, if any.
* `single_header` - (Optional) Redact a single header. See [Single Header](#single-header) below for details.
* `single_query_argument` - (Optional) Redact a single query argument. See [Single Query Argument](#single-query-argument) below for details.
* `uri_path` - (Optional) Redact the request URI path. This is the part of a web request that identifies a resource, for example, `/images/daily-ad.jpg`.

### Single Header

Redact a single header. Provide the name of the header to redact, for example, `User-Agent` or `Referer` (provided as lowercase strings).

The `single_header` block supports the following arguments:

* `name` - (Optional) The name of the query header to redact. This setting must be provided as lower case characters.

### Single Query Argument

Redact a single query argument. Provide the name of the query argument to redact, such as `UserName` or `SalesRegion` (provided as lowercase strings).

The `single_query_argument` block supports the following arguments:

* `name` - (Optional) The name of the query header to redact. This setting must be provided as lower case characters.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The Amazon Resource Name (ARN) of the WAFv2 Web ACL.

## Import

WAFv2 Web ACL Logging Configurations can be imported using the WAFv2 Web ACL ARN e.g.

```
$ terraform import aws_wafv2_web_acl_logging_configuration.example arn:aws:wafv2:us-west-2:123456789012:regional/webacl/test-logs/a1b2c3d4-5678-90ab-cdef
