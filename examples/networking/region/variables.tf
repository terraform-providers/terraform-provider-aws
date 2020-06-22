variable "region" {
  description = "The name of the AWS region to set up a network within"
}

variable "base_cidr_block" {}

provider "aws" {
  region = "${var.region}"
}

variable "region_numbers" {
  default = {
    us-east-1 = 0
    us-west-1 = 1
    us-west-2 = 2
    eu-west-1 = 3
  }
}
