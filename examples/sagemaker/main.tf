terraform {
  required_version = ">= 0.12"
}

provider "aws" {
  region = "us-west-2"
}

data "aws_caller_identity" "current" {}

data "aws_region" "current" {}

resource "aws_iam_role" "foo" {
  name               = "terraform-sagemaker-example"
  path               = "/"
  assume_role_policy = data.aws_iam_policy_document.assume_role.json
}

data "aws_iam_policy_document" "assume_role" {
  statement {
    actions = ["sts:AssumeRole"]
    principals {
      type        = "Service"
      identifiers = ["sagemaker.amazonaws.com"]
    }
  }
}

resource "aws_iam_policy" "foo" {
  name        = "terraform-sagemaker-example"
  description = "Allow Sagemaker to create model"
  policy      = data.aws_iam_policy_document.foo.json
}

data "aws_iam_policy_document" "foo" {
  statement {
    effect = "Allow"
    actions = [
      "sagemaker:*"
    ]
    resources = [
      "*"
    ]
  }
  statement {
    effect = "Allow"
    actions = [
      "cloudwatch:PutMetricData",
      "logs:CreateLogStream",
      "logs:PutLogEvents",
      "logs:CreateLogGroup",
      "logs:DescribeLogStreams",
      "ecr:GetAuthorizationToken",
      "ecr:BatchCheckLayerAvailability",
      "ecr:GetDownloadUrlForLayer",
      "ecr:BatchGetImage"
    ]
    resources = [
    "*"]
  }
  statement {
    effect = "Allow"
    actions = [
      "s3:GetObject"
    ]
    resources = [
      "arn:aws:s3:::${aws_s3_bucket.foo.bucket}",
      "arn:aws:s3:::${aws_s3_bucket.foo.bucket}/*"
    ]
  }
}

resource "aws_iam_role_policy_attachment" "foo" {
  role       = aws_iam_role.foo.name
  policy_arn = aws_iam_policy.foo.arn
}

resource "random_integer" "bucket_suffix" {
  min = 1
  max = 99999
}

resource "aws_s3_bucket" "foo" {
  bucket        = "terraform-sagemaker-example-${random_integer.bucket_suffix.result}"
  acl           = "private"
  force_destroy = true
}

resource "aws_s3_bucket_object" "object" {
  bucket = aws_s3_bucket.foo.bucket
  key    = "model.tar.gz"
  source = "model.tar.gz"
}

resource "aws_sagemaker_model" "foo" {
  name               = "terraform-sagemaker-example"
  execution_role_arn = aws_iam_role.foo.arn

  primary_container {
    image          = "${data.aws_caller_identity.current.account_id}.dkr.ecr.${data.aws_region.current.name}.amazonaws.com/foo:latest"
    model_data_url = "https://s3-us-west-2.amazonaws.com/${aws_s3_bucket.foo.bucket}/model.tar.gz"
  }

  tags = {
    foo = "bar"
  }
}

resource "aws_sagemaker_endpoint_configuration" "foo" {
  name = "terraform-sagemaker-example"

  production_variants {
    variant_name           = "variant-1"
    model_name             = aws_sagemaker_model.foo.name
    initial_instance_count = 1
    instance_type          = "ml.t2.medium"
    initial_variant_weight = 1
  }

  tags = {
    foo = "bar"
  }
}

resource "aws_sagemaker_endpoint" "foo" {
  name                 = "terraform-sagemaker-example"
  endpoint_config_name = aws_sagemaker_endpoint_configuration.foo.name

  tags = {
    foo = "bar"
  }
}
