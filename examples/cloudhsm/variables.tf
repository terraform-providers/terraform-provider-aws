variable "aws_region" {
  description = "AWS region to launch cloudHSM cluster."
  default     = "eu-west-1"
}

variable "subnets" {
  default = ["10.0.1.0/24", "10.0.2.0/24"]
  type    = "list"
}
