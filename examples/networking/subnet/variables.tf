variable "vpc_id" {
}

variable "availability_zone" {
}

data "aws_availability_zone" "target" {
  name = var.availability_zone
}

data "aws_vpc" "target" {
  id = var.vpc_id
}

variable "az_numbers" {
  default = {
    a = 0
    b = 1
    c = 2
    d = 3
    e = 4
    f = 5
    g = 6
    h = 7
    i = 8
    j = 9
    k = 10
    l = 11
    m = 12
    n = 13
  }
}
