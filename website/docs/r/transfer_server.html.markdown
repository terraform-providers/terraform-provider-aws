---
layout: "aws"
page_title: "AWS: aws_transfer_server"
sidebar_current: "docs-aws-resource-transfer-server"
description: |-
  Provides a AWS Transfer Server resource.
---

# aws_transfer_server

Provides a AWS Transfer Server resource.


```hcl
resource "aws_iam_role" "foo" {
	name = "tf-test-transfer-server-iam-role"
  
	assume_role_policy = <<EOF
{
	"Version": "2012-10-17",
	"Statement": [
		{
		"Effect": "Allow",
		"Principal": {
			"Service": "transfer.amazonaws.com"
		},
		"Action": "sts:AssumeRole"
		}
	]
}
EOF
}

resource "aws_iam_role_policy" "foo" {
	name = "tf-test-transfer-server-iam-policy-%s"
	role = "${aws_iam_role.foo.id}"
	policy = <<POLICY
{
	"Version": "2012-10-17",
	"Statement": [
		{
		"Sid": "AllowFullAccesstoCloudWatchLogs",
		"Effect": "Allow",
		"Action": [
			"logs:*"
		],
		"Resource": "*"
		}
	]
}
POLICY
}


resource "aws_transfer_server" "foo" {
  identity_provider_type = "SERVICE_MANAGED"
  logging_role = "${aws_iam_role.foo.arn}"

  tags {
	NAME   = "tf-acc-test-transfer-server"
	ENV    = "test"
  }
}
```

## Argument Reference

The following arguments are supported:

* `invocation_role` - (Optional) Amazon Resource Name (ARN) of the IAM role used to authenticate the user account with an `identity_provider_type` of `API_GATEWAY`.
* `url` - (Optional) - URL of the service endpoint used to authenticate users with an `identity_provider_type` of `API_GATEWAY`.
* `identity_provider_type` - (Optional) The mode of authentication enabled for this service. The default value is `SERVICE_MANAGED`, which allows you to store and access SFTP user credentials within the service. `API_GATEWAY` indicates that user authentication requires a call to an API Gateway endpoint URL provided by you to integrate an identity provider of your choice.
* `logging_role` - (Optional) Amazon Resource Name (ARN) of an IAM role that allows the service to write your SFTP users’ activity to your Amazon CloudWatch logs for monitoring and auditing purposes.
* `force_destroy` - (Optional) A boolean that indicates all users associated with the server should be deleted so that the Server can be destroyed without error. The default value is `false`.
* `tags` - (Optional) A mapping of tags to assign to the resource.

## Attributes Reference
In addition to all arguments above, the following attributes are exported:

* `arn` - Amazon Resource Name (ARN) of Transfer Server
* `id`  - The Server ID of the Transfer Server (e.g. `s-12345678`)
* `endpoint` - The endpoint of the Transfer Server (e.g. `s-12345678.server.transfer.REGION.amazonaws.com`)

## Import

Transfer Servers can be imported using the `server id`, e.g.

```
$ terraform import aws_transfer_server.bar s-12345678
```