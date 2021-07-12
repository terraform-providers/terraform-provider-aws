---
subcategory: "Quantum Ledger Database (QLDB) Stream"
layout: "aws"
page_title: "AWS: aws_qldb_stream"
description: |-
  Provides a QLDB Stream resource.
---

# Resource: aws_qldb_stream

Provides an AWS Quantum Ledger Database (QLDB) Stream resource

~> **NOTE:** Deletion protection is enabled by default. To successfully delete this resource via Terraform, `deletion_protection = false` must be applied before attempting deletion.

## Example Usage

```terraform
resource "aws_qldb_stream" "sample-ledger-stream" {
  ledger_name          = "existing-ledger-name"
  stream_name          = "sample-ledger-stream"
  role_arn             = "sample-role-arn"
  inclusive_start_time = "2021-01-01T00:00:00Z"

  kinesis_configuration = {
    aggegation_enabled = false
    stream_arn         = "arn:aws:kinesis:us-east-1:xxxxxxxxxxxx:stream/example-kinesis-stream"
  }

  tags = {
    "example" = "tag"
  }
}
```

## Argument Reference

The following arguments are supported:
* `deletion_protection` - (Optional) The deletion protection for the QLDB Stream instance. By default it is `true`. To delete this resource via Terraform, this value must be configured to `false` and applied first before attempting deletion.
* `exclusive_end_time` - (Optional) The exclusive date and time that specifies when the stream ends. If you don't define this parameter, the stream runs indefinitely until you cancel it.  It must be in ISO 8601 date and time format and in Universal Coordinated Time (UTC). For example: 2019-06-13T21:36:34Z.
* `inclusive_start_time` - (Required) The inclusive start date and time from which to start streaming journal data. This parameter must be in ISO 8601 date and time format and in Universal Coordinated Time (UTC). For example: 2019-06-13T21:36:34Z.  This cannot be in the future and must be before ExclusiveEndTime.  If you provide a value that is before the ledger's CreationDateTime, QLDB effectively defaults it to the ledger's CreationDateTime.
* `kinesis_configuration` - (Required) The configuration settings of the Kinesis Data Streams destination for your stream request.  The `aggregation_enabled` property is optional.  The `stream_arn` is required.
* `ledger_name` - (Required) The name of the QLDB ledger.  Pattern: (?!^.*--)(?!^[0-9]+$)(?!^-)(?!.*-$)^[A-Za-z0-9-]+$
* `role_arn` - (Required) The Amazon Resource Name (ARN) of the IAM role that grants QLDB permissions for a journal stream to write data records to a Kinesis Data Streams resource.
* `stream_name` - (Required) The name that you want to assign to the QLDB journal stream. User-defined names can help identify and indicate the purpose of a stream.  Your stream name must be unique among other active streams for a given ledger. Stream names have the same naming constraints as ledger names, as defined in Quotas in Amazon QLDB in the Amazon QLDB Developer Guide.  Pattern: `(?!^.*--)(?!^[0-9]+$)(?!^-)(?!.*-$)^[A-Za-z0-9-]+$`
* `tags` - (Optional) Key-value map of resource tags

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The Name of the QLDB Ledger
* `arn` - The ARN of the QLDB Ledger

## Import

QLDB Streams cannot currently be imported.

