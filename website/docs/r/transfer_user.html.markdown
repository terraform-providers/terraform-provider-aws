---
subcategory: "Transfer"
layout: "aws"
page_title: "AWS: aws_transfer_user"
description: |-
  Provides a AWS Transfer User resource.
---

# Resource: aws_transfer_user

Provides a AWS Transfer User resource. Managing SSH keys can be accomplished with the [`aws_transfer_ssh_key` resource](/docs/providers/aws/r/transfer_ssh_key.html).


```hcl
resource "aws_transfer_server" "foo" {
  identity_provider_type = "SERVICE_MANAGED"

  tags = {
    NAME = "tf-acc-test-transfer-server"
  }
}

resource "aws_iam_role" "foo" {
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

resource "aws_iam_role_policy" "foo" {
  name = "tf-test-transfer-user-iam-policy"
  role = aws_iam_role.foo.id

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
  server_id = aws_transfer_server.foo.id
  user_name = "tftestuser"
  role      = aws_iam_role.foo.arn

  home_directory_type = "LOGICAL"
  home_directory_mappings {
    entry  = "/test.pdf"
    target = "/bucket3/test-path/tftestuser.pdf"
  }
}
```

## Argument Reference

The following arguments are supported:

* `server_id` - (Requirement) The Server ID of the Transfer Server (e.g. `s-12345678`)
* `user_name` - (Requirement) The name used for log in to your SFTP server.
* `home_directory` - (Optional) The landing directory (folder) for a user when they log in to the server using their SFTP client.  It should begin with a `/`.  The first item in the path is the name of the home bucket (accessible as `${Transfer:HomeBucket}` in the policy) and the rest is the home directory (accessible as `${Transfer:HomeDirectory}` in the policy). For example, `/example-bucket-1234/username` would set the home bucket to `example-bucket-1234` and the home directory to `username`.
* `home_directory_mappings` - (Optional) Logical directory mappings that specify what S3 paths and keys should be visible to your user and how you want to make them visible. documented below.
* `home_directory_type` - (Optional) The type of landing directory (folder) you mapped for your users' home directory. Valid values are `PATH` and `LOGICAL`.
* `policy` - (Optional) An IAM JSON policy document that scopes down user access to portions of their Amazon S3 bucket. IAM variables you can use inside this policy include `${Transfer:UserName}`, `${Transfer:HomeDirectory}`, and `${Transfer:HomeBucket}`. Since the IAM variable syntax matches Terraform's interpolation syntax, they must be escaped inside Terraform configuration strings (`$${Transfer:UserName}`).  These are evaluated on-the-fly when navigating the bucket.
* `role` - (Requirement) Amazon Resource Name (ARN) of an IAM role that allows the service to controls your user’s access to your Amazon S3 bucket.
* `tags` - (Optional) A map of tags to assign to the resource.

Home Directory Mappings (`home_directory_mappings`) support the following:

* `entry` - (Requirement) Represents an entry and a target.
* `target` - (Requirement) Represents the map target.

## Attributes Reference
In addition to all arguments above, the following attributes are exported:

* `arn` - Amazon Resource Name (ARN) of Transfer User

## Import

Transfer Users can be imported using the `server_id` and `user_name` separated by `/`.

```
$ terraform import aws_transfer_user.bar s-12345678/test-username
```
