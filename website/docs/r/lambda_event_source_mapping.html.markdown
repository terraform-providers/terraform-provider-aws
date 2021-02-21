---
subcategory: "Lambda"
layout: "aws"
page_title: "AWS: aws_lambda_event_source_mapping"
description: |-
  Provides a Lambda event source mapping. This allows Lambda functions to get events from Kinesis, DynamoDB and SQS
---

# Resource: aws_lambda_event_source_mapping

Provides a Lambda event source mapping. This allows Lambda functions to get events from Kinesis, DynamoDB and SQS.

For information about Lambda and how to use it, see [What is AWS Lambda?][1].
For information about event source mappings, see [CreateEventSourceMapping][2] in the API docs.

## Example Usage

### DynamoDB

```hcl
resource "aws_lambda_event_source_mapping" "example" {
  event_source_arn  = aws_dynamodb_table.example.stream_arn
  function_name     = aws_lambda_function.example.arn
  starting_position = "LATEST"
}
```

### Kinesis

```hcl
resource "aws_lambda_event_source_mapping" "example" {
  event_source_arn  = aws_kinesis_stream.example.arn
  function_name     = aws_lambda_function.example.arn
  starting_position = "LATEST"
}
```

### Managed Streaming for Kafka (MSK)

```hcl
resource "aws_lambda_event_source_mapping" "example" {
  event_source_arn  = aws_msk_cluster.example.arn
  function_name     = aws_lambda_function.example.arn
  topics            = ["Example"]
  starting_position = "TRIM_HORIZON"
}
```

### SQS

```hcl
resource "aws_lambda_event_source_mapping" "example" {
  event_source_arn = aws_sqs_queue.sqs_queue_test.arn
  function_name    = aws_lambda_function.example.arn
}
```

## Argument Reference

* `batch_size` - (Optional) The largest number of records that Lambda will retrieve from your event source at the time of invocation. Defaults to `100` for DynamoDB, Kinesis and MSK, `10` for SQS.
* `maximum_batching_window_in_seconds` - (Optional) The maximum amount of time to gather records before invoking the function, in seconds (between 0 and 300). Records will continue to buffer (or accumulate in the case of an SQS queue event source) until either `maximum_batching_window_in_seconds` expires or `batch_size` has been met. For streaming event sources, defaults to as soon as records are available in the stream. If the batch it reads from the stream/queue only has one record in it, Lambda only sends one record to the function.
* `event_source_arn` - (Required) The event source ARN - can be a Kinesis stream, DynamoDB stream, SQS queue or MSK cluster.
* `enabled` - (Optional) Determines if the mapping will be enabled on creation. Defaults to `true`.
* `function_name` - (Required) The name or the ARN of the Lambda function that will be subscribing to events.
* `starting_position` - (Optional) The position in the stream where AWS Lambda should start reading. Must be one of `AT_TIMESTAMP` (Kinesis only), `LATEST` or `TRIM_HORIZON` if getting events from Kinesis, DynamoDB or MSK. Must not be provided if getting events from SQS. More information about these positions can be found in the [AWS DynamoDB Streams API Reference](https://docs.aws.amazon.com/amazondynamodb/latest/APIReference/API_streams_GetShardIterator.html) and [AWS Kinesis API Reference](https://docs.aws.amazon.com/kinesis/latest/APIReference/API_GetShardIterator.html#Kinesis-GetShardIterator-request-ShardIteratorType).
* `starting_position_timestamp` - (Optional) A timestamp in [RFC3339 format](https://tools.ietf.org/html/rfc3339#section-5.8) of the data record which to start reading when using `starting_position` set to `AT_TIMESTAMP`. If a record with this exact timestamp does not exist, the next later record is chosen. If the timestamp is older than the current trim horizon, the oldest available record is chosen.
* `parallelization_factor`: - (Optional) The number of batches to process from each shard concurrently. Only available for stream sources (DynamoDB and Kinesis). Minimum and default of 1, maximum of 10.
* `maximum_retry_attempts`: - (Optional) The maximum number of times to retry when the function returns an error. Only available for stream sources (DynamoDB and Kinesis). Minimum of 0, maximum and default of 10000.
* `maximum_record_age_in_seconds`: - (Optional) The maximum age of a record that Lambda sends to a function for processing. Only available for stream sources (DynamoDB and Kinesis). Minimum of 60, maximum and default of 604800.
* `bisect_batch_on_function_error`: - (Optional) If the function returns an error, split the batch in two and retry. Only available for stream sources (DynamoDB and Kinesis). Defaults to `false`.
* `topics` - (Optional) The name of the Kafka topics. Only available for MSK sources. A single topic name must be specified.
* `destination_config`: - (Optional) An Amazon SQS queue or Amazon SNS topic destination for failed records. Only available for stream sources (DynamoDB and Kinesis). Detailed below.

### destination_config Configuration Block

* `on_failure` - (Optional) The destination configuration for failed invocations. Detailed below.

#### destination_config on_failure Configuration Block

* `destination_arn` - (Required) The Amazon Resource Name (ARN) of the destination resource.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `function_arn` - The the ARN of the Lambda function the event source mapping is sending events to. (Note: this is a computed value that differs from `function_name` above.)
* `last_modified` - The date this resource was last modified.
* `last_processing_result` - The result of the last AWS Lambda invocation of your Lambda function.
* `state` - The state of the event source mapping.
* `state_transition_reason` - The reason the event source mapping is in its current state.
* `uuid` - The UUID of the created event source mapping.


[1]: http://docs.aws.amazon.com/lambda/latest/dg/welcome.html
[2]: http://docs.aws.amazon.com/lambda/latest/dg/API_CreateEventSourceMapping.html


## Import

Lambda event source mappings can be imported using the `UUID` (event source mapping identifier), e.g.

```
$ terraform import aws_lambda_event_source_mapping.event_source_mapping 12345kxodurf3443
```

~> **Note:** Terraform will recreate the imported resource as AWS does not expose `startingPosition` information for existing Lambda event source mappings. For information about retrieving event source mappings, see [GetEventSourceMapping][3] in the API docs.

[3]: https://docs.aws.amazon.com/lambda/latest/dg/API_GetEventSourceMapping.html
