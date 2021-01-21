terraform {
  required_version = ">= 0.12"
}

provider "aws" {
  region = var.aws_region
}

resource "aws_elb" "web" {
  name = "terraform-example-elb"

  # The same availability zone as our instances
  availability_zones = aws_instance.web.*.availability_zone

  listener {
    instance_port     = 80
    instance_protocol = "http"
    lb_port           = 80
    lb_protocol       = "http"
  }

  # The instances are registered automatically
  instances = aws_instance.web.*.id
}

data "aws_ami" "ubuntu" {
  most_recent = true

  filter {
    name   = "name"
    values = ["ubuntu/images/hvm-ssd/ubuntu-focal-20.04-amd64-server-*"]
  }

  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }

  owners = ["099720109477"] # Canonical
}

resource "aws_instance" "web" {
  instance_type = "t2.small"
  ami           = data.aws_ami.ubuntu.id

  # This will create 4 instances
  count = 4
}

