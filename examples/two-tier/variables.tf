variable "public_key_path" {
  description = <<DESCRIPTION
Path to the SSH public key to be used for authentication.
Ensure this keypair is added to your local SSH agent so provisioners can
connect.

Example: ~/.ssh/terraform.pub
DESCRIPTION
}

variable "key_name" {
  description = "Desired name of AWS key pair"
}

variable "aws_region" {
  description = "AWS region to launch servers."
  default     = "us-west-2"
}

# Ubuntu Bionic 18.04 LTS (x64)
variable "aws_amis" {
  default = {
    eu-west-1 = "ami-07ee42ba0209b6d77"
    us-east-1 = "ami-0bcc094591f354be2"
    us-west-1 = "ami-0dd005d3eb03f66e8"
    us-west-2 = "ami-0a634ae95e11c6f91"
  }
}
