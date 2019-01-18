---
layout: "aws"
page_title: "AWS: aws_transfer_ssh_key"
sidebar_current: "docs-aws-resource-transfer-ssh-key"
description: |-
  Provides a AWS Transfer SSH Public Key resource.
---

# aws_transfer_ssh_key

Provides a AWS Transfer User SSH Key resource.


```hcl
resource "aws_transfer_server" "foo" {
	identity_provider_type = "SERVICE_MANAGED"

	tags {
		NAME     = "tf-acc-test-transfer-server"
	}
}


resource "aws_iam_role" "foo" {
	name = "tf-test-transfer-user-iam-role-%s"

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
	name = "tf-test-transfer-user-iam-policy-%s"
	role = "${aws_iam_role.foo.id}"
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


resource "aws_transfer_user" "foo" {
	server_id      = "${aws_transfer_server.foo.id}"
	user_name      = "tftestuser"
	role           = "${aws_iam_role.foo.arn}"

	tags {
		NAME = "tftestuser"
	}
}

resource "aws_transfer_ssh_key" "foo" {
	server_id = "${aws_transfer_server.foo.id}"
	user_name = "${aws_transfer_user.foo.user_name}"
	body 	  = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQD3F6tyPEFEzV0LX3X8BsXdMsQz1x2cEikKDEY0aIj41qgxMCP/iteneqXSIFZBp5vizPvaoIR3Um9xK7PGoW8giupGn+EPuxIA4cDM4vzOqOkiMPhz5XK0whEjkVzTo4+S0puvDZuwIsdiW9mxhJc7tgBNL0cYlWSYVkz4G/fslNfRPW5mYAM49f4fhtxPb5ok4Q2Lg9dPKVHO/Bgeu5woMc7RY0p1ej6D4CKFE6lymSDJpW0YHX/wqE9+cfEauh7xZcG0q9t2ta6F6fmX0agvpFyZo8aFbXeUBr7osSCJNgvavWbM/06niWrOvYX2xwWdhXmXSrbX8ZbabVohBK41 example@example.com"
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