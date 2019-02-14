---
layout: "aws"
page_title: "AWS: aws_cloud9_environment_ec2"
sidebar_current: "docs-aws-resource-cloud9-environment-ec2"
description: |-
  Provides a Cloud9 EC2 Development Environment.
---

# aws_cloud9_environment_ec2

Provides a Cloud9 EC2 Development Environment.

## Example Usage

```hcl
resource "aws_cloud9_environment_ec2" "example" {
  instance_type = "t2.micro"
  name          = "example-env"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the environment.
* `instance_type` - (Required) The type of instance to connect to the environment, e.g. `t2.micro`.
* `automatic_stop_time_minutes` - (Optional) The number of minutes until the running instance is shut down after the environment has last been used.
* `description` - (Optional) The description of the environment.
* `owner_arn` - (Optional) The ARN of the environment owner. This can be ARN of any AWS IAM principal. Defaults to the environment's creator.
* `subnet_id` - (Optional) The ID of the subnet in Amazon VPC that AWS Cloud9 will use to communicate with the Amazon EC2 instance.

## Attributes Reference

In addition the the arguments listed above the following attributes are exported:

* `id` - The ID of the environment.
* `arn` - The ARN of the environment.
* `type` - The type of the environment (e.g. `ssh` or `ec2`)
