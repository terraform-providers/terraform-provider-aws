---
subcategory: "EC2"
layout: "aws"
page_title: "AWS: aws_ebs_snapshot_import"
description: |-
  Provides an elastic block storage snapshot import resource.
---

# Resource: aws_ebs_snapshot_import

Imports a disk image from S3 as a Snapshot.

## Example Usage

```hcl
resource "aws_ebs_snapshot_import" "example" {
  disk_container {
    format = "VHD"
    user_bucket {
      s3_bucket = "disk-images"
      s3_key    = "source.vhd"
    }
  }

  role_name = "disk-image-import"

  tags = {
    Name = "HelloWorld"
  }
}
```

## Argument Reference


The following arguments are supported:

* `client_data` - (Optional) The client-specific data. Detailed below.
* `description` - (Optional) A description of what the snapshot is.
* `disk_container` - (Required) Information about the disk container. Detailed below.
* `encrypted` - (Optional) Whether the snapshot is encrypted.
* `kms_key_id` - (Optional) The ARN for the KMS encryption key.
* `role_name` - (Optional) The name of the IAM Role the VM Import/Export service will assume. This role needs certain permissions. See https://docs.aws.amazon.com/vm-import/latest/userguide/vmie_prereqs.html#vmimport-role. Default: `vmimport`
* `tags` - (Optional) A map of tags to assign to the snapshot

### client_data Configuration Block

* `comment` - (Optional) A user-defined comment about the disk upload.
* `upload_start` - (Optional) The time that the disk upload starts.
* `upload_end` - (Optional) The time that the disk upload ends.
* `upload_size` - (Optional) The size of the uploaded disk image, in GiB.

### disk_container Configuration Block

* `description` - (Optional) The description of the disk image being imported.
* `format` - (Required) The format of the disk image being imported. One of `VHD` or `VMDK`.
* `url` - (Optional) The URL to the Amazon S3-based disk image being imported. It can either be a https URL (https://..) or an Amazon S3 URL (s3://..). One of `url` or `user_bucket` must be set.
* `user_bucket` - (Optional) The Amazon S3 bucket for the disk image. One of `url` or `user_bucket` must be set. Detailed below.

### user_bucket Configuration Block

* `s3_bucket` - The name of the Amazon S3 bucket where the disk image is located.
* `s3_key` - The file name of the disk image.

### Timeouts

`aws_ebs_snapshot_import` provides the following
[Timeouts](/docs/configuration/resources.html#timeouts) configuration options:

- `create` - (Default `60 minutes`) Used for importing the EBS snapshot
- `delete` - (Default `10 minutes`) Used for deleting the EBS snapshot

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - Amazon Resource Name (ARN) of the EBS Snapshot.
* `id` - The snapshot ID (e.g. snap-59fcb34e).
* `owner_id` - The AWS account ID of the EBS snapshot owner.
* `owner_alias` - Value from an Amazon-maintained list (`amazon`, `aws-marketplace`, `microsoft`) of snapshot owners.
* `volume_size` - The size of the drive in GiBs.
* `data_encryption_key_id` - The data encryption key identifier for the snapshot.
* `tags` - A map of tags for the snapshot.

