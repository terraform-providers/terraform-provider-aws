provider "aws" {
  region = "${var.aws_region}"
}

resource "aws_workspaces_directory" "example" {
  directory_id = "${aws_directory_service_directory.example.id}"
  subnet_ids   = ["${aws_subnet.private-a.id}", "${aws_subnet.private-b.id}"]

  # Uncomment this meta-argument if you are creating the IAM resources required by the AWS WorkSpaces service.
  # depends_on = [
  #   "aws_iam_role.workspaces-default"
  # ]
}

data "aws_workspaces_bundle" "value_windows" {
  bundle_id = "wsb-bh8rsxt14" # Value with Windows 10 (English)
}

resource "aws_workspaces_workspace" "example" {
  directory_id = "${aws_workspaces_directory.example.id}"
  bundle_id    = "${data.aws_workspaces_bundle.value_windows.id}"

  # Administrator is always present in a new directory.
  user_name = "Administrator"

  root_volume_encryption_enabled = true
  user_volume_encryption_enabled = true
  volume_encryption_key          = "${aws_kms_key.example.arn}"

  workspace_properties {
    compute_type_name                         = "VALUE"
    user_volume_size_gib                      = 10
    root_volume_size_gib                      = 80
    running_mode                              = "AUTO_STOP"
    running_mode_auto_stop_timeout_in_minutes = 60
  }

  tags = {
    Department = "IT"
  }

  # Uncomment this meta-argument if you are creating the IAM resources required by the AWS WorkSpaces service.
  # depends_on = [
  #   # The role "workspaces_DefaultRole" requires the policy arn:aws:iam::aws:policy/AmazonWorkSpacesServiceAccess
  #   # to create and delete the ENI that the Workspaces service creates for the Workspace
  #   "aws_iam_role_policy_attachment.workspaces-default-service-access",
  # ]
}

resource "aws_workspaces_ip_group" "main" {
  name        = "main"
  description = "Main IP access control group"

  rules {
    source = "10.10.10.10/16"
  }

  rules {
    source      = "11.11.11.11/16"
    description = "Contractors"
  }
}

data "aws_region" "current" {}

data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

locals {
  # Workspace instances are not supported in all AZs in some regions
  # We use joined and split string values here instead of lists for Terraform 0.11 compatibility
  region_workspaces_az_id_strings = {
    "us-east-1" = "${join(",", formatlist("use1-az%d", list("2", "4", "6")))}"
  }

  workspaces_az_id_strings = "${lookup(local.region_workspaces_az_id_strings, data.aws_region.current.name, join(",", data.aws_availability_zones.available.zone_ids))}"
  workspaces_az_ids        = "${split(",", local.workspaces_az_id_strings)}"
}

resource "aws_vpc" "main" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_subnet" "private-a" {
  vpc_id               = "${aws_vpc.main.id}"
  availability_zone_id = "${local.workspaces_az_ids[0]}"
  cidr_block           = "10.0.1.0/24"
}

resource "aws_subnet" "private-b" {
  vpc_id               = "${aws_vpc.main.id}"
  availability_zone_id = "${local.workspaces_az_ids[1]}"
  cidr_block           = "10.0.2.0/24"
}

resource "aws_directory_service_directory" "example" {
  name     = "workspaces.example.com"
  password = "#S1ncerely"
  size     = "Small"
  vpc_settings {
    vpc_id     = "${aws_vpc.main.id}"
    subnet_ids = ["${aws_subnet.private-a.id}", "${aws_subnet.private-b.id}"]
  }
}

resource "aws_kms_key" "example" {
  description = "WorkSpaces example key"
}
