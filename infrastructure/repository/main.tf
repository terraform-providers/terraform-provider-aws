terraform {
  backend "atlas" {
    name = "hashicorp-v2/terraform-provider-aws-repository"
  }

  required_providers {
    github = "2.7.0"
  }

  required_version = "~> 0.12.24"
}

provider "github" {
  organization = "terraform-providers"
}
