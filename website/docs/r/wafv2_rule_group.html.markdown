---
subcategory: "WAFv2"
layout: "aws"
page_title: "AWS: aws_wafv2_rule_group"
description: |-
  Creates a WAFv2 rule group resource.
---

# Resource: aws_wafv2_rule_group

Creates a WAFv2 Rule Group resource.

## Example Usage

### Simple

```hcl
resource "aws_wafv2_rule_group" "example" {
  name     = "example-rule"
  scope    = "REGIONAL"
  capacity = 2

  rule {
    name     = "rule-1"
    priority = 1

    action {
      allow {}
    }

    statement {

      geo_match_statement {
        country_codes = ["US", "NL"]
      }
    }

    visibility_config {
      cloudwatch_metrics_enabled = false
      metric_name                = "friendly-rule-metric-name"
      sampled_requests_enabled   = false
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
```

### Complex

```hcl
resource "aws_wafv2_ip_set" "test" {
  name               = "test"
  scope              = "REGIONAL"
  ip_address_version = "IPV4"
  addresses          = ["1.1.1.1/32", "2.2.2.2/32"]
}

resource "aws_wafv2_regex_pattern_set" "test" {
  name  = "test"
  scope = "REGIONAL"

  regular_expression_list {
    regex_string = "one"
  }
}

resource "aws_wafv2_rule_group" "example" {
  name        = "complex-example"
  description = "An rule group containing all statements"
  scope       = "REGIONAL"
  capacity    = 500

  rule {
    name     = "rule-1"
    priority = 1

    action {
      block {}
    }

    statement {

      not_statement {
        statement {

          and_statement {
            statement {

              geo_match_statement {
                country_codes = ["US"]
              }
            }

            statement {

              byte_match_statement {
                positional_constraint = "CONTAINS"
                search_string         = "word"

                field_to_match {
                  all_query_arguments {}
                }

                text_transformation {
                  priority = 5
                  type     = "CMD_LINE"
                }

                text_transformation {
                  priority = 2
                  type     = "LOWERCASE"
                }
              }
            }
          }
        }
      }
    }

    visibility_config {
      cloudwatch_metrics_enabled = false
      metric_name                = "rule-1"
      sampled_requests_enabled   = false
    }
  }

  rule {
    name     = "rule-2"
    priority = 2

    action {
      count {}
    }

    statement {

      or_statement {
        statement {

          sqli_match_statement {

            field_to_match {
              body {}
            }

            text_transformation {
              priority = 5
              type     = "URL_DECODE"
            }

            text_transformation {
              priority = 4
              type     = "HTML_ENTITY_DECODE"
            }

            text_transformation {
              priority = 3
              type     = "COMPRESS_WHITE_SPACE"
            }
          }
        }

        statement {

          xss_match_statement {

            field_to_match {
              method {}
            }

            text_transformation {
              priority = 2
              type     = "NONE"
            }
          }
        }
      }
    }

    visibility_config {
      cloudwatch_metrics_enabled = false
      metric_name                = "rule-2"
      sampled_requests_enabled   = false
    }
  }

  rule {
    name     = "rule-3"
    priority = 3

    action {
      block {}
    }

    statement {

      size_constraint_statement {
        comparison_operator = "GT"
        size                = 100

        field_to_match {
          single_query_argument {
            name = "username"
          }
        }

        text_transformation {
          priority = 5
          type     = "NONE"
        }
      }
    }

    visibility_config {
      cloudwatch_metrics_enabled = false
      metric_name                = "rule-3"
      sampled_requests_enabled   = false
    }
  }

  rule {
    name     = "rule-4"
    priority = 4

    action {
      block {}
    }

    statement {

      or_statement {
        statement {

          ip_set_reference_statement {
            arn = aws_wafv2_ip_set.test.arn
          }
        }

        statement {

          regex_pattern_set_reference_statement {
            arn = aws_wafv2_regex_pattern_set.test.arn

            field_to_match {
              single_header {
                name = "referer"
              }
            }

            text_transformation {
              priority = 2
              type     = "NONE"
            }
          }
        }
      }
    }

    visibility_config {
      cloudwatch_metrics_enabled = false
      metric_name                = "rule-4"
      sampled_requests_enabled   = false
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }

  tags = {
    Name = "example-and-statement"
    Code = "123456"
  }
}
```

## Argument Reference

The following arguments are supported:

* `capacity` - (Required, Forces new resource) The web ACL capacity units (WCUs) required for this rule group. See [here](https://docs.aws.amazon.com/waf/latest/APIReference/API_CreateRuleGroup.html#API_CreateRuleGroup_RequestSyntax) for general information and [here](https://docs.aws.amazon.com/waf/latest/developerguide/waf-rule-statements-list.html) for capacity specific information.
* `description` - (Optional) A friendly description of the rule group.
* `name` - (Required, Forces new resource) A friendly name of the rule group.
* `rule` - (Optional) The rule blocks used to identify the web requests that you want to `allow`, `block`, or `count`. See [Rules](#rules) below for details.
* `scope` - (Required, Forces new resource) Specifies whether this is for an AWS CloudFront distribution or for a regional application. Valid values are `CLOUDFRONT` or `REGIONAL`. To work with CloudFront, you must also specify the region `us-east-1` (N. Virginia) on the AWS provider.
* `tags` - (Optional) An array of key:value pairs to associate with the resource.
* `visibility_config` - (Required) Defines and enables Amazon CloudWatch metrics and web request sample collection. See [Visibility Configuration](#visibility-configuration) below for details.

### Rules

Each `rule` supports the following arguments:

* `action` - (Required) The action that AWS WAF should take on a web request when it matches the rule's statement. Settings at the `aws_wafv2_web_acl` level can override the rule action setting. See [Action](#action) below for details.
* `name` - (Required, Forces new resource) A friendly name of the rule.
* `priority` - (Required) If you define more than one Rule in a WebACL, AWS WAF evaluates each request against the `rules` in order based on the value of `priority`. AWS WAF processes rules with lower priority first.
* `statement` - (Required) The AWS WAF processing statement for the rule, for example `byte_match_statement` or `geo_match_statement`. See [Statement](#statement) below for details.
* `visibility_config` - (Required) Defines and enables Amazon CloudWatch metrics and web request sample collection. See [Visibility Configuration](#visibility-configuration) below for details.

### Action

The `action` block supports the following arguments:

~> **NOTE**: One of `allow`, `block`, or `count`, expressed as an empty configuration block `{}`, is required when specifying an `action`

* `allow` - (Optional) Instructs AWS WAF to allow the web request.
* `block` - (Optional) Instructs AWS WAF to block the web request.
* `count` - (Optional) Instructs AWS WAF to count the web request and allow it.

### Statement

The processing guidance for a Rule, used by AWS WAF to determine whether a web request matches the rule. See the [documentation](https://docs.aws.amazon.com/waf/latest/developerguide/waf-rule-statements-list.html) for more information.

-> **NOTE:** Although the `statement` block is recursive, currently only 3 levels are supported.

The `statement` block supports the following arguments:

* `and_statement` - (Optional) A logical rule statement used to combine other rule statements with AND logic. See [AND Statement](#and-statement) below for details.
* `byte_match_statement` - (Optional) A rule statement that defines a string match search for AWS WAF to apply to web requests. See [Byte Match Statement](#byte-match-statement) below for details.
* `geo_match_statement` - (Optional) A rule statement used to identify web requests based on country of origin. See [GEO Match Statement](#geo-match-statement) below for details.
* `ip_set_reference_statement` - (Optional) A rule statement used to detect web requests coming from particular IP addresses or address ranges. See [IP Set Reference Statement](#ip-set-reference-statement) below for details.
* `not_statement` - (Optional) A logical rule statement used to negate the results of another rule statement. See [NOT Statement](#not-statement) below for details.
* `or_statement` - (Optional) A logical rule statement used to combine other rule statements with OR logic. See [OR Statement](#or-statement) below for details.
* `regex_pattern_set_reference_statement` - (Optional) A rule statement used to search web request components for matches with regular expressions. See [Regex Pattern Set Reference Statement](#regex-pattern-set-reference-statement) below for details.
* `size_constraint_statement` - (Optional) A rule statement that compares a number of bytes against the size of a request component, using a comparison operator, such as greater than (>) or less than (<). See [Size Constraint Statement](#size-constraint-statement) below for more details.
* `sqli_match_statement` - (Optional) An SQL injection match condition identifies the part of web requests, such as the URI or the query string, that you want AWS WAF to inspect. See [SQL Injection Match Statement](#sql-injection-match-statement) below for details.
* `xss_match_statement` - (Optional) A rule statement that defines a cross-site scripting (XSS) match search for AWS WAF to apply to web requests. See [XSS Match Statement](#xss-match-statement) below for details.

### AND Statement

A logical rule statement used to combine other rule statements with `AND` logic. You provide more than one `statement` within the `and_statement`.

The `and_statement` block supports the following arguments:

* `statement` - (Required) The statements to combine with `AND` logic. You can use any statements that can be nested. See [Statement](#statement) above for details.

### Byte Match Statement

The byte match statement provides the bytes to search for, the location in requests that you want AWS WAF to search, and other settings. The bytes to search for are typically a string that corresponds with ASCII characters.

The `byte_match_statement` block supports the following arguments:

* `field_to_match` - (Required) The part of a web request that you want AWS WAF to inspect. See [Field to Match](#field-to-match) below for details.
* `positional_constraint` - (Required) The area within the portion of a web request that you want AWS WAF to search for `search_string`. Valid values include the following: `EXACTLY`, `STARTS_WITH`, `ENDS_WITH`, `CONTAINS`, `CONTAINS_WORD`. See the AWS [documentation](https://docs.aws.amazon.com/waf/latest/APIReference/API_ByteMatchStatement.html) for more information.
* `search_string` - (Required) A string value that you want AWS WAF to search for. AWS WAF searches only in the part of web requests that you designate for inspection in `field_to_match`. The maximum length of the value is 50 bytes.
* `text_transformation` - (Required) Text transformations eliminate some of the unusual formatting that attackers use in web requests in an effort to bypass detection. See [Text Transformation](#text-transformation) below for details.

### GEO Match Statement

The `geo_match_statement` block supports the following arguments:

* `country_codes` - (Required) An array of two-character country codes, for example, [ "US", "CN" ], from the alpha-2 country ISO codes of the `ISO 3166` international standard. See the [documentation](https://docs.aws.amazon.com/waf/latest/APIReference/API_GeoMatchStatement.html) for valid values.
* `forwarded_ip_config` - (Optional) The configuration for inspecting IP addresses in an HTTP header that you specify, instead of using the IP address that's reported by the web request origin. See [Forwarded IP Config](#forwarded-ip-config) below for details.

### IP Set Reference Statement

A rule statement used to detect web requests coming from particular IP addresses or address ranges. To use this, create an `aws_wafv2_ip_set` that specifies the addresses you want to detect, then use the `ARN` of that set in this statement.

The `ip_set_reference_statement` block supports the following arguments:

* `arn` - (Required) The Amazon Resource Name (ARN) of the IP Set that this statement references.
* `ip_set_forwarded_ip_config` - (Optional) The configuration for inspecting IP addresses in an HTTP header that you specify, instead of using the IP address that's reported by the web request origin. See [IPSet Forwarded IP Config](#ipset-forwarded-ip-config) below for more details.

### NOT Statement

A logical rule statement used to negate the results of another rule statement. You provide one `statement` within the `not_statement`.

The `not_statement` block supports the following arguments:

* `statement` - (Required) The statement to negate. You can use any statement that can be nested. See [Statement](#statement) above for details.

### OR Statement

A logical rule statement used to combine other rule statements with `OR` logic. You provide more than one `statement` within the `or_statement`.

The `or_statement` block supports the following arguments:

* `statement` - (Required) The statements to combine with `OR` logic. You can use any statements that can be nested. See [Statement](#statement) above for details.

### Regex Pattern Set Reference Statement

A rule statement used to search web request components for matches with regular expressions. To use this, create a `aws_wafv2_regex_pattern_set` that specifies the expressions that you want to detect, then use the `ARN` of that set in this statement. A web request matches the pattern set rule statement if the request component matches any of the patterns in the set.

The `regex_pattern_set_reference_statement` block supports the following arguments:

* `arn` - (Required) The Amazon Resource Name (ARN) of the Regex Pattern Set that this statement references.
* `field_to_match` - (Required) The part of a web request that you want AWS WAF to inspect. See [Field to Match](#field-to-match) below for details.
* `text_transformation` - (Required) Text transformations eliminate some of the unusual formatting that attackers use in web requests in an effort to bypass detection. See [Text Transformation](#text-transformation) below for details.

### Size Constraint Statement

A rule statement that uses a comparison operator to compare a number of bytes against the size of a request component. AWS WAFv2 inspects up to the first 8192 bytes (8 KB) of a request body, and when inspecting the request URI Path, the slash `/` in
the URI counts as one character.

The `size_constraint_statement` block supports the following arguments:

* `comparison_operator` - (Required) The operator to use to compare the request part to the size setting. Valid values include: `EQ`, `NE`, `LE`, `LT`, `GE`, or `GT`.
* `field_to_match` - (Optional) The part of a web request that you want AWS WAF to inspect. See [Field to Match](#field-to-match) below for details.
* `size` - (Required) The size, in bytes, to compare to the request part, after any transformations. Valid values are integers between 0 and 21474836480, inclusive.
* `text_transformation` - (Required) Text transformations eliminate some of the unusual formatting that attackers use in web requests in an effort to bypass detection. See [Text Transformation](#text-transformation) below for details.

### SQL Injection Match Statement

An SQL injection match condition identifies the part of web requests, such as the URI or the query string, that you want AWS WAF to inspect. Later in the process, when you create a web ACL, you specify whether to allow or block requests that appear to contain malicious SQL code.

The `sqli_match_statement` block supports the following arguments:

* `field_to_match` - (Required) The part of a web request that you want AWS WAF to inspect. See [Field to Match](#field-to-match) below for details.
* `text_transformation` - (Required) Text transformations eliminate some of the unusual formatting that attackers use in web requests in an effort to bypass detection. See [Text Transformation](#text-transformation) below for details.

### XSS Match Statement

The XSS match statement provides the location in requests that you want AWS WAF to search and text transformations to use on the search area before AWS WAF searches for character sequences that are likely to be malicious strings.

The `xss_match_statement` block supports the following arguments:

* `field_to_match` - (Required) The part of a web request that you want AWS WAF to inspect. See [Field to Match](#field-to-match) below for details.
* `text_transformation` - (Required) Text transformations eliminate some of the unusual formatting that attackers use in web requests in an effort to bypass detection. See [Text Transformation](#text-transformation) below for details.

### Field to Match

The part of a web request that you want AWS WAF to inspect. Include the single `field_to_match` type that you want to inspect, with additional specifications as needed, according to the type. You specify a single request component in `field_to_match` for each rule statement that requires it. To inspect more than one component of a web request, create a separate rule statement for each component. See the [documentation](https://docs.aws.amazon.com/waf/latest/developerguide/waf-rule-statement-fields.html#waf-rule-statement-request-component) for more details.

The `field_to_match` block supports the following arguments:

~> **NOTE**: An empty configuration block `{}` should be used when specifying `all_query_arguments`, `body`, `method`, or `query_string` attributes

* `all_query_arguments` - (Optional) Inspect all query arguments.
* `body` - (Optional) Inspect the request body, which immediately follows the request headers.
* `method` - (Optional) Inspect the HTTP method. The method indicates the type of operation that the request is asking the origin to perform.
* `query_string` - (Optional) Inspect the query string. This is the part of a URL that appears after a `?` character, if any.
* `single_header` - (Optional) Inspect a single header. See [Single Header](#single-header) below for details.
* `single_query_argument` - (Optional) Inspect a single query argument. See [Single Query Argument](#single-query-argument) below for details.
* `uri_path` - (Optional) Inspect the request URI path. This is the part of a web request that identifies a resource, for example, `/images/daily-ad.jpg`.

### Forwarded IP Config

The configuration for inspecting IP addresses in an HTTP header that you specify, instead of using the IP address that's reported by the web request origin. Commonly, this is the X-Forwarded-For (XFF) header, but you can specify
any header name. If the specified header isn't present in the request, AWS WAFv2 doesn't apply the rule to the web request at all.
AWS WAFv2 only evaluates the first IP address found in the specified HTTP header.

The `forwarded_ip_config` block supports the following arguments:

* `fallback_behavior` - (Required) - The match status to assign to the web request if the request doesn't have a valid IP address in the specified position. Valid values include: `MATCH` or `NO_MATCH`.
* `header_name` - (Required) - The name of the HTTP header to use for the IP address.

### IPSet Forwarded IP Config

The configuration for inspecting IP addresses in an HTTP header that you specify, instead of using the IP address that's reported by the web request origin. Commonly, this is the X-Forwarded-For (XFF) header, but you can specify any header name.

The `ip_set_forwarded_ip_config` block supports the following arguments:

* `fallback_behavior` - (Required) - The match status to assign to the web request if the request doesn't have a valid IP address in the specified position. Valid values include: `MATCH` or `NO_MATCH`.
* `header_name` - (Required) - The name of the HTTP header to use for the IP address.
* `position` - (Required) - The position in the header to search for the IP address. Valid values include: `FIRST`, `LAST`, or `ANY`. If `ANY` is specified and the header contains more than 10 IP addresses, AWS WAFv2 inspects the last 10.

### Single Header

Inspect a single header. Provide the name of the header to inspect, for example, `User-Agent` or `Referer` (provided as lowercase strings).

The `single_header` block supports the following arguments:

* `name` - (Optional) The name of the query header to inspect. This setting must be provided as lower case characters.

### Single Query Argument

Inspect a single query argument. Provide the name of the query argument to inspect, such as `UserName` or `SalesRegion` (provided as lowercase strings).

The `single_query_argument` block supports the following arguments:

* `name` - (Optional) The name of the query header to inspect. This setting must be provided as lower case characters.

### Text Transformation

The `text_transformation` block supports the following arguments:

* `priority` - (Required) The relative processing order for multiple transformations that are defined for a rule statement. AWS WAF processes all transformations, from lowest priority to highest, before inspecting the transformed content.
* `type` - (Required) The transformation to apply, you can specify the following types: `NONE`, `COMPRESS_WHITE_SPACE`, `HTML_ENTITY_DECODE`, `LOWERCASE`, `CMD_LINE`, `URL_DECODE`. See the [documentation](https://docs.aws.amazon.com/waf/latest/APIReference/API_TextTransformation.html) for more details.

### Visibility Configuration

The `visibility_config` block supports the following arguments:

* `cloudwatch_metrics_enabled` - (Required) A boolean indicating whether the associated resource sends metrics to CloudWatch. For the list of available metrics, see [AWS WAF Metrics](https://docs.aws.amazon.com/waf/latest/developerguide/monitoring-cloudwatch.html#waf-metrics).
* `metric_name` - (Required, Forces new resource) A friendly name of the CloudWatch metric. The name can contain only alphanumeric characters (A-Z, a-z, 0-9) hyphen(-) and underscore (_), with length from one to 128 characters. It can't contain whitespace or metric names reserved for AWS WAF, for example `All` and `Default_Action`.
* `sampled_requests_enabled` - (Required) A boolean indicating whether AWS WAF should store a sampling of the web requests that match the rules. You can view the sampled requests through the AWS WAF console.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the WAF rule group.
* `arn` - The ARN of the WAF rule group.

## Import

WAFv2 Rule Group can be imported using `ID/name/scope` e.g.

```
$ terraform import aws_wafv2_rule_group.example a1b2c3d4-d5f6-7777-8888-9999aaaabbbbcccc/example/REGIONAL
```
