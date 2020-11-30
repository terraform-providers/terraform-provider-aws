---
subcategory: "XRay"
layout: "aws"
page_title: "AWS: aws_xray_sampling_rule"
description: |-
    Creates and manages an AWS XRay Sampling Rule.
---

# Resource: aws_xray_sampling_rule

Creates and manages an AWS XRay Sampling Rule.

## Example Usage

```hcl
resource "aws_xray_sampling_rule" "example" {
  rule_name      = "example"
  priority       = 10000
  version        = 1
  reservoir_size = 1
  fixed_rate     = 0.05
  url_path       = "*"
  host           = "*"
  http_method    = "*"
  service_type   = "*"
  service_name   = "*"
  resource_arn   = "*"

  attributes = {
    Hello = "Tris"
  }
}
```

## Argument Reference

* `rule_name` - (Required) The name of the sampling rule.
* `resource_arn` - (Required) Matches the ARN of the AWS resource on which the service runs.
* `priority` - (Required) The priority of the sampling rule.
* `fixed_rate` - (Required) The percentage of matching requests to instrument, after the reservoir is exhausted.
* `reservoir_size` - (Required) A fixed number of matching requests to instrument per second, prior to applying the fixed rate. The reservoir is not used directly by services, but applies to all services using the rule collectively.
* `service_name` - (Required) Matches the `name` that the service uses to identify itself in segments.
* `service_type` - (Required) Matches the `origin` that the service uses to identify its type in segments.
* `host` - (Required) Matches the hostname from a request URL.
* `http_method` - (Required) Matches the HTTP method of a request.
* `url_path` - (Required) Matches the path from a request URL.
* `version` - (Required) The version of the sampling rule format (`1` )
* `attributes` - (Optional) Matches attributes derived from the request.
* `tags` - (Optional) Key-value mapping of resource tags

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The name of the sampling rule.
* `arn` - The ARN of the sampling rule.

## Import

XRay Sampling Rules can be imported using the name, e.g.

```
$ terraform import aws_xray_sampling_rule.example example
```
