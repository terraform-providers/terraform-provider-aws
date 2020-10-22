terraform {
  required_version = ">= 0.12"
}

# First accepts the Peering attachment.
provider "aws" {
  alias = "first"

  region     = var.aws_first_region
  access_key = var.aws_first_access_key
  secret_key = var.aws_first_secret_key
}

# Second creates the Peering attachment.
provider "aws" {
  alias = "second"

  region     = var.aws_second_region
  access_key = var.aws_second_access_key
  secret_key = var.aws_second_secret_key
}

data "aws_caller_identity" "first" {
  provider = aws.first
}

resource "aws_ec2_transit_gateway" "first" {
  provider = aws.first

  tags = {
    Name = "terraform-example"
  }
}

resource "aws_ec2_transit_gateway" "second" {
  provider = aws.second

  tags = {
    Name = "terraform-example"
  }
}

# Create the Peering attachment in the second account...
resource "aws_ec2_transit_gateway_peering_attachment" "example" {
  provider = aws.second

  peer_account_id         = data.aws_caller_identity.first.account_id
  peer_region             = var.aws_first_region
  peer_transit_gateway_id = aws_ec2_transit_gateway.first.id
  transit_gateway_id      = aws_ec2_transit_gateway.second.id
  tags = {
    Name = "terraform-example"
    Side = "Creator"
  }
}

# ...and accept it in the first account.
resource "aws_ec2_transit_gateway_peering_attachment_accepter" "example" {
  provider = aws.first

  transit_gateway_attachment_id = aws_ec2_transit_gateway_peering_attachment.example.id
  tags = {
    Name = "terraform-example"
    Side = "Acceptor"
  }
}
