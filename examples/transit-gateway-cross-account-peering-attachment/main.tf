// First accepts the Peering attachment.
provider "aws" {
  alias = "first"

  region     = "${var.aws_first_region}"
  access_key = "${var.aws_first_access_key}"
  secret_key = "${var.aws_first_secret_key}"
}

// Second creates the Peering attachment.
provider "aws" {
  alias = "second"

  region     = "${var.aws_second_region}"
  access_key = "${var.aws_second_access_key}"
  secret_key = "${var.aws_second_secret_key}"
}

data "aws_caller_identity" "first" {
  provider = "aws.first"
}

data "aws_caller_identity" "second" {
  provider = "aws.second"
}

resource "aws_ec2_transit_gateway" "first" {
  provider = "aws.first"

  tags = {
    Name = "terraform-example"
  }
}

resource "aws_ram_resource_share" "example" {
  provider = "aws.first"

  name = "terraform-example"

  tags = {
    Name = "terraform-example"
  }
}

// Share the transit gateway...
resource "aws_ram_resource_association" "example" {
  provider = "aws.first"

  resource_arn       = "${aws_ec2_transit_gateway.first.arn}"
  resource_share_arn = "${aws_ram_resource_share.example.id}"
}

// ...with the second account.
resource "aws_ram_principal_association" "example" {
  provider = "aws.first"

  principal          = "${data.aws_caller_identity.second.account_id}"
  resource_share_arn = "${aws_ram_resource_share.example.id}"
}

resource "aws_ec2_transit_gateway" "second" {
  provider = "aws.second"

  tags = {
    Name = "terraform-example"
  }
}

// Create the Peering attachment in the second account...
resource "aws_ec2_transit_gateway_peering_attachment" "example" {
  provider = "aws.second"

  peer_account_id         = "${data.aws_caller_identity.first.account_id}"
  peer_region             = "${var.aws_first_region}"
  peer_transit_gateway_id = "${aws_ec2_transit_gateway.first.id}"
  transit_gateway_id      = "${aws_ec2_transit_gateway.second.id}"
  tags = {
    Name = "terraform-example"
    Side = "Creator"
  }
  depends_on = ["aws_ram_principal_association.example", "aws_ram_resource_association.example"]

}

// ...it then needs to accepted by the first account.

// ...terraform currently doesnt have resource for Transit Gateway Peering Attachment Acceptance
