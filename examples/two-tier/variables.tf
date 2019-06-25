variable "aws_access_key" {
  description = "AWS access key."
  default     = "REPLACE_WITH_YOUR"
}

variable "aws_secret_key" {
  description = "AWS secret key."
  default     = "REPLACE_WITH_YOUR"
}

variable "deployment_name" {
  description = "Desired name of Deployment"
  default     = "test02"
}

variable "public_key_path" {
  description = <<DESCRIPTION
Path to the SSH public key to be used for authentication.
Ensure this keypair is added to your local SSH agent so provisioners can
connect.

Example: ~/.ssh/terraform.pub
DESCRIPTION
  default     = "~/.ssh/REPLACE_WITH_YOUR.pub.key"
}

variable "key_name" {
  description = "Desired name of AWS key pair"
  default     = "REPLACE_WITH_YOUR.pub"
}

variable "private_key_path" {
  description = <<DESCRIPTION
Path to the SSH private key to be used for authentication.
Ensure this keypair is added to your local SSH agent so provisioners can
connect.

Example: ~/.ssh/terraform.priv.key
DESCRIPTION
  default     = "~/.ssh/REPLACE_WITH_YOUR.priv.key"
}

variable "aws_region" {
  description = "AWS region to launch servers."
  default     = "ap-southeast-1"
}

# Ubuntu Bionic 18.04 LTS (x64, hvm:ebs-ssd)
variable "aws_amis" {
  default = {
    us-east-1 = "ami-5cc39523"
    us-west-1 = "ami-d7b355b4"
    ap-northeast-1  = "ami-e875a197"
    sa-east-1 = "ami-ccd48ea0"
    ap-southeast-1  = "ami-31e7e44d"
    ca-central-1  = "ami-c3e567a7"
    ap-south-1  = "ami-ee8ea481"
    eu-central-1  = "ami-3c635cd7"
    eu-west-1 = "ami-d2414e38"
    cn-north-1 = "ami-a001dfcd"
    cn-northwest-1 = "ami-e1bbaf83"
    ap-northeast-2  = "ami-65d86d0b"
    ap-southeast-2  = "ami-23c51c41"
    us-west-2 = "ami-39c28c41"
    us-east-2 = "ami-67142d02"
    eu-west-2 = "ami-ddb950ba"
    ap-northeast-3  = "ami-3aa8a647"
    eu-west-3 = "ami-daf040a7"
  }
}
