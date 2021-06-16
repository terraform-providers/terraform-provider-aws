---
subcategory: "Kinesis"
layout: "aws"
page_title: "AWS: aws_kinesis_stream_consumer"
description: |-
  Manages a Kinesis Stream Consumer.
---

# Resource: aws_kinesis_stream_consumer

Provides a resource to manage a Kinesis Stream Consumer.

-> **Note:** You can register up to 20 consumers per stream. A given consumer can only be registered with one stream at a time.

For more details, see the [Amazon Kinesis Stream Consumer Documentation][1].

## Example Usage

```terraform
resource "aws_kinesis_stream" "example" {
  name        = "example-stream"
  shard_count = 1
}

resource "aws_kinesis_stream_consumer" "example" {
  name       = "example-consumer"
  stream_arn = aws_kinesis_stream.example.arn
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required, Forces new resource) Name of the stream consumer.
* `stream_arn` – (Required, Forces new resource) Amazon Resource Name (ARN) of the data stream the consumer is registered with.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - Amazon Resource Name (ARN) of the stream consumer.
* `creation_timestamp` - Approximate timestamp in [RFC3339 format](https://tools.ietf.org/html/rfc3339#section-5.8) of when the stream consumer was created.
* `id` - Amazon Resource Name (ARN) of the stream consumer.

## Import

Kinesis Stream Consumers can be imported using the Amazon Resource Name (ARN) e.g.

```
$ terraform import aws_kinesis_stream_consumer.example arn:aws:kinesis:us-west-2:123456789012:stream/example/consumer/example:1616044553
```

[1]: https://docs.aws.amazon.com/streams/latest/dev/amazon-kinesis-consumers.html