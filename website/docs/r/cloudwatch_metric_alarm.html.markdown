---
layout: "aws"
page_title: "AWS: cloudwatch_metric_alarm"
sidebar_current: "docs-aws-resource-cloudwatch-metric-alarm"
description: |-
  Provides a CloudWatch Metric Alarm resource.
---

# aws_cloudwatch_metric_alarm

Provides a CloudWatch Metric Alarm resource.

## Example Usage

```hcl
resource "aws_cloudwatch_metric_alarm" "foobar" {
  alarm_name                = "terraform-test-foobar5"
  comparison_operator       = "GreaterThanOrEqualToThreshold"
  evaluation_periods        = "2"
  metric_name               = "CPUUtilization"
  namespace                 = "AWS/EC2"
  period                    = "120"
  statistic                 = "Average"
  threshold                 = "80"
  alarm_description         = "This metric monitors ec2 cpu utilization"
  insufficient_data_actions = []
}
```

## Example in Conjunction with Scaling Policies

```hcl
resource "aws_autoscaling_policy" "bat" {
  name                   = "foobar3-terraform-test"
  scaling_adjustment     = 4
  adjustment_type        = "ChangeInCapacity"
  cooldown               = 300
  autoscaling_group_name = "${aws_autoscaling_group.bar.name}"
}

resource "aws_cloudwatch_metric_alarm" "bat" {
  alarm_name          = "terraform-test-foobar5"
  comparison_operator = "GreaterThanOrEqualToThreshold"
  evaluation_periods  = "2"
  metric_name         = "CPUUtilization"
  namespace           = "AWS/EC2"
  period              = "120"
  statistic           = "Average"
  threshold           = "80"

  dimensions = {
    AutoScalingGroupName = "${aws_autoscaling_group.bar.name}"
  }

  alarm_description = "This metric monitors ec2 cpu utilization"
  alarm_actions     = ["${aws_autoscaling_policy.bat.arn}"]
}
```

## Example with an Expression

```hcl
resource "aws_cloudwatch_metric_alarm" "foobar" {
  alarm_name                = "terraform-test-foobar%d"
  comparison_operator       = "GreaterThanOrEqualToThreshold"
  evaluation_periods        = "2"
	threshold                 = "0.1"
  alarm_description         = "Request error rate has exceeded 10%"
  insufficient_data_actions = []
	metric_query {
		id = "e1"
		expression = "m2/m1*100"
		label = "Error Rate"
		return_data = "true"
	}
	metric_query {
		id = "m1"
		metric {
			metric_name = "RequestCount"
			namespace   = "AWS/ApplicationELB"
			period      = "120"
			stat        = "Sum"
			unit        = "Count"
			dimensions = {
				LoadBalancer = "app/web"
			}
		}
	}
	metric_query {
		id = "m2"
		metric {
			metric_name = "HTTPCode_ELB_5XX_Count"
			namespace   = "AWS/ApplicationELB"
			period      = "120"
			stat        = "Sum"
			unit        = "Count"
			dimensions = {
				LoadBalancer = "app/web"
			}
		}
	}
}
```

~> **NOTE:**  You cannot create a metric alarm consisting of both `statistic` and `extended_statistic` parameters.
You must choose one or the other

## Argument Reference

See [related part of AWS Docs](https://docs.aws.amazon.com/AmazonCloudWatch/latest/APIReference/API_PutMetricAlarm.html)
for details about valid values.

The following arguments are supported:

* `alarm_name` - (Required) The descriptive name for the alarm. This name must be unique within the user's AWS account
* `arn` - The ARN of the cloudwatch metric alarm.
* `comparison_operator` - (Required) The arithmetic operation to use when comparing the specified Statistic and Threshold. The specified Statistic value is used as the first operand. Either of the following is supported: `GreaterThanOrEqualToThreshold`, `GreaterThanThreshold`, `LessThanThreshold`, `LessThanOrEqualToThreshold`.
* `evaluation_periods` - (Required) The number of periods over which data is compared to the specified threshold.
* `metric_name` - (Optional) The name for the alarm's associated metric.
  See docs for [supported metrics](https://docs.aws.amazon.com/AmazonCloudWatch/latest/DeveloperGuide/CW_Support_For_AWS.html).
* `namespace` - (Optional) The namespace for the alarm's associated metric. See docs for the [list of namespaces](https://docs.aws.amazon.com/AmazonCloudWatch/latest/DeveloperGuide/aws-namespaces.html).
  See docs for [supported metrics](https://docs.aws.amazon.com/AmazonCloudWatch/latest/DeveloperGuide/CW_Support_For_AWS.html).
* `period` - (Optional) The period in seconds over which the specified `statistic` is applied.
* `statistic` - (Optional) The statistic to apply to the alarm's associated metric.
   Either of the following is supported: `SampleCount`, `Average`, `Sum`, `Minimum`, `Maximum`
* `threshold` - (Required) The value against which the specified statistic is compared.
* `actions_enabled` - (Optional) Indicates whether or not actions should be executed during any changes to the alarm's state. Defaults to `true`.
* `alarm_actions` - (Optional) The list of actions to execute when this alarm transitions into an ALARM state from any other state. Each action is specified as an Amazon Resource Number (ARN).
* `alarm_description` - (Optional) The description for the alarm.
* `datapoints_to_alarm` - (Optional) The number of datapoints that must be breaching to trigger the alarm.
* `dimensions` - (Optional) The dimensions for the alarm's associated metric.  For the list of available dimensions see the AWS documentation [here](http://docs.aws.amazon.com/AmazonCloudWatch/latest/DeveloperGuide/CW_Support_For_AWS.html).
* `insufficient_data_actions` - (Optional) The list of actions to execute when this alarm transitions into an INSUFFICIENT_DATA state from any other state. Each action is specified as an Amazon Resource Number (ARN).
* `ok_actions` - (Optional) The list of actions to execute when this alarm transitions into an OK state from any other state. Each action is specified as an Amazon Resource Number (ARN).
* `unit` - (Optional) The unit for the alarm's associated metric.
* `extended_statistic` - (Optional) The percentile statistic for the metric associated with the alarm. Specify a value between p0.0 and p100.
* `treat_missing_data` - (Optional) Sets how this alarm is to handle missing data points. The following values are supported: `missing`, `ignore`, `breaching` and `notBreaching`. Defaults to `missing`.
* `evaluate_low_sample_count_percentiles` - (Optional) Used only for alarms
based on percentiles. If you specify `ignore`, the alarm state will not
change during periods with too few data points to be statistically significant.
If you specify `evaluate` or omit this parameter, the alarm will always be
evaluated and possibly change state no matter how many data points are available.
The following values are supported: `ignore`, and `evaluate`.
* `metric_query` (Optional) Enables you to create an alarm based on a metric math expression. You may specify at most 20.

~> **NOTE:**  If you specify at least one `metric_query`, you may not specify a `metric_name`, `namespace`, `period` or `statistic`. If you do not specify a `metric_query`, you must specify each of these (although you may use `extended_statistic` instead of `statistic`).

### Nested fields

#### `metric_query`

* `id` - (Required) A short name used to tie this object to the results in the response. If you are performing math expressions on this set of data, this name represents that data and can serve as a variable in the mathematical expression. The valid characters are letters, numbers, and underscore. The first character must be a lowercase letter. 
* `expression` - (Optional) The math expression to be performed on the returned data, if this object is performing a math expression. This expression can use the id of the other metrics to refer to those metrics, and can also use the id of other expressions to use the result of those expressions. For more information about metric math expressions, see Metric Math Syntax and Functions in the [Amazon CloudWatch User Guide](https://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/using-metric-math.html#metric-math-syntax). 
* `label` - (Optional) A human-readable label for this metric or expression. This is especially useful if this is an expression, so that you know what the value represents.
* `return_data` (Optional) Specify exactly one `metric_query` to be `true` to use that `metric_query` result as the alarm.
* `metric` (Optional) The metric to be returned, along with statistics, period, and units. Use this parameter only if this object is retrieving a metric and not performing a math expression on returned data.

~> **NOTE:**  You must specify either `metric` or `expression`. Not both.

#### `metric`

* `dimensions` - (Optional) The dimensions for this metric.  For the list of available dimensions see the AWS documentation [here](http://docs.aws.amazon.com/AmazonCloudWatch/latest/DeveloperGuide/CW_Support_For_AWS.html).
* `metric_name` - (Required) The name for this metric.
  See docs for [supported metrics](https://docs.aws.amazon.com/AmazonCloudWatch/latest/DeveloperGuide/CW_Support_For_AWS.html).
* `namespace` - (Required) The namespace for this metric. See docs for the [list of namespaces](https://docs.aws.amazon.com/AmazonCloudWatch/latest/DeveloperGuide/aws-namespaces.html).
  See docs for [supported metrics](https://docs.aws.amazon.com/AmazonCloudWatch/latest/DeveloperGuide/CW_Support_For_AWS.html).
* `period` - (Required) The period in seconds over which the specified `stat` is applied.
* `stat` - (Required) The statistic to apply to this metric.
   Either of the following is supported: `SampleCount`, `Average`, `Sum`, `Minimum`, `Maximum`
* `unit` - (Optional) The unit for this metric.


## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the health check

## Import

Cloud Metric Alarms can be imported using the `alarm_name`, e.g.

```
$ terraform import aws_cloudwatch_metric_alarm.test alarm-12345
```
