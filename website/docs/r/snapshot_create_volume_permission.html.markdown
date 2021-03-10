---
subcategory: "EC2"
layout: "aws"
page_title: "AWS: aws_snapshot_create_volume_permission"
description: |-
  Adds create volume permission to an EBS Snapshot
---

# Resource: aws_snapshot_create_volume_permission

Adds permission to create volumes off of a given EBS Snapshot.

## Example Usage

### Share with Account ID

```hcl
resource "aws_snapshot_create_volume_permission" "example_perm" {
  snapshot_id = aws_ebs_snapshot.example_snapshot.id
  account_id  = "12345678"
}

resource "aws_ebs_volume" "example" {
  availability_zone = "us-west-2a"
  size              = 40
}

resource "aws_ebs_snapshot" "example_snapshot" {
  volume_id = aws_ebs_volume.example.id
}
```

### Public Snapshot

```hcl
resource "aws_snapshot_create_volume_permission" "example_perm" {
  snapshot_id = "${aws_ebs_snapshot.example_snapshot.id}"
  group       = "all"
}

resource "aws_ebs_volume" "example" {
  availability_zone = "us-west-2a"
  size              = 40
}

resource "aws_ebs_snapshot" "example_snapshot" {
  volume_id = "${aws_ebs_volume.example.id}"
}
```

## Argument Reference

The following arguments are supported:


  * `snapshot_id` - (Required) A snapshot ID.
  * `account_id` - (Optional) An AWS Account ID to add create volume permissions. Conflicts with `group`.
  * `group` - (Optional) Only takes the value `all`. Conflicts with `account_id`.

## Attributes Reference

In addition to all arguments above, the following attribute is exported:


  * `id` - A combination of "`snapshot_id`-`account_id`-`group`".
