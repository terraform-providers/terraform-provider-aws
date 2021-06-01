---
subcategory: "Transfer"
layout: "aws"
page_title: "AWS: aws_transfer_ssh_key"
description: |-
  Provides a AWS Transfer SSH Public Key resource.
---

# Resource: aws_transfer_ssh_key

Provides a AWS Transfer User SSH Key resource.


```hcl
resource "aws_transfer_ssh_key" "example" {
  server_id = aws_transfer_server.example.id
  user_name = aws_transfer_user.example.user_name
  body      = "... SSH key ..."
}

resource "aws_transfer_server" "example" {
  identity_provider_type = "SERVICE_MANAGED"

  tags = {
    NAME = "tf-acc-test-transfer-server"
  }
}

resource "aws_transfer_user" "example" {
  server_id = aws_transfer_server.example.id
  user_name = "tftestuser"
  role      = aws_iam_role.example.arn

  tags = {
    NAME = "tftestuser"
  }
}

resource "aws_iam_role" "example" {
  name = "tf-test-transfer-user-iam-role"

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

resource "aws_iam_role_policy" "example" {
  name = "tf-test-transfer-user-iam-policy"
  role = aws_iam_role.example.id

  policy = <<POLICY
{
	"Version": "2012-10-17",
	"Statement": [
		{
			"Sid": "AllowFullAccesstoS3",
			"Effect": "Allow",
			"Action": [
				"s3:*"
			],
			"Resource": "*"
		}
	]
}
POLICY
}
```

## Argument Reference

The following arguments are supported:

* `server_id` - (Requirement) The Server ID of the Transfer Server (e.g. `s-12345678`)
* `user_name` - (Requirement) The name of the user account that is assigned to one or more servers.
* `body` - (Requirement) The public key portion of an SSH key pair.

## Import

Transfer SSH Public Key can be imported using the `server_id` and `user_name` and `ssh_public_key_id` separated by `/`.

```
$ terraform import aws_transfer_ssh_key.bar s-12345678/test-username/key-12345
```
