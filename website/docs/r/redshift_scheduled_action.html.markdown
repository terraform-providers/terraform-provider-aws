---
layout: "aws"
page_title: "AWS: aws_redshift_snapshot_schedule_association"
sidebar_current: "docs-aws-resource-redshift-snapshot-schedule-association"
description: |-
  Provides an Association Redshift Cluster and Snapshot Schedule resource.
---

# Resource: aws_redshift_snapshot_schedule_association

## Example Usage

```hcl
resource "aws_redshift_cluster" "default" {
  cluster_identifier = "tf-redshift-cluster"
  database_name      = "mydb"
  master_username    = "foo"
  master_password    = "Mustbe8characters"
  node_type          = "dc1.large"
  cluster_type       = "single-node"
}

resource "aws_redshift_snapshot_schedule" "default" {
	identifier = "tf-redshift-snapshot-schedule"
	definitions = [
		"rate(12 hours)",
	]
}

resource "aws_redshift_snapshot_schedule_association" "default" {
	  cluster_identifier  = "${aws_redshift_cluster.default.id}"
    schedule_identifier = "${aws_redshift_snapshot_schedule.default.id}"
}
```

## Argument Reference

The following arguments are supported:

* `cluster_identifier` - (Required, Forces new resource) The cluster identifier.
* `schedule_identifier` - (Required, Forces new resource) The snapshot schedule identifier.

## Import

Redshift Snapshot Schedule Association can be imported using the `<cluster-identifier>/<schedule-identifier>`, e.g.

```
$ terraform import aws_redshift_snapshot_schedule_association.default tf-redshift-cluster/tf-redshift-snapshot-schedule
```