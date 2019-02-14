---
layout: "aws"
page_title: "AWS: aws_cloudwatch_dashboard"
sidebar_current: "docs-aws-resource-cloudwatch-dashboard"
description: |-
  Provides a CloudWatch Dashboard resource.
---

# aws_cloudwatch_dashboard

Provides a CloudWatch Dashboard resource.

## Example Usage

```hcl
resource "aws_cloudwatch_dashboard" "main" {
  dashboard_name = "my-dashboard"

  dashboard_body = <<EOF
 {
   "widgets": [
       {
          "type":"metric",
          "x":0,
          "y":0,
          "width":12,
          "height":6,
          "properties":{
             "metrics":[
                [
                   "AWS/EC2",
                   "CPUUtilization",
                   "InstanceId",
                   "i-012345"
                ]
             ],
             "period":300,
             "stat":"Average",
             "region":"us-east-1",
             "title":"EC2 Instance CPU"
          }
       },
       {
          "type":"text",
          "x":0,
          "y":7,
          "width":3,
          "height":3,
          "properties":{
             "markdown":"Hello world"
          }
       }
   ]
 }
 EOF
}
```

## Argument Reference

The following arguments are supported:

* `dashboard_name` - (Required) The name of the dashboard.
* `dashboard_body` - (Required) The detailed information about the dashboard, including what widgets are included and their location on the dashboard. You can read more about the body structure in the [documentation](https://docs.aws.amazon.com/AmazonCloudWatch/latest/APIReference/CloudWatch-Dashboard-Body-Structure.html).

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `dashboard_arn` - The Amazon Resource Name (ARN) of the dashboard.

## Import

CloudWatch dashboards can be imported using the `dashboard_name`, e.g.

```
$ terraform import aws_cloudwatch_dashboard.sample <dashboard_name>
```
