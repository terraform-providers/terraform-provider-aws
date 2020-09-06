---
subcategory: "Autoscaling"
layout: "aws"
page_title: "AWS: aws_autoscaling_attachment"
description: |-
  Provides an AutoScaling Group Attachment resource.
---

# Resource: aws_autoscaling_attachment

Provides an AutoScaling Attachment resource.

~> **NOTE on AutoScaling Groups and ASG Attachments:** Terraform currently provides
both a standalone ASG Attachment resource (describing an ASG attached to
an ELB or ALB), and an [AutoScaling Group resource](autoscaling_group.html) with
`load_balancers` and `target_group_arns` defined in-line. At this time you can use an ASG with in-line
`load balancers` or `target_group_arns` in conjunction with an ASG Attachment resource, however, to prevent
unintended resource updates, the `aws_autoscaling_group` resource must be configured
to ignore changes to the `load_balancers` and `target_group_arns` arguments within a [`lifecycle` configuration block](/docs/configuration/resources.html#lifecycle-lifecycle-customizations).

## Example Usage

```hcl
# Create a new load balancer attachment
resource "aws_autoscaling_attachment" "asg_attachment_bar" {
  autoscaling_group_name = aws_autoscaling_group.asg.id
  elb                    = aws_elb.bar.id
}
```

```hcl
# Create a new ALB Target Group attachment
resource "aws_autoscaling_attachment" "asg_attachment_bar" {
  autoscaling_group_name = aws_autoscaling_group.asg.id
  alb_target_group_arn   = aws_alb_target_group.test.arn
}
```

## With An AutoScaling Group Resource

```hcl
resource "aws_autoscaling_group" "asg" {
  # ... other configuration ...

  lifecycle {
    ignore_changes = [load_balancers, target_group_arns]
  }
}

resource "aws_autoscaling_attachment" "asg_attachment_bar" {
  autoscaling_group_name = aws_autoscaling_group.asg.id
  elb                    = aws_elb.test.id
}
```

## Argument Reference

The following arguments are supported:

* `autoscaling_group_name` - (Required) Name of ASG to associate with the ELB.
* `elb` - (Optional) The name of the ELB.
* `alb_target_group_arn` - (Optional) The ARN of an ALB Target Group.

