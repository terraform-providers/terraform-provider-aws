---
subcategory: "Transfer"
layout: "aws"
page_title: "AWS: aws_transfer_server"
description: |-
  Get information on an AWS Transfer Server resource
---

# Data Source: aws_transfer_server

Use this data source to get the ARN of an AWS Transfer Server for use in other
resources.

## Example Usage

```hcl
data "aws_transfer_server" "example" {
  server_id = "s-1234567"
}
```

## Argument Reference

* `server_id` - (Required) ID for an SFTP server.

## Attributes Reference

* `arn` - Amazon Resource Name (ARN) of Transfer Server
* `endpoint` - The endpoint of the Transfer Server (e.g. `s-12345678.server.transfer.REGION.amazonaws.com`)
* `id`  - The Server ID of the Transfer Server (e.g. `s-12345678`)
* `identity_provider_type` - The mode of authentication enabled for this service. The default value is `SERVICE_MANAGED`, which allows you to store and access SFTP user credentials within the service. `API_GATEWAY` indicates that user authentication requires a call to an API Gateway endpoint URL provided by you to integrate an identity provider of your choice.
* `invocation_role` - Amazon Resource Name (ARN) of the IAM role used to authenticate the user account with an `identity_provider_type` of `API_GATEWAY`.
* `logging_role` - Amazon Resource Name (ARN) of an IAM role that allows the service to write your SFTP users’ activity to your Amazon CloudWatch logs for monitoring and auditing purposes.
* `url` - URL of the service endpoint used to authenticate users with an `identity_provider_type` of `API_GATEWAY`.
