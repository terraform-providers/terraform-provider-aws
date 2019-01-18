---
layout: "aws"
page_title: "AWS: aws_ebs_snapshot_copy"
sidebar_current: "docs-aws-resource-ebs-snapshot-copy"
description: |-
  Duplicates an existing Amazon snapshot
---

# aws_ebs_snapshot_copy

Creates a Snapshot of a snapshot.

## Example Usage

```hcl
resource "aws_ebs_volume" "example" {
  availability_zone = "us-west-2a"
  size              = 40

  tags = {
    Name = "HelloWorld"
  }
}

resource "aws_ebs_snapshot" "example_snapshot" {
  volume_id = "${aws_ebs_volume.example.id}"

  tags = {
    Name = "HelloWorld_snap"
  }
}

resource "aws_ebs_snapshot_copy" "example_copy" {
  source_snapshot_id = "${aws_ebs_snapshot.example_snapshot.id}"
  source_region      = "us-west-2"

  tags = {
    Name = "HelloWorld_copy_snap"
  }
}
```

## Argument Reference

The following arguments are supported:

* `description` - (Optional) A description of what the snapshot is.
* `encrypted` - Whether the snapshot is encrypted.
* `kms_key_id` - The ARN for the KMS encryption key.
* `source_snapshot_id` The ARN for the snapshot to be copied.
* `source_region` The region of the source snapshot.
* `tags` - A mapping of tags for the snapshot.

## Attributes Reference

The following attributes are exported:

* `id` - The snapshot ID (e.g. snap-59fcb34e).
* `owner_id` - The AWS account ID of the snapshot owner.
* `owner_alias` - Value from an Amazon-maintained list (`amazon`, `aws-marketplace`, `microsoft`) of snapshot owners.
* `encrypted` - Whether the snapshot is encrypted.
* `volume_size` - The size of the drive in GiBs.
* `kms_key_id` - The ARN for the KMS encryption key.
* `data_encryption_key_id` - The data encryption key identifier for the snapshot.
* `source_snapshot_id` The ARN of the copied snapshot.
* `source_region` The region of the source snapshot.
* `tags` - A mapping of tags for the snapshot.
