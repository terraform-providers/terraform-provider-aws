terraform {
  required_version = ">= 0.12"
}

# First account owns the transit gateway and accepts the VPC attachment.
provider "aws" {
  alias = "first"

  region     = var.aws_region
  access_key = var.aws_first_access_key
  secret_key = var.aws_first_secret_key
}

# Second account owns the VPC and creates the VPC attachment.
provider "aws" {
  alias = "second"

  region     = var.aws_region
  access_key = var.aws_second_access_key
  secret_key = var.aws_second_secret_key
}

data "aws_availability_zones" "available" {
  provider = aws.second

  state = "available"
}

data "aws_caller_identity" "second" {
  provider = aws.second
}

resource "aws_ec2_transit_gateway" "example" {
  provider = aws.first

  tags = {
    Name = "terraform-example"
  }
}

resource "aws_ram_resource_share" "example" {
  provider = aws.first

  name = "terraform-example"

  tags = {
    Name = "terraform-example"
  }
}

# Share the transit gateway...
resource "aws_ram_resource_association" "example" {
  provider = aws.first

  resource_arn       = aws_ec2_transit_gateway.example.arn
  resource_share_arn = aws_ram_resource_share.example.id
}

# ...with the second account.
resource "aws_ram_principal_association" "example" {
  provider = aws.first

  principal          = data.aws_caller_identity.second.account_id
  resource_share_arn = aws_ram_resource_share.example.id
}

resource "aws_vpc" "example" {
  provider = aws.second

  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-example"
  }
}

resource "aws_subnet" "example" {
  provider = aws.second

  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "10.0.0.0/24"
  vpc_id            = aws_vpc.example.id

  tags = {
    Name = "terraform-example"
  }
}

# Create the VPC attachment in the second account...
resource "aws_ec2_transit_gateway_vpc_attachment" "example" {
  provider = aws.second

  depends_on = [
    aws_ram_principal_association.example,
    aws_ram_resource_association.example,
  ]

  subnet_ids         = [aws_subnet.example.id]
  transit_gateway_id = aws_ec2_transit_gateway.example.id
  vpc_id             = aws_vpc.example.id

  tags = {
    Name = "terraform-example"
    Side = "Creator"
  }
}

# ...and accept it in the first account.
resource "aws_ec2_transit_gateway_vpc_attachment_accepter" "example" {
  provider = aws.first

  transit_gateway_attachment_id = aws_ec2_transit_gateway_vpc_attachment.example.id

  tags = {
    Name = "terraform-example"
    Side = "Accepter"
  }
}
