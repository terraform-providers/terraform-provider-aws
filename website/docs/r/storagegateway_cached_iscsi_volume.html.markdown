---
layout: "aws"
page_title: "AWS: aws_storagegateway_cached_iscsi_volume"
sidebar_current: "docs-aws-resource-storagegateway-cached-iscsi-volume"
description: |-
  Manages an AWS Storage Gateway cached iSCSI volume
---

# aws_storagegateway_cached_iscsi_volume

Manages an AWS Storage Gateway cached iSCSI volume.

~> **NOTE:** The gateway must have cache added (e.g. via the [`aws_storagegateway_cache`](/docs/providers/aws/r/storagegateway_cache.html) resource) before creating volumes otherwise the Storage Gateway API will return an error.

~> **NOTE:** The gateway must have an upload buffer added (e.g. via the [`aws_storagegateway_upload_buffer`](/docs/providers/aws/r/storagegateway_upload_buffer.html) resource) before the volume is operational to clients, however the Storage Gateway API will allow volume creation without error in that case and return volume status as `UPLOAD BUFFER NOT CONFIGURED`.

## Example Usage

~> **NOTE:** These examples are referencing the [`aws_storagegateway_cache`](/docs/providers/aws/r/storagegateway_cache.html) resource `gateway_arn` attribute to ensure Terraform properly adds cache before creating the volume. If you are not using this method, you may need to declare an expicit dependency (e.g. via `depends_on = ["aws_storagegateway_cache.example"]`) to ensure proper ordering.

### Create Empty Cached iSCSI Volume

```hcl
resource "aws_storagegateway_cached_iscsi_volume" "example" {
  gateway_arn          = "${aws_storagegateway_cache.example.gateway_arn}"
  network_interface_id = "${aws_instance.example.private_ip}"
  target_name          = "example"
  volume_size_in_bytes = 5368709120 # 5 GB
}
```

### Create Cached iSCSI Volume From Snapshot

```hcl
resource "aws_storagegateway_cached_iscsi_volume" "example" {
  gateway_arn          = "${aws_storagegateway_cache.example.gateway_arn}"
  network_interface_id = "${aws_instance.example.private_ip}"
  snapshot_id          = "${aws_ebs_snapshot.example.id}"
  target_name          = "example"
  volume_size_in_bytes = "${aws_ebs_snapshot.example.volume_size * 1024 * 1024 * 1024}"
}
```

### Create Cached iSCSI Volume From Source Volume

```hcl
resource "aws_storagegateway_cached_iscsi_volume" "example" {
  gateway_arn          = "${aws_storagegateway_cache.example.gateway_arn}"
  network_interface_id = "${aws_instance.example.private_ip}"
  source_volume_arn    = "${aws_storagegateway_cached_iscsi_volume.existing.arn}"
  target_name          = "example"
  volume_size_in_bytes = "${aws_storagegateway_cached_iscsi_volume.existing.volume_size_in_bytes}"
}
```

## Argument Reference

The following arguments are supported:

* `gateway_arn` - (Required) The Amazon Resource Name (ARN) of the gateway.
* `network_interface_id` - (Required) The network interface of the gateway on which to expose the iSCSI target. Only IPv4 addresses are accepted.
* `target_name` - (Required) The name of the iSCSI target used by initiators to connect to the target and as a suffix for the target ARN. The target name must be unique across all volumes of a gateway.
* `volume_size_in_bytes` - (Required) The size of the volume in bytes.
* `snapshot_id` - (Optional) The snapshot ID of the snapshot to restore as the new cached volume. e.g. `snap-1122aabb`.
* `source_volume_arn` - (Optional) The ARN for an existing volume. Specifying this ARN makes the new volume into an exact copy of the specified existing volume's latest recovery point. The `volume_size_in_bytes` value for this new volume must be equal to or larger than the size of the existing volume, in bytes.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - Volume Amazon Resource Name (ARN), e.g. `arn:aws:storagegateway:us-east-1:123456789012:gateway/sgw-12345678/volume/vol-12345678`.
* `chap_enabled` - Whether mutual CHAP is enabled for the iSCSI target.
* `id` - Volume Amazon Resource Name (ARN), e.g. `arn:aws:storagegateway:us-east-1:123456789012:gateway/sgw-12345678/volume/vol-12345678`.
* `lun_number` - Logical disk number.
* `network_interface_port` - The port used to communicate with iSCSI targets.
* `target_arn` - Target Amazon Resource Name (ARN), e.g. `arn:aws:storagegateway:us-east-1:123456789012:gateway/sgw-12345678/target/iqn.1997-05.com.amazon:TargetName`.
* `volume_arn` - Volume Amazon Resource Name (ARN), e.g. `arn:aws:storagegateway:us-east-1:123456789012:gateway/sgw-12345678/volume/vol-12345678`.
* `volume_id` - Volume ID, e.g. `vol-12345678`.

## Import

`aws_storagegateway_cached_iscsi_volume` can be imported by using the volume Amazon Resource Name (ARN), e.g.

```
$ terraform import aws_storagegateway_cache.example arn:aws:storagegateway:us-east-1:123456789012:gateway/sgw-12345678/volume/vol-12345678
```
