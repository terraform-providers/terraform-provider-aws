provider "aws" {
  region = "${var.aws_region}"
}

data "aws_availability_zones" "available" {}

resource "aws_vpc" "cloudhsm2_vpc" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "example-aws_cloudhsm_v2_cluster"
  }
}

resource "aws_subnet" "cloudhsm2_subnets" {
  count                   = 2
  vpc_id                  = "${aws_vpc.cloudhsm2_vpc.id}"
  cidr_block              = "${element(var.subnets, count.index)}"
  map_public_ip_on_launch = false
  availability_zone       = "${element(data.aws_availability_zones.available.names, count.index)}"

  tags = {
    Name = "example-aws_cloudhsm_v2_cluster"
  }
}

resource "aws_cloudhsm_v2_cluster" "cloudhsm_v2_cluster" {
  hsm_type   = "hsm1.medium"
  subnet_ids = ["${aws_subnet.cloudhsm2_subnets.*.id}"]

  tags = {
    Name = "example-aws_cloudhsm_v2_cluster"
  }
}

resource "aws_cloudhsm_v2_hsm" "cloudhsm_v2_hsm" {
  subnet_id  = "${aws_subnet.cloudhsm2_subnets.0.id}"
  cluster_id = "${aws_cloudhsm_v2_cluster.cloudhsm_v2_cluster.cluster_id}"
}

data "aws_cloudhsm_v2_cluster" "cluster" {
  cluster_id = "${aws_cloudhsm_v2_cluster.cloudhsm_v2_cluster.cluster_id}"
  depends_on = ["aws_cloudhsm_v2_hsm.cloudhsm_v2_hsm"]
}
