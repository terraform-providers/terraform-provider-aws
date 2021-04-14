terraform {
  required_version = ">= 0.12"
}

# First account owns the VGW.
provider "aws" {
  alias = "first"

  region     = var.aws_region
  access_key = var.aws_first_access_key
  secret_key = var.aws_first_secret_key
}

# Second account owns the DXGW.
provider "aws" {
  alias = "second"

  region     = var.aws_region
  access_key = var.aws_second_access_key
  secret_key = var.aws_second_secret_key
}

data "aws_caller_identity" "first" {
  provider = aws.first
}

resource "aws_vpc" "example" {
  provider = aws.first

  cidr_block = "10.255.255.0/28"

  tags = {
    Name = "terraform-example"
  }
}

resource "aws_vpn_gateway" "example" {
  provider = aws.first

  vpc_id = aws_vpc.example.id

  tags = {
    Name = "terraform-example"
  }
}

# Create the association proposal in the first account...
resource "aws_dx_gateway_association_proposal" "example" {
  provider = aws.first

  dx_gateway_id               = aws_dx_gateway.example.id
  dx_gateway_owner_account_id = aws_dx_gateway.example.owner_account_id
  associated_gateway_id       = aws_vpn_gateway.example.id
}

# ...and accept it in the second account, creating the association.
resource "aws_dx_gateway_association" "example" {
  provider = aws.second

  proposal_id                         = aws_dx_gateway_association_proposal.example.id
  dx_gateway_id                       = aws_dx_gateway.example.id
  associated_gateway_owner_account_id = data.aws_caller_identity.first.account_id
}

resource "aws_dx_gateway" "example" {
  provider = aws.second

  name            = "terraform-example"
  amazon_side_asn = "64512"
}
