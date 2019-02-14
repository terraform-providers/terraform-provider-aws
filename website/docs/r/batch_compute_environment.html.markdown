---
layout: "aws"
page_title: "AWS: aws_batch_compute_environment"
sidebar_current: "docs-aws-resource-batch-compute-environment"
description: |-
  Creates a AWS Batch compute environment.
---

# aws_batch_compute_environment

Creates a AWS Batch compute environment. Compute environments contain the Amazon ECS container instances that are used to run containerized batch jobs.

For information about AWS Batch, see [What is AWS Batch?][1] .
For information about compute environment, see [Compute Environments][2] .

~> **Note:** To prevent a race condition during environment deletion, make sure to set `depends_on` to the related `aws_iam_role_policy_attachment`;
   otherwise, the policy may be destroyed too soon and the compute environment will then get stuck in the `DELETING` state, see [Troubleshooting AWS Batch][3] .

## Example Usage

```hcl
resource "aws_iam_role" "ecs_instance_role" {
  name = "ecs_instance_role"

  assume_role_policy = <<EOF
{
    "Version": "2012-10-17",
    "Statement": [
	{
	    "Action": "sts:AssumeRole",
	    "Effect": "Allow",
	    "Principal": {
		"Service": "ec2.amazonaws.com"
	    }
	}
    ]
}
EOF
}

resource "aws_iam_role_policy_attachment" "ecs_instance_role" {
  role       = "${aws_iam_role.ecs_instance_role.name}"
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonEC2ContainerServiceforEC2Role"
}

resource "aws_iam_instance_profile" "ecs_instance_role" {
  name = "ecs_instance_role"
  role = "${aws_iam_role.ecs_instance_role.name}"
}

resource "aws_iam_role" "aws_batch_service_role" {
  name = "aws_batch_service_role"

  assume_role_policy = <<EOF
{
    "Version": "2012-10-17",
    "Statement": [
	{
	    "Action": "sts:AssumeRole",
	    "Effect": "Allow",
	    "Principal": {
		"Service": "batch.amazonaws.com"
	    }
	}
    ]
}
EOF
}

resource "aws_iam_role_policy_attachment" "aws_batch_service_role" {
  role       = "${aws_iam_role.aws_batch_service_role.name}"
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSBatchServiceRole"
}

resource "aws_security_group" "sample" {
  name = "aws_batch_compute_environment_security_group"
}

resource "aws_vpc" "sample" {
  cidr_block = "10.1.0.0/16"
}

resource "aws_subnet" "sample" {
  vpc_id     = "${aws_vpc.sample.id}"
  cidr_block = "10.1.1.0/24"
}

resource "aws_batch_compute_environment" "sample" {
  compute_environment_name = "sample"

  compute_resources {
    instance_role = "${aws_iam_instance_profile.ecs_instance_role.arn}"

    instance_type = [
      "c4.large",
    ]

    max_vcpus = 16
    min_vcpus = 0

    security_group_ids = [
      "${aws_security_group.sample.id}",
    ]

    subnets = [
      "${aws_subnet.sample.id}",
    ]

    type = "EC2"
  }

  service_role = "${aws_iam_role.aws_batch_service_role.arn}"
  type         = "MANAGED"
  depends_on   = ["aws_iam_role_policy_attachment.aws_batch_service_role"]
}
```

## Argument Reference

* `compute_environment_name` - (Required) The name for your compute environment. Up to 128 letters (uppercase and lowercase), numbers, and underscores are allowed.
* `compute_resources` - (Optional) Details of the compute resources managed by the compute environment. This parameter is required for managed compute environments. See details below.
* `service_role` - (Required) The full Amazon Resource Name (ARN) of the IAM role that allows AWS Batch to make calls to other AWS services on your behalf.
* `state` - (Optional) The state of the compute environment. If the state is `ENABLED`, then the compute environment accepts jobs from a queue and can scale out automatically based on queues. Valid items are `ENABLED` or `DISABLED`. Defaults to `ENABLED`.
* `type` - (Required) The type of the compute environment. Valid items are `MANAGED` or `UNMANAGED`.

**compute_resources** is a child block with a single argument:

* `bid_percentage` - (Optional) Integer of minimum percentage that a Spot Instance price must be when compared with the On-Demand price for that instance type before instances are launched. For example, if your bid percentage is 20% (`20`), then the Spot price must be below 20% of the current On-Demand price for that EC2 instance. This parameter is required for SPOT compute environments.
* `desired_vcpus` - (Optional) The desired number of EC2 vCPUS in the compute environment.
* `ec2_key_pair` - (Optional) The EC2 key pair that is used for instances launched in the compute environment.
* `image_id` - (Optional) The Amazon Machine Image (AMI) ID used for instances launched in the compute environment.
* `instance_role` - (Required) The Amazon ECS instance role applied to Amazon EC2 instances in a compute environment.
* `instance_type` - (Required) A list of instance types that may be launched.
* `max_vcpus` - (Required) The maximum number of EC2 vCPUs that an environment can reach.
* `min_vcpus` - (Required) The minimum number of EC2 vCPUs that an environment should maintain.
* `security_group_ids` - (Required) A list of EC2 security group that are associated with instances launched in the compute environment.
* `spot_iam_fleet_role` - (Optional) The Amazon Resource Name (ARN) of the Amazon EC2 Spot Fleet IAM role applied to a SPOT compute environment. This parameter is required for SPOT compute environments.
* `subnets` - (Required) A list of VPC subnets into which the compute resources are launched.
* `tags` - (Optional) Key-value pair tags to be applied to resources that are launched in the compute environment.
* `type` - (Required) The type of compute environment. Valid items are `EC2` or `SPOT`.

## Attributes Reference

* `arn` - The Amazon Resource Name (ARN) of the compute environment.
* `ecs_cluster_arn` - The Amazon Resource Name (ARN) of the underlying Amazon ECS cluster used by the compute environment.
* `status` - The current status of the compute environment (for example, CREATING or VALID).
* `status_reason` - A short, human-readable string to provide additional details about the current status of the compute environment.

[1]: http://docs.aws.amazon.com/batch/latest/userguide/what-is-batch.html
[2]: http://docs.aws.amazon.com/batch/latest/userguide/compute_environments.html
[3]: http://docs.aws.amazon.com/batch/latest/userguide/troubleshooting.html
