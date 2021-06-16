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

### With Redacted Fields

```terraform
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

### With Logging Filter

```terraform
resource "aws_wafv2_web_acl_logging_configuration" "example" {
  log_destination_configs = [aws_kinesis_firehose_delivery_stream.example.arn]
  resource_arn            = aws_wafv2_web_acl.example.arn

  logging_filter {
    default_behavior = "KEEP"

    filter {
      behavior = "DROP"

      condition {
        action_condition {
          action = "COUNT"
        }
      }

      condition {
        label_name_condition {
          label_name = "awswaf:111122223333:rulegroup:testRules:LabelNameZ"
        }
      }

      requirement = "MEETS_ALL"
    }

    filter {
      behavior = "KEEP"

      condition {
        action_condition {
          action = "ALLOW"
        }
      }

      requirement = "MEETS_ANY"
    }
  }
}
```

## Argument Reference

The following arguments are supported:

* `log_destination_configs` - (Required) The Amazon Kinesis Data Firehose Amazon Resource Name (ARNs) that you want to associate with the web ACL. Currently, only 1 ARN is supported.
* `logging_filter` - (Optional) A configuration block that specifies which web requests are kept in the logs and which are dropped. You can filter on the rule action and on the web request labels that were applied by matching rules during web ACL evaluation. See [Logging Filter](#logging-filter) below for more details.
* `redacted_fields` - (Optional) The parts of the request that you want to keep out of the logs. Up to 100 `redacted_fields` blocks are supported. See [Redacted Fields](#redacted-fields) below for more details.
* `resource_arn` - (Required) The Amazon Resource Name (ARN) of the web ACL that you want to associate with `log_destination_configs`.

### Logging Filter

The `logging_filter` block supports the following arguments:

* `default_behavior` - (Required) Default handling for logs that don't match any of the specified filtering conditions. Valid values: `KEEP` or `DROP`.
* `filter` - (Required) Filter(s) that you want to apply to the logs. See [Filter](#filter) below for more details.

### Filter

The `filter` block supports the following arguments:

* `behavior` - (Required) How to handle logs that satisfy the filter's conditions and requirement. Valid values: `KEEP` or `DROP`.
* `condition` - (Required) Match condition(s) for the filter. See [Condition](#condition) below for more details.
* `requirement` - (Required) Logic to apply to the filtering conditions. You can specify that, in order to satisfy the filter, a log must match all conditions or must match at least one condition. Valid values: `MEETS_ALL` or `MEETS_ANY`.

### Condition

The `condition` block supports the following arguments:

~> **Note:** Either `action_condition` or `label_name_condition` must be specified.  

* `action_condition` - (Optional) A single action condition. See [Action Condition](#action-condition) below for more details.
* `label_name_condition` - (Optional) A single label name condition. See [Label Name Condition](#label-name-condition) below for more details.

### Action Condition

The `action_condition` block supports the following argument:

* `action` - (Required) The action setting that a log record must contain in order to meet the condition. Valid values: `ALLOW`, `BLOCK`, `COUNT`.

### Label Name Condition

The `label_name_condition` block supports the following argument:

* `label_name` - (Required) The label name that a log record must contain in order to meet the condition. This must be a fully qualified label name. Fully qualified labels have a prefix, optional namespaces, and label name. The prefix identifies the rule group or web ACL context of the rule that added the label.

### Redacted Fields

The `redacted_fields` block supports the following arguments:

~> **NOTE:** Only one of `method`, `query_string`, `single_header` or `uri_path` can be specified.

* `all_query_arguments` - (Optional, **DEPRECATED**) Redact all query arguments.
* `body` - (Optional, **DEPRECATED**) Redact the request body, which immediately follows the request headers.
* `method` - (Optional) Redact the HTTP method. Must be specified as an empty configuration block `{}`. The method indicates the type of operation that the request is asking the origin to perform.
* `query_string` - (Optional) Redact the query string. Must be specified as an empty configuration block `{}`. This is the part of a URL that appears after a `?` character, if any.
* `single_header` - (Optional) Redact a single header. See [Single Header](#single-header) below for details.
* `single_query_argument` - (Optional, **DEPRECATED**) Redact a single query argument. See [Single Query Argument](#single-query-argument) below for details.
* `uri_path` - (Optional) Redact the request URI path. Must be specified as an empty configuration block `{}`. This is the part of a web request that identifies a resource, for example, `/images/daily-ad.jpg`.

### Single Header

Redact a single header. Provide the name of the header to redact, for example, `User-Agent` or `Referer` (provided as lowercase strings).

The `single_header` block supports the following arguments:

* `name` - (Optional) The name of the query header to redact. This setting must be provided as lower case characters.

### Single Query Argument (**DEPRECATED**)

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
```
