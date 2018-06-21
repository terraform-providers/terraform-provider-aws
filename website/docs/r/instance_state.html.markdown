---
layout: "aws"
page_title: "AWS: aws_instance_state"
sidebar_current: "docs-aws-resource-instance-state"
description: |-
  Provides an EC2 instance state resource. This allows changing state of the instance.
---

# aws_instance_state

Provides an EC2 instance resource state. This allows changing state of the instance to be in the stopped state or running state.

## Example Usage

```hcl
# Create a new instance of the latest Ubuntu 14.04 on an
# t2.micro node with an AWS Tag naming it "HelloWorld"
provider "aws" {
  region = "us-west-2"
}

data "aws_ami" "ubuntu" {
  most_recent = true

  filter {
    name   = "name"
    values = ["ubuntu/images/hvm-ssd/ubuntu-trusty-14.04-amd64-server-*"]
  }

  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }

  owners = ["099720109477"] # Canonical
}

resource "aws_instance" "web" {

  ami           = "${data.aws_ami.ubuntu.id}"
  instance_type = "t2.micro"

  tags {
    Name = "HelloWorld"
  }
}


resource "aws_instance_state" "stopped" {
  instance_id   = "${aws_instance.web.id}"
  state         = "stopped"
  depends_on    = ["aws_instance.web"]
}

```

## Argument Reference

The following arguments are supported:

* `instance_id` - (Required) The instance id.
* `state` - (Required) State of instance.
* `force` - (Optional) Force to set state. Default: `false`

### Timeouts

The `timeouts` block allows you to specify [timeouts](https://www.terraform.io/docs/configuration/resources.html#timeouts) for certain actions:

* `create` - (Defaults to 5 mins) Used when launching the instance (until it reaches the initial `running` state)
* `update` - (Defaults to 5 mins) Used when stopping and starting the instance when necessary during update - e.g. when changing instance type
* `delete` - (Defaults to 1 mins) Used when terminating the instance

