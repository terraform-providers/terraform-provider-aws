---
subcategory: "CloudWatch"
layout: "aws"
page_title: "AWS: aws_cloudwatch_metrics"
description: |-
  Get metrics from CloudWatch.
---

# Data Source: aws_cloudwatch_metrics

Use this data source to get metrics from CloudWatch

## Example Usage

```hcl
data "aws_cloudwatch_metrics" "example" {
  namespace   = "AWS/EC2"
  metric_name = "NetworkPacketsIn"
}
```

## Argument Reference

The following arguments are supported:

* `namespace` - (Optional) The namespace of the CloudWatch Metric
* `metric_name` - (Optional) The metric name of the CloudWatch Metric
* `dimension` - (Optional) The dimension of the CloudWatch metric
    * `dimension.#.name` - The dimension name.
    * `dimension.#.value` - The dimension value.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `metrics` - A list of all Metrics matching the namespace, metric_name or dimension
    * `metrics.#.namespace` - The namespace of the CloudWatch Metric.
    * `metrics.#.metric_name` - The metric name of the CloudWatch Metric.
    * `metrics.dimension` - The dimension of the CloudWatch metric
        * `metrics.dimension.#.name` - The dimension name.
        * `metrics.dimension.#.value` - The dimension value.
