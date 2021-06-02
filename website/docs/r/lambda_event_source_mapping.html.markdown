---
subcategory: "Lambda"
layout: "aws"
page_title: "AWS: aws_lambda_event_source_mapping"
description: |-
  Provides a Lambda event source mapping. This allows Lambda functions to get events from Kinesis, DynamoDB, SQS and Managed Streaming for Apache Kafka (MSK).
---

# Resource: aws_lambda_event_source_mapping

Provides a Lambda event source mapping. This allows Lambda functions to get events from Kinesis, DynamoDB, SQS and Managed Streaming for Apache Kafka (MSK).

For information about Lambda and how to use it, see [What is AWS Lambda?][1].
For information about event source mappings, see [CreateEventSourceMapping][2] in the API docs.

## Example Usage

### DynamoDB

```terraform
resource "aws_lambda_event_source_mapping" "example" {
  event_source_arn  = aws_dynamodb_table.example.stream_arn
  function_name     = aws_lambda_function.example.arn
  starting_position = "LATEST"
}
```

### Kinesis

```terraform
resource "aws_lambda_event_source_mapping" "example" {
  event_source_arn  = aws_kinesis_stream.example.arn
  function_name     = aws_lambda_function.example.arn
  starting_position = "LATEST"
}
```

### Managed Streaming for Apache Kafka (MSK)

```terraform
resource "aws_lambda_event_source_mapping" "example" {
  event_source_arn  = aws_msk_cluster.example.arn
  function_name     = aws_lambda_function.example.arn
  topics            = ["Example"]
  starting_position = "TRIM_HORIZON"
}
```

### Self Managed Apache Kafka

```terraform
resource "aws_lambda_event_source_mapping" "example" {
  function_name     = aws_lambda_function.example.arn
  topics            = ["Example"]
  starting_position = "TRIM_HORIZON"

  self_managed_event_source {
    endpoints = {
      KAFKA_BOOTSTRAP_SERVERS = "kafka1.example.com:9092,kafka2.example.com:9092"
    }
  }

  source_access_configuration {
    type = "VPC_SUBNET"
    uri  = "subnet:subnet-example1"
  }

  source_access_configuration {
    type = "VPC_SUBNET"
    uri  = "subnet:subnet-example2"
  }

  source_access_configuration {
    type = "VPC_SECURITY_GROUP"
    uri  = "security_group:sg-example"
  }
}
```

### SQS

```terraform
resource "aws_lambda_event_source_mapping" "example" {
  event_source_arn = aws_sqs_queue.sqs_queue_test.arn
  function_name    = aws_lambda_function.example.arn
}
```

## Argument Reference

* `batch_size` - (Optional) The largest number of records that Lambda will retrieve from your event source at the time of invocation. Defaults to `100` for DynamoDB, Kinesis, MQ and MSK, `10` for SQS.
* `bisect_batch_on_function_error`: - (Optional) If the function returns an error, split the batch in two and retry. Only available for stream sources (DynamoDB and Kinesis). Defaults to `false`.
* `destination_config`: - (Optional) An Amazon SQS queue or Amazon SNS topic destination for failed records. Only available for stream sources (DynamoDB and Kinesis). Detailed below.
* `enabled` - (Optional) Determines if the mapping will be enabled on creation. Defaults to `true`.
* `event_source_arn` - (Optional) The event source ARN - this is required for Kinesis stream, DynamoDB stream, SQS queue, MQ broker or MSK cluster.  It is incompatible with a Self Managed Kafka source.
* `function_name` - (Required) The name or the ARN of the Lambda function that will be subscribing to events.
* `function_response_types` - (Optional) A list of current response type enums applied to the event source mapping for [AWS Lambda checkpointing](https://docs.aws.amazon.com/lambda/latest/dg/with-ddb.html#services-ddb-batchfailurereporting). Only available for stream sources (DynamoDB and Kinesis). Valid values: `ReportBatchItemFailures`.
* `maximum_batching_window_in_seconds` - (Optional) The maximum amount of time to gather records before invoking the function, in seconds (between 0 and 300). Records will continue to buffer (or accumulate in the case of an SQS queue event source) until either `maximum_batching_window_in_seconds` expires or `batch_size` has been met. For streaming event sources, defaults to as soon as records are available in the stream. If the batch it reads from the stream/queue only has one record in it, Lambda only sends one record to the function. Only available for stream sources (DynamoDB and Kinesis) and SQS standard queues.
* `maximum_record_age_in_seconds`: - (Optional) The maximum age of a record that Lambda sends to a function for processing. Only available for stream sources (DynamoDB and Kinesis). Must be either -1 (forever, and the default value) or between 60 and 604800 (inclusive).
* `maximum_retry_attempts`: - (Optional) The maximum number of times to retry when the function returns an error. Only available for stream sources (DynamoDB and Kinesis). Minimum and default of -1 (forever), maximum of 10000.
* `parallelization_factor`: - (Optional) The number of batches to process from each shard concurrently. Only available for stream sources (DynamoDB and Kinesis). Minimum and default of 1, maximum of 10.
* `queues` - (Optional) The name of the Amazon MQ broker destination queue to consume. Only available for MQ sources. A single queue name must be specified.
* `self_managed_event_source`: - (Optional) For Self Managed Kafka sources, the location of the self managed cluster. If set, configuration must also include `source_access_configuration`. Detailed below.
* `source_access_configuration`: (Optional) For Self Managed Kafka sources, the access configuration for the source. If set, configuration must also include `self_managed_event_source`. Detailed below.
* `starting_position` - (Optional) The position in the stream where AWS Lambda should start reading. Must be one of `AT_TIMESTAMP` (Kinesis only), `LATEST` or `TRIM_HORIZON` if getting events from Kinesis, DynamoDB or MSK. Must not be provided if getting events from SQS. More information about these positions can be found in the [AWS DynamoDB Streams API Reference](https://docs.aws.amazon.com/amazondynamodb/latest/APIReference/API_streams_GetShardIterator.html) and [AWS Kinesis API Reference](https://docs.aws.amazon.com/kinesis/latest/APIReference/API_GetShardIterator.html#Kinesis-GetShardIterator-request-ShardIteratorType).
* `starting_position_timestamp` - (Optional) A timestamp in [RFC3339 format](https://tools.ietf.org/html/rfc3339#section-5.8) of the data record which to start reading when using `starting_position` set to `AT_TIMESTAMP`. If a record with this exact timestamp does not exist, the next later record is chosen. If the timestamp is older than the current trim horizon, the oldest available record is chosen.
* `topics` - (Optional) The name of the Kafka topics. Only available for MSK sources. A single topic name must be specified.
* `tumbling_window_in_seconds` - (Optional) The duration in seconds of a processing window for [AWS Lambda streaming analytics](https://docs.aws.amazon.com/lambda/latest/dg/with-kinesis.html#services-kinesis-windows). The range is between 1 second up to 900 seconds. Only available for stream sources (DynamoDB and Kinesis).

### destination_config Configuration Block

* `on_failure` - (Optional) The destination configuration for failed invocations. Detailed below.

#### destination_config on_failure Configuration Block

* `destination_arn` - (Required) The Amazon Resource Name (ARN) of the destination resource.

### self_managed_event_source Configuration Block

* `endpoints` - (Required) A map of endpoints for the self managed source.  For Kafka self-managed sources, the key should be `KAFKA_BOOTSTRAP_SERVERS` and the value should be a string with a comma separated list of broker endpoints.

### source_access_configuration Configuration Block

* `type` - (Required) The type of this configuration.  For Self Managed Kafka you will need to supply blocks for type `VPC_SUBNET` and `VPC_SECURITY_GROUP`.
* `uri` - (Required) The URI for this configuration.  For type `VPC_SUBNET` the value should be `subnet:subnet_id` where `subnet_id` is the value you would find in an aws_subnet resource's id attribute.  For type `VPC_SECURITY_GROUP` the value should be `security_group:security_group_id` where `security_group_id` is the value you would find in an aws_security_group resource's id attribute.

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
