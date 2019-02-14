---
layout: "aws"
page_title: "AWS: aws_storagegateway_local_disk"
sidebar_current: "docs-aws-datasource-storagegateway-local-disk"
description: |-
  Retrieve information about a Storage Gateway local disk
---

# Data Source: aws_storagegateway_local_disk

Retrieve information about a Storage Gateway local disk. The disk identifier is useful for adding the disk as a cache or upload buffer to a gateway.

## Example Usage

```hcl
data "aws_storagegateway_local_disk" "test" {
  disk_path   = "${aws_volume_attachment.test.device_name}"
  gateway_arn = "${aws_storagegateway_gateway.test.arn}"
}
```

## Argument Reference

* `gateway_arn` - (Required) The Amazon Resource Name (ARN) of the gateway.
* `disk_node` - (Optional) The device node of the local disk to retrieve. For example, `/dev/sdb`.
* `disk_path` - (Optional) The device path of the local disk to retrieve. For example, `/dev/xvdb` or `/dev/nvme1n1`.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `disk_id` - The disk identifier. e.g. `pci-0000:03:00.0-scsi-0:0:0:0`
* `id` - The disk identifier. e.g. `pci-0000:03:00.0-scsi-0:0:0:0`
