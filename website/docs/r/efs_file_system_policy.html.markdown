---
subcategory: "EFS"
layout: "aws"
page_title: "AWS: aws_efs_file_system_policy"
description: |-
  Provides an Elastic File System (EFS) File System Policy resource.
---

# Resource: aws_efs_file_system_policy

Provides an Elastic File System (EFS) File System Policy resource.

## Example Usage

```hcl
resource "aws_efs_file_system" "fs" {
  creation_token = "my-product"
}

resource "aws_efs_file_system_policy" "policy" {
  file_system_id = aws_efs_file_system.fs.id

  policy = <<POLICY
{
    "Version": "2012-10-17",
    "Id": "ExamplePolicy01",
    "Statement": [
        {
            "Sid": "ExampleStatement01",
            "Effect": "Allow",
            "Principal": {
                "AWS": "*"
            },
            "Resource": "${aws_efs_file_system.test.arn}",
            "Action": [
                "elasticfilesystem:ClientMount",
                "elasticfilesystem:ClientWrite"
            ],
            "Condition": {
                "Bool": {
                    "aws:SecureTransport": "true"
                }
            }
        }
    ]
}
POLICY
}
```

## Argument Reference

The following arguments are supported:

* `file_system_id` - (Required) The ID of the EFS file system.
* `policy` - (Required) The JSON formatted file system policy for the EFS file system. see [Docs](https://docs.aws.amazon.com/efs/latest/ug/access-control-overview.html#access-control-manage-access-intro-resource-policies) for more info.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID that identifies the file system (e.g. fs-ccfc0d65).

## Import

The EFS file system policies can be imported using the `id`, e.g.

```
$ terraform import aws_efs_file_system_policy.foo fs-6fa144c6
```
