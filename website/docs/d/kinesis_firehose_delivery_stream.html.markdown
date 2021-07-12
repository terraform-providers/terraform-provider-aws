---
subcategory: "Kinesis Firehose"
layout: "aws"
page_title: "AWS: aws_kinesis_firehose_delivery_stream"
description: |-
  Provides an AWS Kinesis Firehose Delivery Stream data source.
---

# Data Source: aws_kinesis_firehose_delivery_stream

Use this data source to get information about a Kinesis Firehose Delivery Stream for use in other resources.

For more details, see the [Amazon Kinesis Firehose Documentation][1].

## Example Usage

```terraform
data "aws_kinesis_firehose_delivery_stream" "stream" {
  name = "stream-name"
}
```

## Argument Reference

* `name` - (Required) The name of the Kinesis Stream.

## Attributes Reference

`id` is set to the Amazon Resource Name (ARN) of the Kinesis Stream. In addition, the following attributes
are exported:

* `arn` - The Amazon Resource Name (ARN) of the Kinesis Stream (same as id).
* `name` - The name of the Kinesis Stream.

[1]: https://aws.amazon.com/documentation/firehose/
