---
layout: "aws"
page_title: "AWS: aws_storagegateway_upload_buffer"
sidebar_current: "docs-aws-resource-storagegateway-upload-buffer"
description: |-
  Manages an AWS Storage Gateway upload buffer
---

# aws_storagegateway_upload_buffer

Manages an AWS Storage Gateway upload buffer.

~> **NOTE:** The Storage Gateway API provides no method to remove an upload buffer disk. Destroying this Terraform resource does not perform any Storage Gateway actions.

## Example Usage

```hcl
resource "aws_storagegateway_upload_buffer" "example" {
  disk_id     = "${data.aws_storagegateway_local_disk.example.id}"
  gateway_arn = "${aws_storagegateway_gateway.example.arn}"
}
```

## Argument Reference

The following arguments are supported:

* `disk_id` - (Required) Local disk identifier. For example, `pci-0000:03:00.0-scsi-0:0:0:0`.
* `gateway_arn` - (Required) The Amazon Resource Name (ARN) of the gateway.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - Combined gateway Amazon Resource Name (ARN) and local disk identifier.

## Import

`aws_storagegateway_upload_buffer` can be imported by using the gateway Amazon Resource Name (ARN) and local disk identifier separated with a colon (`:`), e.g.

```
$ terraform import aws_storagegateway_upload_buffer.example arn:aws:storagegateway:us-east-1:123456789012:gateway/sgw-12345678:pci-0000:03:00.0-scsi-0:0:0:0
```
