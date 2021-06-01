---
subcategory: "Storage Gateway"
layout: "aws"
page_title: "AWS: aws_storagegateway_stored_iscsi_volume"
description: |-
  Manages an AWS Storage Gateway stored iSCSI volume
---

# Resource: aws_storagegateway_stored_iscsi_volume

Manages an AWS Storage Gateway stored iSCSI volume.

~> **NOTE:** The gateway must have a working storage added (e.g. via the [`aws_storagegateway_working_storage`](/docs/providers/aws/r/storagegateway_working_storage.html) resource) before the volume is operational to clients, however the Storage Gateway API will allow volume creation without error in that case and return volume status as `WORKING STORAGE NOT CONFIGURED`.

## Example Usage

### Create Empty Stored iSCSI Volume

```hcl
resource "aws_storagegateway_stored_iscsi_volume" "example" {
  gateway_arn            = aws_storagegateway_cache.example.gateway_arn
  network_interface_id   = aws_instance.example.private_ip
  target_name            = "example"
  preserve_existing_data = false
  disk_id                = data.aws_storagegateway_local_disk.test.id
}
```

### Create Stored iSCSI Volume From Snapshot

```hcl
resource "aws_storagegateway_stored_iscsi_volume" "example" {
  gateway_arn            = aws_storagegateway_cache.example.gateway_arn
  network_interface_id   = aws_instance.example.private_ip
  snapshot_id            = aws_ebs_snapshot.example.id
  target_name            = "example"
  preserve_existing_data = false
  disk_id                = data.aws_storagegateway_local_disk.test.id
}
```

## Argument Reference

The following arguments are supported:

* `gateway_arn` - (Required) The Amazon Resource Name (ARN) of the gateway.
* `network_interface_id` - (Required) The network interface of the gateway on which to expose the iSCSI target. Only IPv4 addresses are accepted.
* `target_name` - (Required) The name of the iSCSI target used by initiators to connect to the target and as a suffix for the target ARN. The target name must be unique across all volumes of a gateway.
* `disk_id` - (Required) The unique identifier for the gateway local disk that is configured as a stored volume.
* `preserve_existing_data` - (Required) Specify this field as `true` if you want to preserve the data on the local disk. Otherwise, specifying this field as false creates an empty volume.
* `snapshot_id` - (Optional) The snapshot ID of the snapshot to restore as the new stored volume. e.g. `snap-1122aabb`.
* `kms_encrypted` - (Optional) `true` to use Amazon S3 server side encryption with your own AWS KMS key, or `false` to use a key managed by Amazon S3. Optional.
* `kms_key` - (Optional) The Amazon Resource Name (ARN) of the AWS KMS key used for Amazon S3 server side encryption. This value can only be set when `kms_encrypted` is `true`.
* `tags` - (Optional) Key-value mapping of resource tags

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - Volume Amazon Resource Name (ARN), e.g. `arn:aws:storagegateway:us-east-1:123456789012:gateway/sgw-12345678/volume/vol-12345678`.
* `chap_enabled` - Whether mutual CHAP is enabled for the iSCSI target.
* `id` - Volume Amazon Resource Name (ARN), e.g. `arn:aws:storagegateway:us-east-1:123456789012:gateway/sgw-12345678/volume/vol-12345678`.
* `lun_number` - Logical disk number.
* `network_interface_port` - The port used to communicate with iSCSI targets.
* `target_arn` - Target Amazon Resource Name (ARN), e.g. `arn:aws:storagegateway:us-east-1:123456789012:gateway/sgw-12345678/target/iqn.1997-05.com.amazon:TargetName`.
* `volume_arn` - Volume Amazon Resource Name (ARN), e.g. `arn:aws:storagegateway:us-east-1:123456789012:gateway/sgw-12345678/volume/vol-12345678`.
* `volume_id` - Volume ID, e.g. `vol-12345678`.
* `volume_status` - indicates the state of the storage volume.
* `volume_type` - indicates the type of the volume.
* `volume_size_in_bytes` - The size of the data stored on the volume in bytes.
* `volume_attachment_status` - A value that indicates whether a storage volume is attached to, detached from, or is in the process of detaching from a gateway.

## Import

`aws_storagegateway_stored_iscsi_volume` can be imported by using the volume Amazon Resource Name (ARN), e.g.

```
$ terraform import aws_storagegateway_cache.example arn:aws:storagegateway:us-east-1:123456789012:gateway/sgw-12345678/volume/vol-12345678
```
