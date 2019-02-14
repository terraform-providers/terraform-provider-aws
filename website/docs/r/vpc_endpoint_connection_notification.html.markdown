---
layout: "aws"
page_title: "AWS: aws_vpc_endpoint_connection_notification"
sidebar_current: "docs-aws-resource-vpc-endpoint-connection-notification"
description: |-
  Provides a VPC Endpoint connection notification resource.
---

# aws_vpc_endpoint_connection_notification

Provides a VPC Endpoint connection notification resource.
Connection notifications notify subscribers of VPC Endpoint events.

## Example Usage

```hcl
resource "aws_sns_topic" "topic" {
  name = "vpce-notification-topic"

  policy = <<POLICY
{
    "Version":"2012-10-17",
    "Statement":[{
        "Effect": "Allow",
        "Principal": {
            "Service": "vpce.amazonaws.com"
        },
        "Action": "SNS:Publish",
        "Resource": "arn:aws:sns:*:*:vpce-notification-topic"
    }]
}
POLICY
}

resource "aws_vpc_endpoint_service" "foo" {
  acceptance_required        = false
  network_load_balancer_arns = ["${aws_lb.test.arn}"]
}

resource "aws_vpc_endpoint_connection_notification" "foo" {
  vpc_endpoint_service_id     = "${aws_vpc_endpoint_service.foo.id}"
  connection_notification_arn = "${aws_sns_topic.topic.arn}"
  connection_events           = ["Accept", "Reject"]
}
```

## Argument Reference

The following arguments are supported:

* `vpc_endpoint_service_id` - (Optional) The ID of the VPC Endpoint Service to receive notifications for.
* `vpc_endpoint_id` - (Optional) The ID of the VPC Endpoint to receive notifications for.
* `connection_notification_arn` - (Required) The ARN of the SNS topic for the notifications.
* `connection_events` - (Required) One or more endpoint [events](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_CreateVpcEndpointConnectionNotification.html#API_CreateVpcEndpointConnectionNotification_RequestParameters) for which to receive notifications.

~> **NOTE:** One of `vpc_endpoint_service_id` or `vpc_endpoint_id` must be specified.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the VPC connection notification.
* `state` - The state of the notification.
* `notification_type` - The type of notification.

## Import

VPC Endpoint connection notifications can be imported using the `VPC endpoint connection notification id`, e.g.

```
$ terraform import aws_vpc_endpoint_connection_notification.foo vpce-nfn-09e6ed3b4efba2263
```
