---
layout: "aws"
page_title: "AWS: aws_iam_instance_profile"
sidebar_current: "docs-aws-datasource-iam-instance-profile"
description: |-
  Get information on a Amazon IAM Instance Profile
---

# Data Source: aws_iam_instance_profile

This data source can be used to fetch information about a specific
IAM instance profile. By using this data source, you can reference IAM
instance profile properties without having to hard code ARNs as input.

## Example Usage

```hcl
data "aws_iam_instance_profile" "example" {
  name = "an_example_instance_profile_name"
}
```

## Argument Reference

* `name` - (Required) The friendly IAM instance profile name to match.

## Attributes Reference

* `arn` - The Amazon Resource Name (ARN) specifying the instance profile.

* `create_date` - The string representation of the date the instance profile
   was created.

* `path` - The path to the instance profile.

* `role_arn` - The role arn associated with this instance profile.

* `role_id` - The role id associated with this instance profile.

* `role_name` - The role name associated with this instance profile.
