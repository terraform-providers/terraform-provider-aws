---
subcategory: "Redshift"
layout: "aws"
page_title: "AWS: aws_redshift_scheduled_action"
description: |-
  Provides a Redshift Scheduled Action resource.
---

# Resource: aws_redshift_scheduled_action

## Example Pause Cluster

```hcl
resource "aws_iam_role" "redshift_scheduled_action" {
  name               = "redshift_scheduled_action"
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": [
          "scheduler.redshift.amazonaws.com"
        ]
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_iam_policy" "redshift_scheduled_action" {
  name   = "redshift_scheduled_action"
  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
      {
          "Sid": "VisualEditor0",
          "Effect": "Allow",
          "Action": [
              "redshift:PauseCluster",
              "redshift:ResumeCluster",
              "redshift:ResizeCluster"
          ],
          "Resource": "*"
      }
  ]
}
EOF
}

resource "aws_iam_role_policy_attachment" "redshift_scheduled_action" {
  policy_arn = aws_iam_policy.redshift_scheduled_action.arn
  role       = aws_iam_role.redshift_scheduled_action.name
}

resource "aws_redshift_scheduled_action" "default" {
  name     = "tf-redshift-scheduled-action"
  schedule = "cron(00 23 * * ? *)"
  iam_role = aws_iam_role.redshift_scheduled_action.arn

  target_action {
    action             = "PauseCluster"
    cluster_identifier = "tf-redshift001"
  }
}
```

## Example Resize Cluster

```hcl
resource "aws_redshift_scheduled_action" "default" {
  name     = "tf-redshift-scheduled-action"
  schedule = "cron(00 23 * * ? *)"
  iam_role = aws_iam_role.redshift_scheduled_action.arn

  target_action {
    action             = "ResizeCluster"
    cluster_identifier = "tf-redshift001"
    classic            = false
    cluster_type       = "multi-node"
    node_type          = "dc1.large"
    number_of_nodes    = 2
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required, Forces new resource) The scheduled action name.
* `description` - (Optional) The description of the scheduled action.
* `active` - (Optional) Whether to enable the scheduled action. Default is `true` .
* `start_time` - (Optional) The start time in UTC when the schedule is active, in UTC RFC3339 format(for example, YYYY-MM-DDTHH:MM:SSZ).
* `end_time` - (Optional) The end time in UTC when the schedule is active, in UTC RFC3339 format(for example, YYYY-MM-DDTHH:MM:SSZ).
* `schedule` - (Required) The schedule of action. The schedule is defined format of "at expression" or "cron expression", for example `at(2016-03-04T17:27:00)` or `cron(0 10 ? * MON *)`. See [Scheduled Action](https://docs.aws.amazon.com/redshift/latest/APIReference/API_ScheduledAction.html) for more information.
* `iam_role` - (Required) The IAM role to assume to run the scheduled action.
* `target_action` - (Required) Target action, documented below.

### Nested Blocks

#### `target_action`

* `action` - (Required) The action type of the scheduled action. Possible values are `PauseCluster`,  `ResumeCluster` and `ResizeCluster`.
* `cluster_identifier` - (Required) The target identifier for the redshift cluster of the scheduled action.
* `classic` - (Optional) Indicate resize operation is using the classic resize process. Default is `false`.
* `cluster_type` - (Optional)　The new cluster type for the specified cluster.
* `node_type` - (Optional) The new node type for the nodes you are addingThe new node type for the nodes you are adding.
* `number_of_nodes` - (Optional) The new number of nodes for the cluster.

## Import

Redshift Scheduled Action can be imported using the `name`, e.g.

```
$ terraform import aws_redshift_scheduled_action.default tf-redshift-scheduled-action
```
