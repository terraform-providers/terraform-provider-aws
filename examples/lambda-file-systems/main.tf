terraform {
  required_version = ">= 0.12"
}

provider "aws" {
  region = var.aws_region
}

provider "archive" {}

data "archive_file" "zip" {
  type        = "zip"
  source_file = "hello_lambda.py"
  output_path = "hello_lambda.zip"
}

data "aws_iam_policy_document" "assume_role_policy" {
  statement {
    sid    = ""
    effect = "Allow"

    principals {
      identifiers = ["lambda.amazonaws.com"]
      type        = "Service"
    }

    actions = ["sts:AssumeRole"]
  }
}

data "aws_partition" "current" {}

data "aws_iam_policy" "AWSLambdaVPCAccessExecutionRole" {
  arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AWSLambdaVPCAccessExecutionRole"
}

data "aws_iam_policy" "AmazonElasticFileSystemClientFullAccess" {
  arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AmazonElasticFileSystemClientFullAccess"
}

resource "aws_iam_role" "iam_role_for_lambda" {
  assume_role_policy = data.aws_iam_policy_document.assume_role_policy.json
}

resource "aws_iam_role_policy_attachment" "AWSLambdaVPCAccessExecutionRole-attach" {
  role       = aws_iam_role.iam_role_for_lambda.name
  policy_arn = data.aws_iam_policy.AWSLambdaVPCAccessExecutionRole.arn
}

resource "aws_iam_role_policy_attachment" "AmazonElasticFileSystemClientFullAccess-attach" {
  role       = aws_iam_role.iam_role_for_lambda.name
  policy_arn = data.aws_iam_policy.AmazonElasticFileSystemClientFullAccess.arn
}

# Default VPC
resource "aws_default_vpc" "default" {
}

data "aws_availability_zones" "available" {
  state = "available"
}

# Two default subnets in the Default VPC
resource "aws_default_subnet" "default_az1" {
  availability_zone = data.aws_availability_zones.available.names[0]
}

resource "aws_default_subnet" "default_az2" {
  availability_zone = data.aws_availability_zones.available.names[1]
}

# Default security group
resource "aws_default_security_group" "default" {
  vpc_id = aws_default_vpc.default.id

  ingress {
    protocol  = -1
    self      = true
    from_port = 0
    to_port   = 0
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

# EFS file system
resource "aws_efs_file_system" "efs_for_lambda" {
  tags = {
    Name = "efs_for_lambda"
  }
}

# Two mount targets connect the file system to the subnets
resource "aws_efs_mount_target" "mount_target_az1" {
  file_system_id  = aws_efs_file_system.efs_for_lambda.id
  subnet_id       = aws_default_subnet.default_az1.id
  security_groups = [aws_default_security_group.default.id]
}

resource "aws_efs_mount_target" "mount_target_az2" {
  file_system_id  = aws_efs_file_system.efs_for_lambda.id
  subnet_id       = aws_default_subnet.default_az2.id
  security_groups = [aws_default_security_group.default.id]
}

# EFS access point used by lambda file system
resource "aws_efs_access_point" "access_point_lambda" {
  file_system_id = aws_efs_file_system.efs_for_lambda.id

  root_directory {
    path = "/lambda"
    creation_info {
      owner_gid   = 1000
      owner_uid   = 1000
      permissions = "777"
    }
  }

  posix_user {
    gid = 1000
    uid = 1000
  }
}

resource "aws_lambda_function" "example_lambda" {
  function_name = "hello_lambda"

  filename         = data.archive_file.zip.output_path
  source_code_hash = data.archive_file.zip.output_base64sha256

  role    = aws_iam_role.iam_role_for_lambda.arn
  handler = "hello_lambda.lambda_handler"
  runtime = "python3.7"

  timeout = 60

  environment {
    variables = {
      greeting = "Hello"
    }
  }

  vpc_config {
    subnet_ids         = [aws_default_subnet.default_az1.id, aws_default_subnet.default_az2.id]
    security_group_ids = [aws_default_security_group.default.id]
  }

  file_system_config {
    arn              = aws_efs_access_point.access_point_lambda.arn
    local_mount_path = "/mnt/efs"
  }

  # Explicitly declare dependency on EFS mount target.
  # When creating or updating Lambda functions, mount target must be in 'available' lifecycle state.
  depends_on = [
    aws_efs_mount_target.mount_target_az1,
    aws_efs_mount_target.mount_target_az2,
  ]
}
