---
subcategory: "EC2"
layout: "aws"
page_title: "AWS: aws_ebs_snapshot"
description: |-
  Provides an elastic block storage snapshot resource.
---

# Resource: aws_ebs_snapshot

Creates a Snapshot of an EBS Volume.

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
  volume_id = aws_ebs_volume.example.id

  tags = {
    Name = "HelloWorld_snap"
  }
}
```

## Argument Reference

The following arguments are supported:

* `volume_id` - (Required) The Volume ID of which to make a snapshot.
* `description` - (Optional) A description of what the snapshot is.
* `tags` - (Optional) A map of tags to assign to the snapshot

### Timeouts

`aws_ebs_snapshot` provides the following
[Timeouts](/docs/configuration/resources.html#timeouts) configuration options:

- `create` - (Default `10 minutes`) Used for creating the ebs snapshot
- `delete` - (Default `10 minutes`) Used for deleting the ebs snapshot

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - Amazon Resource Name (ARN) of the EBS Snapshot.
* `id` - The snapshot ID (e.g. snap-59fcb34e).
* `owner_id` - The AWS account ID of the EBS snapshot owner.
* `owner_alias` - Value from an Amazon-maintained list (`amazon`, `aws-marketplace`, `microsoft`) of snapshot owners.
* `encrypted` - Whether the snapshot is encrypted.
* `volume_size` - The size of the drive in GiBs.
* `kms_key_id` - The ARN for the KMS encryption key.
* `data_encryption_key_id` - The data encryption key identifier for the snapshot.
* `tags` - A map of tags for the snapshot.

## Import

EBS Snapshot can be imported using the `id`, e.g.

```
$ terraform import aws_ebs_snapshot.id snap-049df61146c4d7901
```
