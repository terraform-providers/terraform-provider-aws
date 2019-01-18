---
layout: "aws"
page_title: "AWS: aws_codebuild_project"
sidebar_current: "docs-aws-resource-codebuild-project"
description: |-
  Provides a CodeBuild Project resource.
---

# aws_codebuild_project

Provides a CodeBuild Project resource. See also the [`aws_codebuild_webhook` resource](/docs/providers/aws/r/codebuild_webhook.html), which manages the webhook to the source (e.g. the "rebuild every time a code change is pushed" option in the CodeBuild web console).

## Example Usage

```hcl
resource "aws_s3_bucket" "example" {
  bucket = "example"
  acl    = "private"
}

resource "aws_iam_role" "example" {
  name = "example"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "codebuild.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "example" {
  role = "${aws_iam_role.example.name}"

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Resource": [
        "*"
      ],
      "Action": [
        "logs:CreateLogGroup",
        "logs:CreateLogStream",
        "logs:PutLogEvents"
      ]
    },
    {
      "Effect": "Allow",
      "Action": [
        "ec2:CreateNetworkInterface",
        "ec2:DescribeDhcpOptions",
        "ec2:DescribeNetworkInterfaces",
        "ec2:DeleteNetworkInterface",
        "ec2:DescribeSubnets",
        "ec2:DescribeSecurityGroups",
        "ec2:DescribeVpcs"
      ],
      "Resource": "*"
    },
    {
      "Effect": "Allow",
      "Action": [
        "s3:*"
      ],
      "Resource": [
        "${aws_s3_bucket.example.arn}",
        "${aws_s3_bucket.example.arn}/*"
      ]
    }
  ]
}
POLICY
}

resource "aws_codebuild_project" "example" {
  name          = "test-project"
  description   = "test_codebuild_project"
  build_timeout = "5"
  service_role  = "${aws_iam_role.example.arn}"

  artifacts {
    type = "NO_ARTIFACTS"
  }

  cache {
    type     = "S3"
    location = "${aws_s3_bucket.example.bucket}"
  }

  environment {
    compute_type = "BUILD_GENERAL1_SMALL"
    image        = "aws/codebuild/nodejs:6.3.1"
    type         = "LINUX_CONTAINER"

    environment_variable {
      "name"  = "SOME_KEY1"
      "value" = "SOME_VALUE1"
    }

    environment_variable {
      "name"  = "SOME_KEY2"
      "value" = "SOME_VALUE2"
      "type"  = "PARAMETER_STORE"
    }
  }

  source {
    type            = "GITHUB"
    location        = "https://github.com/mitchellh/packer.git"
    git_clone_depth = 1
  }

  vpc_config {
    vpc_id = "vpc-725fca"

    subnets = [
      "subnet-ba35d2e0",
      "subnet-ab129af1",
    ]

    security_group_ids = [
      "sg-f9f27d91",
      "sg-e4f48g23",
    ]
  }

  tags = {
    "Environment" = "Test"
  }
}
```

## Argument Reference

The following arguments are supported:

* `artifacts` - (Required) Information about the project's build output artifacts. Artifact blocks are documented below.
* `environment` - (Required) Information about the project's build environment. Environment blocks are documented below.
* `name` - (Required) The projects name.
* `source` - (Required) Information about the project's input source code. Source blocks are documented below.
* `badge_enabled` - (Optional) Generates a publicly-accessible URL for the projects build badge. Available as `badge_url` attribute when enabled.
* `build_timeout` - (Optional) How long in minutes, from 5 to 480 (8 hours), for AWS CodeBuild to wait until timing out any related build that does not get marked as completed. The default is 60 minutes.
* `cache` - (Optional) Information about the cache storage for the project. Cache blocks are documented below.
* `description` - (Optional) A short description of the project.
* `encryption_key` - (Optional) The AWS Key Management Service (AWS KMS) customer master key (CMK) to be used for encrypting the build project's build output artifacts.
* `service_role` - (Required) The Amazon Resource Name (ARN) of the AWS Identity and Access Management (IAM) role that enables AWS CodeBuild to interact with dependent AWS services on behalf of the AWS account.
* `tags` - (Optional) A mapping of tags to assign to the resource.
* `vpc_config` - (Optional) Configuration for the builds to run inside a VPC. VPC config blocks are documented below.
* `secondary_artifacts` - (Optional) A set of secondary artifacts to be used inside the build. Secondary artifacts blocks are documented below.
* `secondary_sources` - (Optional) A set of secondary sources to be used inside the build. Secondary sources blocks are documented below.

`artifacts` supports the following:

* `type` - (Required) The build output artifact's type. Valid values for this parameter are: `CODEPIPELINE`, `NO_ARTIFACTS` or `S3`.
* `encryption_disabled` - (Optional) If set to true, output artifacts will not be encrypted. If `type` is set to `NO_ARTIFACTS` then this value will be ignored. Defaults to `false`.
* `location` - (Optional) Information about the build output artifact location. If `type` is set to `CODEPIPELINE` or `NO_ARTIFACTS` then this value will be ignored. If `type` is set to `S3`, this is the name of the output bucket. If `path` is not also specified, then `location` can also specify the path of the output artifact in the output bucket.
* `name` - (Optional) The name of the project. If `type` is set to `S3`, this is the name of the output artifact object
* `namespace_type` - (Optional) The namespace to use in storing build artifacts. If `type` is set to `S3`, then valid values for this parameter are: `BUILD_ID` or `NONE`.
* `packaging` - (Optional) The type of build output artifact to create. If `type` is set to `S3`, valid values for this parameter are: `NONE` or `ZIP`
* `path` - (Optional) If `type` is set to `S3`, this is the path to the output artifact

`cache` supports the following:

* `type` - (Optional) The type of storage that will be used for the AWS CodeBuild project cache. Valid values: `NO_CACHE` and `S3`. Defaults to `NO_CACHE`.
* `location` - (Required when cache type is `S3`) The location where the AWS CodeBuild project stores cached resources. For type `S3` the value must be a valid S3 bucket name/prefix.

`environment` supports the following:

* `compute_type` - (Required) Information about the compute resources the build project will use. Available values for this parameter are: `BUILD_GENERAL1_SMALL`, `BUILD_GENERAL1_MEDIUM` or `BUILD_GENERAL1_LARGE`. `BUILD_GENERAL1_SMALL` is only valid if `type` is set to `LINUX_CONTAINER`
* `image` - (Required) The *image identifier* of the Docker image to use for this build project ([list of Docker images provided by AWS CodeBuild.](https://docs.aws.amazon.com/codebuild/latest/userguide/build-env-ref-available.html)). You can read more about the AWS curated environment images in the [documentation](https://docs.aws.amazon.com/codebuild/latest/APIReference/API_ListCuratedEnvironmentImages.html).
* `type` - (Required) The type of build environment to use for related builds. Available values are: `LINUX_CONTAINER` or `WINDOWS_CONTAINER`.
* `environment_variable` - (Optional) A set of environment variables to make available to builds for this build project.
* `privileged_mode` - (Optional) If set to true, enables running the Docker daemon inside a Docker container. Defaults to `false`.
* `certificate` - (Optional) The ARN of the S3 bucket, path prefix and object key that contains the PEM-encoded certificate.

`environment_variable` supports the following:

* `name` - (Required) The environment variable's name or key.
* `value` - (Required) The environment variable's value.
* `type` - (Optional) The type of environment variable. Valid values: `PARAMETER_STORE`, `PLAINTEXT`.

`source` supports the following:

* `type` - (Required) The type of repository that contains the source code to be built. Valid values for this parameter are: `CODECOMMIT`, `CODEPIPELINE`, `GITHUB`, `GITHUB_ENTERPRISE`, `BITBUCKET`, `S3` or `NO_SOURCE`.
* `auth` - (Optional) Information about the authorization settings for AWS CodeBuild to access the source code to be built. Auth blocks are documented below.
* `buildspec` - (Optional) The build spec declaration to use for this build project's related builds. This must be set when `type` is `NO_SOURCE`.
* `git_clone_depth` - (Optional) Truncate git history to this many commits.
* `insecure_ssl` - (Optional) Ignore SSL warnings when connecting to source control.
* `location` - (Optional) The location of the source code from git or s3.
* `report_build_status` - (Optional) Set to `true` to report the status of a build's start and finish to your source provider. This option is only valid when the `type` is `BITBUCKET` or `GITHUB`.

`auth` supports the following:

* `type` - (Required) The authorization type to use. The only valid value is `OAUTH`
* `resource` - (Optional) The resource value that applies to the specified authorization type.

`vpc_config` supports the following:

* `security_group_ids` - (Required) The security group IDs to assign to running builds.
* `subnets` - (Required) The subnet IDs within which to run builds.
* `vpc_id` - (Required) The ID of the VPC within which to run builds.


`secondary_artifacts` supports the following:

* `type` - (Required) The build output artifact's type. Valid values for this parameter are: `CODEPIPELINE`, `NO_ARTIFACTS` or `S3`.
* `artifact_identifier` - (Required) The artifact identifier. Must be the same specified inside AWS CodeBuild buildspec.
* `encryption_disabled` - (Optional) If set to true, output artifacts will not be encrypted. If `type` is set to `NO_ARTIFACTS` then this value will be ignored. Defaults to `false`.
* `location` - (Optional) Information about the build output artifact location. If `type` is set to `CODEPIPELINE` or `NO_ARTIFACTS` then this value will be ignored. If `type` is set to `S3`, this is the name of the output bucket. If `path` is not also specified, then `location` can also specify the path of the output artifact in the output bucket.
* `name` - (Optional) The name of the project. If `type` is set to `S3`, this is the name of the output artifact object
* `namespace_type` - (Optional) The namespace to use in storing build artifacts. If `type` is set to `S3`, then valid values for this parameter are: `BUILD_ID` or `NONE`.
* `packaging` - (Optional) The type of build output artifact to create. If `type` is set to `S3`, valid values for this parameter are: `NONE` or `ZIP`
* `path` - (Optional) If `type` is set to `S3`, this is the path to the output artifact

`secondary_sources` supports the following:

* `type` - (Required) The type of repository that contains the source code to be built. Valid values for this parameter are: `CODECOMMIT`, `CODEPIPELINE`, `GITHUB`, `GITHUB_ENTERPRISE`, `BITBUCKET` or `S3`.
* `source_identifier` - (Required) The source identifier. Source data will be put inside a folder named as this parameter inside AWS CodeBuild source directory
* `auth` - (Optional) Information about the authorization settings for AWS CodeBuild to access the source code to be built. Auth blocks are documented below.
* `buildspec` - (Optional) The build spec declaration to use for this build project's related builds.
* `git_clone_depth` - (Optional) Truncate git history to this many commits.
* `insecure_ssl` - (Optional) Ignore SSL warnings when connecting to source control.
* `location` - (Optional) The location of the source code from git or s3.
* `report_build_status` - (Optional) Set to `true` to report the status of a build's start and finish to your source provider. This option is only valid when your source provider is `GITHUB`, `BITBUCKET`, or `GITHUB_ENTERPRISE`.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The name (if imported via `name`) or ARN (if created via Terraform or imported via ARN) of the CodeBuild project.
* `arn` - The ARN of the CodeBuild project.
* `badge_url` - The URL of the build badge when `badge_enabled` is enabled.

## Import

CodeBuild Project can be imported using the `name`, e.g.

```
$ terraform import aws_codebuild_project.name project-name
```
