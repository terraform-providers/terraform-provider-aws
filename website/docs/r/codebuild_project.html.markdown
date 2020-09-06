---
subcategory: "CodeBuild"
layout: "aws"
page_title: "AWS: aws_codebuild_project"
description: |-
  Provides a CodeBuild Project resource.
---

# Resource: aws_codebuild_project

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
  role = aws_iam_role.example.name

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
        "ec2:CreateNetworkInterfacePermission"
      ],
      "Resource": [
        "arn:aws:ec2:us-east-1:123456789012:network-interface/*"
      ],
      "Condition": {
        "StringEquals": {
          "ec2:Subnet": [
            "${aws_subnet.example1.arn}",
            "${aws_subnet.example2.arn}"
          ],
          "ec2:AuthorizedService": "codebuild.amazonaws.com"
        }
      }
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
  service_role  = aws_iam_role.example.arn

  artifacts {
    type = "NO_ARTIFACTS"
  }

  cache {
    type     = "S3"
    location = aws_s3_bucket.example.bucket
  }

  environment {
    compute_type                = "BUILD_GENERAL1_SMALL"
    image                       = "aws/codebuild/standard:1.0"
    type                        = "LINUX_CONTAINER"
    image_pull_credentials_type = "CODEBUILD"

    environment_variable {
      name  = "SOME_KEY1"
      value = "SOME_VALUE1"
    }

    environment_variable {
      name  = "SOME_KEY2"
      value = "SOME_VALUE2"
      type  = "PARAMETER_STORE"
    }
  }

  logs_config {
    cloudwatch_logs {
      group_name  = "log-group"
      stream_name = "log-stream"
    }

    s3_logs {
      status   = "ENABLED"
      location = "${aws_s3_bucket.example.id}/build-log"
    }
  }

  source {
    type            = "GITHUB"
    location        = "https://github.com/mitchellh/packer.git"
    git_clone_depth = 1

    git_submodules_config {
      fetch_submodules = true
    }
  }

  source_version = "master"

  vpc_config {
    vpc_id = aws_vpc.example.id

    subnets = [
      aws_subnet.example1.id,
      aws_subnet.example2.id,
    ]

    security_group_ids = [
      aws_security_group.example1.id,
      aws_security_group.example2.id,
    ]
  }

  tags = {
    Environment = "Test"
  }
}

resource "aws_codebuild_project" "project-with-cache" {
  name           = "test-project-cache"
  description    = "test_codebuild_project_cache"
  build_timeout  = "5"
  queued_timeout = "5"

  service_role = aws_iam_role.example.arn

  artifacts {
    type = "NO_ARTIFACTS"
  }

  cache {
    type  = "LOCAL"
    modes = ["LOCAL_DOCKER_LAYER_CACHE", "LOCAL_SOURCE_CACHE"]
  }

  environment {
    compute_type                = "BUILD_GENERAL1_SMALL"
    image                       = "aws/codebuild/standard:1.0"
    type                        = "LINUX_CONTAINER"
    image_pull_credentials_type = "CODEBUILD"

    environment_variable {
      name  = "SOME_KEY1"
      value = "SOME_VALUE1"
    }
  }

  source {
    type            = "GITHUB"
    location        = "https://github.com/mitchellh/packer.git"
    git_clone_depth = 1
  }

  tags = {
    Environment = "Test"
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
* `queued_timeout` - (Optional) How long in minutes, from 5 to 480 (8 hours), a build is allowed to be queued before it times out. The default is 8 hours.
* `cache` - (Optional) Information about the cache storage for the project. Cache blocks are documented below.
* `description` - (Optional) A short description of the project.
* `encryption_key` - (Optional) The AWS Key Management Service (AWS KMS) customer master key (CMK) to be used for encrypting the build project's build output artifacts.
* `logs_config` - (Optional) Configuration for the builds to store log data to CloudWatch or S3.
* `service_role` - (Required) The Amazon Resource Name (ARN) of the AWS Identity and Access Management (IAM) role that enables AWS CodeBuild to interact with dependent AWS services on behalf of the AWS account.
* `source_version` - (Optional) A version of the build input to be built for this project. If not specified, the latest version is used.
* `tags` - (Optional) A map of tags to assign to the resource.
* `vpc_config` - (Optional) Configuration for the builds to run inside a VPC. VPC config blocks are documented below.
* `secondary_artifacts` - (Optional) A set of secondary artifacts to be used inside the build. Secondary artifacts blocks are documented below.
* `secondary_sources` - (Optional) A set of secondary sources to be used inside the build. Secondary sources blocks are documented below.

`artifacts` supports the following:

* `type` - (Required) The build output artifact's type. Valid values for this parameter are: `CODEPIPELINE`, `NO_ARTIFACTS` or `S3`.
* `artifact_identifier` - (Optional) The artifact identifier. Must be the same specified inside AWS CodeBuild buildspec.
* `encryption_disabled` - (Optional) If set to true, output artifacts will not be encrypted. If `type` is set to `NO_ARTIFACTS` then this value will be ignored. Defaults to `false`.
* `override_artifact_name` (Optional) If set to true, a name specified in the build spec file overrides the artifact name.
* `location` - (Optional) Information about the build output artifact location. If `type` is set to `CODEPIPELINE` or `NO_ARTIFACTS` then this value will be ignored. If `type` is set to `S3`, this is the name of the output bucket.
* `name` - (Optional) The name of the project. If `type` is set to `S3`, this is the name of the output artifact object
* `namespace_type` - (Optional) The namespace to use in storing build artifacts. If `type` is set to `S3`, then valid values for this parameter are: `BUILD_ID` or `NONE`.
* `packaging` - (Optional) The type of build output artifact to create. If `type` is set to `S3`, valid values for this parameter are: `NONE` or `ZIP`
* `path` - (Optional) If `type` is set to `S3`, this is the path to the output artifact

`cache` supports the following:

* `type` - (Optional) The type of storage that will be used for the AWS CodeBuild project cache. Valid values: `NO_CACHE`, `LOCAL`, and `S3`. Defaults to `NO_CACHE`.
* `location` - (Required when cache type is `S3`) The location where the AWS CodeBuild project stores cached resources. For type `S3` the value must be a valid S3 bucket name/prefix.
* `modes` - (Required when cache type is `LOCAL`) Specifies settings that AWS CodeBuild uses to store and reuse build dependencies. Valid values:  `LOCAL_SOURCE_CACHE`, `LOCAL_DOCKER_LAYER_CACHE`, and `LOCAL_CUSTOM_CACHE`

`environment` supports the following:

* `compute_type` - (Required) Information about the compute resources the build project will use. Available values for this parameter are: `BUILD_GENERAL1_SMALL`, `BUILD_GENERAL1_MEDIUM`, `BUILD_GENERAL1_LARGE` or `BUILD_GENERAL1_2XLARGE`. `BUILD_GENERAL1_SMALL` is only valid if `type` is set to `LINUX_CONTAINER`. When `type` is set to `LINUX_GPU_CONTAINER`, `compute_type` need to be `BUILD_GENERAL1_LARGE`.
* `image` - (Required) The Docker image to use for this build project. Valid values include [Docker images provided by CodeBuild](https://docs.aws.amazon.com/codebuild/latest/userguide/build-env-ref-available.html) (e.g `aws/codebuild/standard:2.0`), [Docker Hub images](https://hub.docker.com/) (e.g. `hashicorp/terraform:latest`), and full Docker repository URIs such as those for ECR (e.g. `137112412989.dkr.ecr.us-west-2.amazonaws.com/amazonlinux:latest`).
* `type` - (Required) The type of build environment to use for related builds. Available values are: `LINUX_CONTAINER`, `LINUX_GPU_CONTAINER`, `WINDOWS_CONTAINER`, `WINDOWS_SERVER_2019_CONTAINER` or `ARM_CONTAINER`.
* `image_pull_credentials_type` - (Optional) The type of credentials AWS CodeBuild uses to pull images in your build. Available values for this parameter are `CODEBUILD` or `SERVICE_ROLE`. When you use a cross-account or private registry image, you must use SERVICE_ROLE credentials. When you use an AWS CodeBuild curated image, you must use CODEBUILD credentials. Default to `CODEBUILD`
* `environment_variable` - (Optional) A set of environment variables to make available to builds for this build project.
* `privileged_mode` - (Optional) If set to true, enables running the Docker daemon inside a Docker container. Defaults to `false`.
* `certificate` - (Optional) The ARN of the S3 bucket, path prefix and object key that contains the PEM-encoded certificate.
* `registry_credential` - (Optional) Information about credentials for access to a private Docker registry. Registry Credential config blocks are documented below.

`environment_variable` supports the following:

* `name` - (Required) The environment variable's name or key.
* `value` - (Required) The environment variable's value.
* `type` - (Optional) The type of environment variable. Valid values: `PARAMETER_STORE`, `PLAINTEXT`, and `SECRETS_MANAGER`.

`logs_config` supports the following:

* `cloudwatch_logs` - (Optional) Configuration for the builds to store logs to CloudWatch
* `s3_logs` - (Optional) Configuration for the builds to store logs to S3.

`cloudwatch_logs` supports the following:

* `status` - (Optional) Current status of logs in CloudWatch Logs for a build project. Valid values: `ENABLED`, `DISABLED`. Defaults to `ENABLED`.
* `group_name` - (Optional) The group name of the logs in CloudWatch Logs.
* `stream_name` - (Optional) The stream name of the logs in CloudWatch Logs.

`s3_logs` supports the following:

* `status` - (Optional) Current status of logs in S3 for a build project. Valid values: `ENABLED`, `DISABLED`. Defaults to `DISABLED`.
* `location` - (Optional) The name of the S3 bucket and the path prefix for S3 logs. Must be set if status is `ENABLED`, otherwise it must be empty.
* `encryption_disabled` - (Optional) Set to `true` if you do not want S3 logs encrypted. Defaults to `false`.

`source` supports the following:

* `type` - (Required) The type of repository that contains the source code to be built. Valid values for this parameter are: `CODECOMMIT`, `CODEPIPELINE`, `GITHUB`, `GITHUB_ENTERPRISE`, `BITBUCKET`, `S3` or `NO_SOURCE`.
* `auth` - (Optional) Information about the authorization settings for AWS CodeBuild to access the source code to be built. Auth blocks are documented below.
* `buildspec` - (Optional) The build spec declaration to use for this build project's related builds. This must be set when `type` is `NO_SOURCE`.
* `git_clone_depth` - (Optional) Truncate git history to this many commits.
* `git_submodules_config` - (Optional) Information about the Git submodules configuration for an AWS CodeBuild build project. Git submodules config blocks are documented below. This option is only valid when the `type` is `CODECOMMIT`, `GITHUB` or `GITHUB_ENTERPRISE`.
* `insecure_ssl` - (Optional) Ignore SSL warnings when connecting to source control.
* `location` - (Optional) The location of the source code from git or s3.
* `report_build_status` - (Optional) Set to `true` to report the status of a build's start and finish to your source provider. This option is only valid when the `type` is `BITBUCKET` or `GITHUB`.


`auth` supports the following:

* `type` - (Required) The authorization type to use. The only valid value is `OAUTH`
* `resource` - (Optional) The resource value that applies to the specified authorization type.

`git_submodules_config` supports the following:

* `fetch_submodules` - (Required) If set to true, fetches Git submodules for the AWS CodeBuild build project.

`vpc_config` supports the following:

* `security_group_ids` - (Required) The security group IDs to assign to running builds.
* `subnets` - (Required) The subnet IDs within which to run builds.
* `vpc_id` - (Required) The ID of the VPC within which to run builds.

`registry_credential` supports the following:

* `credential` - (Required) The Amazon Resource Name (ARN) or name of credentials created using AWS Secrets Manager.
* `credential_provider` - (Required) The service that created the credentials to access a private Docker registry. The valid value, SECRETS_MANAGER, is for AWS Secrets Manager.

`secondary_artifacts` supports the following:

* `type` - (Required) The build output artifact's type. Valid values for this parameter are: `CODEPIPELINE`, `NO_ARTIFACTS` or `S3`.
* `artifact_identifier` - (Required) The artifact identifier. Must be the same specified inside AWS CodeBuild buildspec.
* `encryption_disabled` - (Optional) If set to true, output artifacts will not be encrypted. If `type` is set to `NO_ARTIFACTS` then this value will be ignored. Defaults to `false`.
* `override_artifact_name` (Optional) If set to true, a name specified in the build spec file overrides the artifact name.
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
* `git_submodules_config` - (Optional) Information about the Git submodules configuration for an AWS CodeBuild build project. Git submodules config blocks are documented below. This option is only valid when the `type` is `CODECOMMIT`, `GITHUB` or `GITHUB_ENTERPRISE`.
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
