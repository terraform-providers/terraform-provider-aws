---
subcategory: "Storage Gateway"
layout: "aws"
page_title: "AWS: aws_storagegateway_tape_pool"
description: |-
  Manages an AWS Storage Gateway Tape Pool
---

# Resource: aws_storagegateway_tape_pool

Manages an AWS Storage Gateway Tape Pool.

## Example Usage

```hcl
resource "aws_storagegateway_tape_pool" "example" {
  pool_name     = "example"
  storage_class = "GLACIER"
}
```

## Argument Reference

The following arguments are supported:

* `pool_name` - (Required) The name of the new custom tape pool.
* `storage_class` - (Required) The storage class that is associated with the new custom pool. When you use your backup application to eject the tape, the tape is archived directly into the storage class that corresponds to the pool. Possible values are `DEEP_ARCHIVE` or `GLACIER`.
* `retention_lock_type` - (Required) Tape retention lock can be configured in two modes. When configured in governance mode, AWS accounts with specific IAM permissions are authorized to remove the tape retention lock from archived virtual tapes. When configured in compliance mode, the tape retention lock cannot be removed by any user, including the root AWS account. Possible values are `COMPLIANCE`, `GOVERNANCE`, and `NONE`. Default value is `NONE`.
* `retention_lock_time_in_days` - (Optional) Tape retention lock time is set in days. Tape retention lock can be enabled for up to 100 years (36,500 days). Default value is 0.
* `tags` - (Optional) Key-value map of resource tags

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - Volume Amazon Resource Name (ARN), e.g. `aws_storagegateway_tape_pool.example arn:aws:storagegateway:us-east-1:123456789012:tapepool/pool-12345678`.

## Import

`aws_storagegateway_tape_pool` can be imported by using the volume Amazon Resource Name (ARN), e.g.

```
$ terraform import aws_storagegateway_tape_pool.example arn:aws:storagegateway:us-east-1:123456789012:tapepool/pool-12345678
```
