---
layout: "aws"
page_title: "AWS: aws_lambda_event_source_mapping"
sidebar_current: "docs-aws-resource-lambda-event-source-mapping"
description: |-
  Provides a Lambda event source mapping. This allows Lambda functions to get events from Kinesis, DynamoDB and SQS
---

# aws_lambda_event_source_mapping

Provides a Lambda event source mapping. This allows Lambda functions to get events from Kinesis, DynamoDB and SQS

For information about Lambda and how to use it, see [What is AWS Lambda?][1]
For information about event source mappings, see [CreateEventSourceMapping][2] in the API docs.

## Example Usage

### DynamoDB

```hcl
resource "aws_lambda_event_source_mapping" "example" {
  event_source_arn  = "${aws_dynamodb_table.example.stream_arn}"
  function_name     = "${aws_lambda_function.example.arn}"
  starting_position = "LATEST"
}
```

### Kinesis

```hcl
resource "aws_lambda_event_source_mapping" "example" {
  event_source_arn  = "${aws_kinesis_stream.example.arn}"
  function_name     = "${aws_lambda_function.example.arn}"
  starting_position = "LATEST"
}
```

### SQS

```hcl
resource "aws_lambda_event_source_mapping" "example" {
  event_source_arn = "${aws_sqs_queue.sqs_queue_test.arn}"
  function_name    = "${aws_lambda_function.example.arn}"
}
```

## Argument Reference

* `batch_size` - (Optional) The largest number of records that Lambda will retrieve from your event source at the time of invocation. Defaults to `100` for DynamoDB and Kinesis, `10` for SQS.
* `event_source_arn` - (Required) The event source ARN - can either be a Kinesis or DynamoDB stream.
* `enabled` - (Optional) Determines if the mapping will be enabled on creation. Defaults to `true`.
* `function_name` - (Required) The name or the ARN of the Lambda function that will be subscribing to events.
* `starting_position` - (Optional) The position in the stream where AWS Lambda should start reading. Must be one of `AT_TIMESTAMP` (Kinesis only), `LATEST` or `TRIM_HORIZON` if getting events from Kinesis or DynamoDB. Must not be provided if getting events from SQS. More information about these positions can be found in the [AWS DynamoDB Streams API Reference](https://docs.aws.amazon.com/amazondynamodb/latest/APIReference/API_streams_GetShardIterator.html) and [AWS Kinesis API Reference](https://docs.aws.amazon.com/kinesis/latest/APIReference/API_GetShardIterator.html#Kinesis-GetShardIterator-request-ShardIteratorType).
* `starting_position_timestamp` - (Optional) A timestamp in [RFC3339 format](https://tools.ietf.org/html/rfc3339#section-5.8) of the data record which to start reading when using `starting_position` set to `AT_TIMESTAMP`. If a record with this exact timestamp does not exist, the next later record is chosen. If the timestamp is older than the current trim horizon, the oldest available record is chosen.

## Attributes Reference

* `function_arn` - The the ARN of the Lambda function the event source mapping is sending events to. (Note: this is a computed value that differs from `function_name` above.)
* `last_modified` - The date this resource was last modified.
* `last_processing_result` - The result of the last AWS Lambda invocation of your Lambda function.
* `state` - The state of the event source mapping.
* `state_transition_reason` - The reason the event source mapping is in its current state.
* `uuid` - The UUID of the created event source mapping.


[1]: http://docs.aws.amazon.com/lambda/latest/dg/welcome.html
[2]: http://docs.aws.amazon.com/lambda/latest/dg/API_CreateEventSourceMapping.html


## Import

Lambda Event Source Mappings can be imported using the `UUID` (event source mapping identifier), e.g.

```
$ terraform import aws_lambda_event_source_mapping.event_source_mapping 12345kxodurf3443
```
