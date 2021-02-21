---
subcategory: "IAM"
layout: "aws"
page_title: "AWS: aws_iam_instance_profile"
description: |-
  Provides an IAM instance profile.
---

# Resource: aws_iam_instance_profile

Provides an IAM instance profile.

## Example Usage

```hcl
resource "aws_iam_instance_profile" "test_profile" {
  name = "test_profile"
  role = aws_iam_role.role.name
}

resource "aws_iam_role" "role" {
  name = "test_role"
  path = "/"

  assume_role_policy = <<EOF
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Action": "sts:AssumeRole",
            "Principal": {
               "Service": "ec2.amazonaws.com"
            },
            "Effect": "Allow",
            "Sid": ""
        }
    ]
}
EOF
}
```

## Argument Reference

The following arguments are optional:

* `name` - (Optional, Forces new resource) Name of the instance profile. If omitted, Terraform will assign a random, unique name. Conflicts with `name_prefix`. Can be a string of characters consisting of upper and lowercase alphanumeric characters and these special characters: `_`, `+`, `=`, `,`, `.`, `@`, `-`. Spaces are not allowed.
* `name_prefix` - (Optional, Forces new resource) Creates a unique name beginning with the specified prefix. Conflicts with `name`.
* `path` - (Optional, default "/") Path to the instance profile. For more information about paths, see [IAM Identifiers](https://docs.aws.amazon.com/IAM/latest/UserGuide/Using_Identifiers.html) in the IAM User Guide. Can be a string of characters consisting of either a forward slash (`/`) by itself or a string that must begin and end with forward slashes. Can include any ASCII character from the ! (\u0021) through the DEL character (\u007F), including most punctuation characters, digits, and upper and lowercase letters.
* `role` - (Optional) Name of the role to add to the profile.

## Attributes Reference

In addition to the arguments above, the following attributes are exported:

* `arn` - ARN assigned by AWS to the instance profile.
* `create_date` - Creation timestamp of the instance profile.
* `id` - Instance profile's ID.
* `unique_id` - [Unique ID][1] assigned by AWS.

  [1]: https://docs.aws.amazon.com/IAM/latest/UserGuide/Using_Identifiers.html#GUIDs


## Import

Instance Profiles can be imported using the `name`, e.g.

```
$ terraform import aws_iam_instance_profile.test_profile app-instance-profile-1
```
