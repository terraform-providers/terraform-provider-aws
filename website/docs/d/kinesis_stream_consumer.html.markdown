---
subcategory: "Kinesis"
layout: "aws"
page_title: "AWS: aws_kinesis_stream_consumer"
description: |-
  Provides details about a Kinesis Stream Consumer.
---

# Data Source: aws_kinesis_stream_consumer

Provides details about a Kinesis Stream Consumer.

For more details, see the [Amazon Kinesis Stream Consumer Documentation][1].

## Example Usage

```terraform
data "aws_kinesis_stream_consumer" "example" {
  name       = "example-consumer"
  stream_arn = aws_kinesis_stream.example.arn
}
```

## Argument Reference

* `arn` - (Optional) Amazon Resource Name (ARN) of the stream consumer.
* `name` - (Optional) Name of the stream consumer.
* `stream_arn` - (Required) Amazon Resource Name (ARN) of the data stream the consumer is registered with.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `creation_timestamp` - Approximate timestamp in [RFC3339 format](https://tools.ietf.org/html/rfc3339#section-5.8) of when the stream consumer was created.
* `id` - Amazon Resource Name (ARN) of the stream consumer.
* `status` - The current status of the stream consumer.

[1]: https://docs.aws.amazon.com/streams/latest/dev/amazon-kinesis-consumers.html
