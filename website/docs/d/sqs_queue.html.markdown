---
subcategory: "SQS"
layout: "aws"
page_title: "AWS: aws_sqs_queue"
description: |-
  Get information on an Amazon Simple Queue Service (SQS) Queue
---

# Data Source: aws_sqs_queue

Use this data source to get the ARN and URL of queue in AWS Simple Queue Service (SQS).
By using this data source, you can reference SQS queues without having to hardcode
the ARNs as input.

## Example Usage

```hcl
data "aws_sqs_queue" "example" {
  name = "queue"
}
```

## Argument Reference

* `name` - (Required) The name of the queue to match.

## Attributes Reference

* `arn` - The Amazon Resource Name (ARN) of the queue.
* `url` - The URL of the queue.
* `tags` - A map of tags for the resource.
