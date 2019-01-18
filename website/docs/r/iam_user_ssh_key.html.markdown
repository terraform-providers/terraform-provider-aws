---
layout: "aws"
page_title: "AWS: aws_iam_user_ssh_key"
sidebar_current: "docs-aws-resource-iam-user-ssh-key"
description: |-
  Uploads an SSH public key and associates it with the specified IAM user.
---

# aws_iam_user_ssh_key

Uploads an SSH public key and associates it with the specified IAM user.

## Example Usage

```hcl
resource "aws_iam_user" "user" {
  name = "test-user"
  path = "/"
}

resource "aws_iam_user_ssh_key" "user" {
  username   = "${aws_iam_user.user.name}"
  encoding   = "SSH"
  public_key = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQD3F6tyPEFEzV0LX3X8BsXdMsQz1x2cEikKDEY0aIj41qgxMCP/iteneqXSIFZBp5vizPvaoIR3Um9xK7PGoW8giupGn+EPuxIA4cDM4vzOqOkiMPhz5XK0whEjkVzTo4+S0puvDZuwIsdiW9mxhJc7tgBNL0cYlWSYVkz4G/fslNfRPW5mYAM49f4fhtxPb5ok4Q2Lg9dPKVHO/Bgeu5woMc7RY0p1ej6D4CKFE6lymSDJpW0YHX/wqE9+cfEauh7xZcG0q9t2ta6F6fmX0agvpFyZo8aFbXeUBr7osSCJNgvavWbM/06niWrOvYX2xwWdhXmXSrbX8ZbabVohBK41 mytest@mydomain.com"
}
```

## Argument Reference

The following arguments are supported:

* `username` - (Required) The name of the IAM user to associate the SSH public key with.
* `encoding` - (Required) Specifies the public key encoding format to use in the response. To retrieve the public key in ssh-rsa format, use `SSH`. To retrieve the public key in PEM format, use `PEM`.
* `public_key` - (Required) The SSH public key. The public key must be encoded in ssh-rsa format or PEM format.
* `status` - (Optional) The status to assign to the SSH public key. Active means the key can be used for authentication with an AWS CodeCommit repository. Inactive means the key cannot be used. Default is `active`.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `ssh_public_key_id` - The unique identifier for the SSH public key.
* `fingerprint` - The MD5 message digest of the SSH public key.

## Import

SSH public keys can be imported using the `username`, `ssh_public_key_id`, and `encoding` e.g.

```
$ terraform import aws_iam_user_ssh_key.user user:APKAJNCNNJICVN7CFKCA:SSH
```
